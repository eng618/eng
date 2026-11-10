package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/version"
)

const (
	defaultHTTPTimeout = 2 * time.Second
	// transientRetryDelay is the pause before repeating a gateway failure.
	transientRetryDelay = 200 * time.Millisecond
	// transientRetries is how many extra attempts follow HTTP 502, 503, or 504.
	// Those statuses come from the gateway in front of OpenPanel, not from a
	// rejected event.
	transientRetries = 1
	// ShutdownTimeout is how long the process should wait for queued telemetry
	// to be accepted. It covers two HTTP attempts, the pause between them, and
	// a short margin so a cold TLS handshake is not abandoned when the CLI exits.
	ShutdownTimeout    = 2*defaultHTTPTimeout + transientRetryDelay + 500*time.Millisecond
	eventQueueCapacity = 100
	sdkName            = "eng"
	maxErrorSnippet    = 160
)

// EventData holds the inner payload fields expected by OpenPanel track API.
type EventData struct {
	Name       string         `json:"name,omitempty"`
	ProfileID  string         `json:"profileId,omitempty"`
	Properties map[string]any `json:"properties,omitempty"`
}

// Payload represents an OpenPanel event or identify request.
type Payload struct {
	Type    string    `json:"type"`
	Payload EventData `json:"payload"`
}

// Client is a client for sending events to an OpenPanel instance.
type Client struct {
	cfg        config.TelemetryConfig
	httpClient *http.Client
	eventsChan chan Payload
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	closed     bool
	mu         sync.Mutex
	isVerbose  bool
}

// NewClient initializes a new OpenPanel telemetry client.
func NewClient(cfg config.TelemetryConfig, verbose bool) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
		eventsChan: make(chan Payload, eventQueueCapacity),
		ctx:        ctx,
		cancel:     cancel,
		isVerbose:  verbose,
	}

	// Start background worker
	c.wg.Add(1)
	go c.worker()

	return c
}

// FormatEndpoint constructs the full URL for the OpenPanel track endpoint.
func FormatEndpoint(baseURL string) string {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = config.DefaultAPIURL
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	if strings.HasSuffix(baseURL, "/track") {
		return baseURL
	}
	return baseURL + "/track"
}

// worker processes queued events asynchronously in the background.
func (c *Client) worker() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			// Process any remaining buffered events before exiting
			for len(c.eventsChan) > 0 {
				payload := <-c.eventsChan
				_ = c.sendHTTP(payload)
			}
			return
		case payload, ok := <-c.eventsChan:
			if !ok {
				return
			}
			_ = c.sendHTTP(payload)
		}
	}
}

// Send queues a payload for asynchronous dispatch.
func (c *Client) Send(payload Payload) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed || !c.cfg.Enabled {
		return
	}

	payload = c.prepare(payload)

	select {
	case c.eventsChan <- payload:
	default:
		log.Verbose(c.isVerbose, "Telemetry event queue full, dropping event: %s", payload.Payload.Name)
	}
}

// SendSync sends a payload synchronously, returning any HTTP response status or error.
func (c *Client) SendSync(ctx context.Context, payload Payload) (int, error) {
	return c.sendHTTPRequest(ctx, c.prepare(payload))
}

// prepare fills the anonymous profile id and the OpenPanel hints that make a
// CLI event show up as a client session and a profile. Go's default user agent
// is treated as a server SDK: those events are stored without a session, and
// profiles are only created from session boundaries or an identify payload.
func (c *Client) prepare(payload Payload) Payload {
	if payload.Payload.ProfileID == "" {
		payload.Payload.ProfileID = c.cfg.ProfileID
	}
	if payload.Type != "track" || payload.Payload.ProfileID == "" {
		return payload
	}

	props := cloneProperties(payload.Payload.Properties)
	if _, ok := props["__os"]; !ok {
		props["__os"] = platformOS()
	}
	if _, ok := props["__device"]; !ok {
		props["__device"] = "desktop"
	}
	if _, ok := props["__identify"]; !ok {
		props["__identify"] = map[string]any{
			"profileId": payload.Payload.ProfileID,
			"properties": map[string]any{
				"os":          platformOS(),
				"arch":        runtime.GOARCH,
				"cli_version": version.Version,
				"shell":       DetectShell(),
			},
		}
	}
	payload.Payload.Properties = props
	return payload
}

