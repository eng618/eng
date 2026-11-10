package telemetry_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/telemetry"
	"github.com/eng618/eng/internal/version"
)

func TestFormatEndpoint(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", config.DefaultAPIURL + "/track"},
		{"https://api.openpanel.dev", "https://api.openpanel.dev/track"},
		{"https://api.openpanel.dev/", "https://api.openpanel.dev/track"},
		{"https://analytics.example.com/api", "https://analytics.example.com/api/track"},
		{"https://analytics.example.com/api/", "https://analytics.example.com/api/track"},
		{"https://analytics.example.com/api/track", "https://analytics.example.com/api/track"},
		{"https://analytics.example.com/track", "https://analytics.example.com/track"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			actual := telemetry.FormatEndpoint(tc.input)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestExtractCommandPath(t *testing.T) {
	rootCmd := &cobra.Command{Use: "eng"}
	gitCmd := &cobra.Command{Use: "git"}
	syncCmd := &cobra.Command{Use: "sync"}

	rootCmd.AddCommand(gitCmd)
	gitCmd.AddCommand(syncCmd)

	full, root, sub := telemetry.ExtractCommandPath(syncCmd)
	assert.Equal(t, "git sync", full)
	assert.Equal(t, "git", root)
	assert.Equal(t, "sync", sub)

	fullRoot, rootSub, subSub := telemetry.ExtractCommandPath(rootCmd)
	assert.Equal(t, "root", fullRoot)
	assert.Equal(t, "root", rootSub)
	assert.Empty(t, subSub)
}

func TestExtractSanitizedFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("token", "", "secret token")
	cmd.Flags().Bool("verbose", false, "verbose output")

	_ = cmd.Flags().Set("token", "super-secret-12345")
	_ = cmd.Flags().Set("verbose", "true")

	flags := telemetry.ExtractSanitizedFlags(cmd)
	assert.Contains(t, flags, "--token")
	assert.Contains(t, flags, "--verbose")
	assert.NotContains(t, flags, "super-secret-12345", "Flag value must never be included in sanitized flags")
}

func TestCategorizeError(t *testing.T) {
	assert.Equal(t, "none", telemetry.CategorizeError(nil))
	assert.Equal(t, "not_found", telemetry.CategorizeError(os.ErrNotExist))
	assert.Equal(t, "permission_denied", telemetry.CategorizeError(os.ErrPermission))
	assert.Equal(t, "usage_error", telemetry.CategorizeError(errors.New("unknown command 'foo'")))
	assert.Equal(t, "cancelled", telemetry.CategorizeError(errors.New("context canceled")))
	assert.Equal(t, "timeout", telemetry.CategorizeError(errors.New("request timeout")))
	assert.Equal(t, "execution_error", telemetry.CategorizeError(errors.New("some unexpected internal failure")))
}

func TestTelemetryClient_SendAndDrain(t *testing.T) {
	var receivedHeaders http.Header
	var receivedPayload telemetry.Payload
	var mu sync.Mutex
	receivedCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		receivedHeaders = r.Header.Clone()
		_ = json.NewDecoder(r.Body).Decode(&receivedPayload)
		receivedCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := config.TelemetryConfig{
		Enabled:      true,
		APIURL:       server.URL,
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		ProfileID:    "test-profile-123",
	}

	client := telemetry.NewClient(cfg, true)
	client.Send(telemetry.Payload{
		Type: "track",
		Payload: telemetry.EventData{
			Name: "test_event",
			Properties: map[string]any{
				"key": "value",
			},
		},
	})

	client.Drain(1 * time.Second)

	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, 1, receivedCount)
	assert.Equal(t, "application/json", receivedHeaders.Get("Content-Type"))
	assert.Equal(t, "test-client-id", receivedHeaders.Get("openpanel-client-id"))
	assert.Equal(t, "test-client-secret", receivedHeaders.Get("openpanel-client-secret"))
	assert.Equal(t, "eng", receivedHeaders.Get("openpanel-sdk-name"))
	assert.Equal(t, version.Version, receivedHeaders.Get("openpanel-sdk-version"))
	ua := receivedHeaders.Get("User-Agent")
	assert.Contains(t, ua, "eng/")
	assert.NotContains(t, ua, "Go-http-client")
	assert.Equal(t, "track", receivedPayload.Type)
	assert.Equal(t, "test_event", receivedPayload.Payload.Name)
	assert.Equal(t, "test-profile-123", receivedPayload.Payload.ProfileID)
	assert.Equal(t, "value", receivedPayload.Payload.Properties["key"])
	assert.Equal(t, "desktop", receivedPayload.Payload.Properties["__device"])
	assert.NotEmpty(t, receivedPayload.Payload.Properties["__os"])
	identify, ok := receivedPayload.Payload.Properties["__identify"].(map[string]any)
	require.True(t, ok, "track events must carry an OpenPanel identify payload")
	assert.Equal(t, "test-profile-123", identify["profileId"])
}

