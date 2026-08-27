package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
)

const (
	defaultHTTPTimeout = 2 * time.Second
	eventQueueCapacity = 100
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

	if payload.Payload.ProfileID == "" {
		payload.Payload.ProfileID = c.cfg.ProfileID
	}

	select {
	case c.eventsChan <- payload:
	default:
		log.Verbose(c.isVerbose, "Telemetry event queue full, dropping event: %s", payload.Payload.Name)
	}
}

// SendSync sends a payload synchronously, returning any HTTP response status or error.
func (c *Client) SendSync(ctx context.Context, payload Payload) (int, error) {
	if payload.Payload.ProfileID == "" {
		payload.Payload.ProfileID = c.cfg.ProfileID
	}
	return c.sendHTTPRequest(ctx, payload)
}

// sendHTTP executes the HTTP POST request to OpenPanel.
func (c *Client) sendHTTP(payload Payload) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultHTTPTimeout)
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Verbose(c.isVerbose, "Telemetry request build error: %v", err)
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.cfg.ClientID != "" {
		req.Header.Set("openpanel-client-id", c.cfg.ClientID)
	}
	if c.cfg.ClientSecret != "" {
		req.Header.Set("openpanel-client-secret", c.cfg.ClientSecret)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Verbose(c.isVerbose, "Telemetry dispatch error: %v", err)
		return 0, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		log.Verbose(c.isVerbose, "Telemetry returned non-2xx status: %d", resp.StatusCode)
		return resp.StatusCode, fmt.Errorf("telemetry endpoint returned HTTP %d", resp.StatusCode)
	}

	return resp.StatusCode, nil
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