func cloneProperties(props map[string]any) map[string]any {
	if len(props) == 0 {
		return map[string]any{}
	}
	cloned := make(map[string]any, len(props)+3)
	for key, value := range props {
		cloned[key] = value
	}
	return cloned
}

// sendHTTP executes the HTTP POST request to OpenPanel.
func (c *Client) sendHTTP(payload Payload) error {
	// Each attempt has its own deadline. The parent covers one retry of a
	// gateway response that arrives inside the first deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 2*defaultHTTPTimeout+transientRetryDelay)
	defer cancel()

	_, err := c.sendHTTPRequest(ctx, payload)
	return err
}

func (c *Client) sendHTTPRequest(ctx context.Context, payload Payload) (int, error) {
	endpoint := FormatEndpoint(c.cfg.APIURL)
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Verbose(c.isVerbose, "Telemetry marshal error: %v", err)
		return 0, err
	}

	var status int
	var snippet string
	var reqErr error
	for attempt := 0; attempt <= transientRetries; attempt++ {
		if attempt > 0 {
			if err := wait(ctx, transientRetryDelay); err != nil {
				break
			}
		}
		status, snippet, reqErr = c.postTrack(ctx, endpoint, jsonData)
		if reqErr == nil {
			return status, nil
		}
		if !isTransientStatus(status) {
			break
		}
	}
	if status >= 400 {
		if snippet != "" {
			log.Verbose(c.isVerbose, "Telemetry returned HTTP %d: %s", status, c.redact(snippet))
		} else {
			log.Verbose(c.isVerbose, "Telemetry returned HTTP %d", status)
		}
		return status, reqErr
	}
	log.Verbose(c.isVerbose, "Telemetry dispatch error: %v", reqErr)
	return 0, reqErr
}

func (c *Client) postTrack(ctx context.Context, endpoint string, jsonData []byte) (int, string, error) {
	attemptCtx, cancel := context.WithTimeout(ctx, defaultHTTPTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(attemptCtx, http.MethodPost, endpoint, bytes.NewReader(jsonData))
	if err != nil {
		return 0, "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent())
	req.Header.Set("openpanel-sdk-name", sdkName)
	req.Header.Set("openpanel-sdk-version", version.Version)
	if c.cfg.ClientID != "" {
		req.Header.Set("openpanel-client-id", c.cfg.ClientID)
	}
	if c.cfg.ClientSecret != "" {
		req.Header.Set("openpanel-client-secret", c.cfg.ClientSecret)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode < 400 {
		return resp.StatusCode, "", nil
	}
	return resp.StatusCode, readSnippet(resp.Body), fmt.Errorf("telemetry endpoint returned HTTP %d", resp.StatusCode)
}

func isTransientStatus(code int) bool {
	switch code {
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func wait(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func readSnippet(r io.Reader) string {
	body, err := io.ReadAll(io.LimitReader(r, 512))
	if err != nil || len(bytes.TrimSpace(body)) == 0 {
		return ""
	}
	return oneLine(string(body))
}

func oneLine(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > maxErrorSnippet {
		return s[:maxErrorSnippet]
	}
	return s
}

func (c *Client) redact(s string) string {
	secret := strings.TrimSpace(c.cfg.ClientSecret)
	if len(secret) < 8 {
		return s
	}
	return strings.ReplaceAll(s, secret, "[secret]")
}

// Drain flushes pending events and waits for the worker to finish up to the given timeout.
func (c *Client) Drain(timeout time.Duration) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.closed = true
	c.mu.Unlock()

	// Signal channel close
	close(c.eventsChan)

	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(timeout):
		// Cancel background worker context if timeout exceeded
		c.cancel()
	}
}