func TestTelemetryClient_Disabled(t *testing.T) {
	requestReceived := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := config.TelemetryConfig{
		Enabled:   false,
		APIURL:    server.URL,
		ClientID:  "test-client-id",
		ProfileID: "test-profile",
	}

	client := telemetry.NewClient(cfg, false)
	client.Send(telemetry.Payload{
		Type: "track",
		Payload: telemetry.EventData{
			Name: "test_event",
		},
	})
	client.Drain(200 * time.Millisecond)

	assert.False(t, requestReceived, "Disabled telemetry client must not send requests")
}

func TestTestConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-client-id", r.Header.Get("openpanel-client-id"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := config.TelemetryConfig{
		Enabled:   true,
		APIURL:    server.URL,
		ClientID:  "test-client-id",
		ProfileID: "test-profile",
	}

	status, err := telemetry.TestConnection(cfg)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
}

func TestSendSyncRetriesGatewayFailure(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`{"message":"bad gateway"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	client := telemetry.NewClient(config.TelemetryConfig{
		Enabled:      true,
		APIURL:       server.URL,
		ClientID:     "id",
		ClientSecret: "secret",
		ProfileID:    "profile",
	}, true)
	t.Cleanup(func() { client.Drain(time.Second) })

	var buf bytes.Buffer
	log.SetWriters(&buf, &buf)
	t.Cleanup(log.ResetWriters)

	status, err := client.SendSync(context.Background(), telemetry.Payload{
		Type:    "track",
		Payload: telemetry.EventData{Name: "test_event", ProfileID: "profile"},
	})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, 2, calls)
	assert.NotContains(t, buf.String(), "502")
}

func TestSendSyncDoesNotRetryRejectedEvent(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"Invalid type"}`))
	}))
	t.Cleanup(server.Close)

	client := telemetry.NewClient(config.TelemetryConfig{
		Enabled:  true,
		APIURL:   server.URL,
		ClientID: "id",
	}, true)
	t.Cleanup(func() { client.Drain(time.Second) })

	var buf bytes.Buffer
	log.SetWriters(&buf, &buf)
	t.Cleanup(log.ResetWriters)

	status, err := client.SendSync(context.Background(), telemetry.Payload{
		Type:    "track",
		Payload: telemetry.EventData{Name: "test_event"},
	})
	require.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, 1, calls)
	assert.Contains(t, buf.String(), "HTTP 400")
	assert.Contains(t, buf.String(), "Invalid type")
}

func TestSendSyncReportsRepeatedGatewayFailure(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("  bad   gateway  "))
	}))
	t.Cleanup(server.Close)

	client := telemetry.NewClient(config.TelemetryConfig{
		Enabled:      true,
		APIURL:       server.URL,
		ClientID:     "id",
		ClientSecret: "super-secret-value",
		ProfileID:    "profile",
	}, true)
	t.Cleanup(func() { client.Drain(time.Second) })

	var buf bytes.Buffer
	log.SetWriters(&buf, &buf)
	t.Cleanup(log.ResetWriters)

	status, err := client.SendSync(context.Background(), telemetry.Payload{
		Type:    "track",
		Payload: telemetry.EventData{Name: "test_event", ProfileID: "profile"},
	})
	require.Error(t, err)
	assert.Equal(t, http.StatusBadGateway, status)
	assert.Equal(t, 2, calls)
	assert.Contains(t, buf.String(), "HTTP 502: bad gateway")
	assert.NotContains(t, buf.String(), "super-secret-value")
}

func TestTrackCommand_Integration(t *testing.T) {
	var receivedPayload telemetry.Payload
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		_ = json.NewDecoder(r.Body).Decode(&receivedPayload)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := config.TelemetryConfig{
		Enabled:      true,
		APIURL:       server.URL,
		ClientID:     "id",
		ClientSecret: "secret",
		ProfileID:    "profile-123",
	}

	telemetry.Init(cfg, false)

	cmd := &cobra.Command{Use: "version"}
	telemetry.TrackCommand(cmd, []string{"--update"}, 150*time.Millisecond, nil)

	telemetry.Drain(1 * time.Second)

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, "cli_command_executed", receivedPayload.Payload.Name)
	assert.Equal(t, "version", receivedPayload.Payload.Properties["command"])
	assert.Equal(t, true, receivedPayload.Payload.Properties["success"])
	assert.Equal(t, float64(0), receivedPayload.Payload.Properties["exit_code"])
	assert.Equal(t, float64(150), receivedPayload.Payload.Properties["duration_ms"])
}
