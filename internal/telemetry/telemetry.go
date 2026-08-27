package telemetry

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/cmd/version"
	"github.com/eng618/eng/internal/config"
)

var (
	globalClient *Client
	mu           sync.RWMutex
	once         sync.Once
)

// Init initializes the global telemetry client.
func Init(cfg config.TelemetryConfig, verbose bool) {
	mu.Lock()
	defer mu.Unlock()

	if globalClient != nil {
		globalClient.Drain(200 * time.Millisecond)
	}

	globalClient = NewClient(cfg, verbose)
}

// GetClient returns the active global telemetry client.
func GetClient() *Client {
	mu.RLock()
	defer mu.RUnlock()

	if globalClient == nil {
		// Initialize lazily if not yet initialized
		cfg := config.GetEffectiveTelemetryConfig()
		globalClient = NewClient(cfg, config.IsVerbose())
	}
	return globalClient
}

// SetClient replaces the global telemetry client (primarily used in tests).
func SetClient(c *Client) {
	mu.Lock()
	defer mu.Unlock()
	globalClient = c
}

// TrackCommand sends a command execution event asynchronously.
func TrackCommand(cmd *cobra.Command, args []string, duration time.Duration, execErr error) {
	client := GetClient()
	if client == nil || !client.cfg.Enabled {
		return
	}

	props := BuildCommandProperties(cmd, args, duration, execErr)
	client.Send(Payload{
		Type: "track",
		Payload: EventData{
			Name:       EventCommandExecuted,
			Properties: props,
		},
	})
}

// Track sends a custom named event asynchronously.
func Track(name string, properties map[string]any) {
	client := GetClient()
	if client == nil || !client.cfg.Enabled {
		return
	}

	client.Send(Payload{
		Type: "track",
		Payload: EventData{
			Name:       name,
			Properties: properties,
		},
	})
}

// Identify sends an identify call for the current profile.
func Identify(properties map[string]any) {
	client := GetClient()
	if client == nil || !client.cfg.Enabled {
		return
	}

	if properties == nil {
		properties = make(map[string]any)
	}
	properties["cli_version"] = version.Version
	properties["os"] = runtime.GOOS
	properties["arch"] = runtime.GOARCH

	client.Send(Payload{
		Type: "identify",
		Payload: EventData{
			Properties: properties,
		},
	})
}

// Drain flushes pending telemetry events with the specified timeout.
func Drain(timeout time.Duration) {
	mu.RLock()
	client := globalClient
	mu.RUnlock()

	if client != nil {
		client.Drain(timeout)
	}
}

// TestConnection sends a synchronous test event to verify endpoint and credentials.
func TestConnection(cfg config.TelemetryConfig) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	testClient := NewClient(cfg, false)
	defer testClient.Drain(200 * time.Millisecond)

	payload := Payload{
		Type: "track",
		Payload: EventData{
			Name:      EventTestConnection,
			ProfileID: cfg.ProfileID,
			Properties: map[string]any{
				"test":        true,
				"timestamp":   time.Now().UTC().Format(time.RFC3339),
				"cli_version": version.Version,
				"os":          runtime.GOOS,
			},
		},
	}

	return testClient.SendSync(ctx, payload)
}
