package utils

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eng618/eng/internal/utils/log"
)

func TestLoadDotfilesSecretsManifest(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "server.manifest")
	require.NoError(t, os.WriteFile(manifestPath, []byte(`# file|bws_secret_prefix|secret_keys
# Project UUID: project-123
bin/containers/app.env|app|DB_PASSWORD,API_TOKEN
`), 0o600))

	manifest, err := LoadDotfilesSecretsManifest(manifestPath)
	require.NoError(t, err)
	assert.Equal(t, "project-123", manifest.ProjectID)
	require.Len(t, manifest.Entries, 1)
	assert.Equal(t, "bin/containers/app.env", manifest.Entries[0].RelativeFile)
	assert.Equal(t, "app", manifest.Entries[0].Prefix)
	assert.Equal(t, []string{"DB_PASSWORD", "API_TOKEN"}, manifest.Entries[0].Keys)
}

func TestBackupDotfilesSecretsCreatesAndUpdatesSecrets(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "bin", "secrets", "server.manifest")
	envPath := filepath.Join(tmpDir, "bin", "containers", "app.env")
	require.NoError(t, os.MkdirAll(filepath.Dir(manifestPath), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Dir(envPath), 0o755))
	require.NoError(t, os.WriteFile(manifestPath, []byte(`# Project UUID: project-123
bin/containers/app.env|app|DB_PASSWORD,API_TOKEN
`), 0o600))
	require.NoError(t, os.WriteFile(envPath, []byte("DB_PASSWORD=secret\nAPI_TOKEN=token\n"), 0o600))

	originalInvoke := bwsInvoke
	originalLookup := bwsLookUp
	defer func() {
		bwsInvoke = originalInvoke
		bwsLookUp = originalLookup
	}()

	var calls []string
	bwsLookUp = func(file string) (string, error) { return "/usr/bin/bws", nil }
	bwsInvoke = func(args ...string) ([]byte, error) {
		calls = append(calls, strings.Join(args, " "))
		switch strings.Join(args, " ") {
		case "secret list project-123":
			return []byte(`[{"id":"existing-id","key":"app/DB_PASSWORD","value":"old"}]`), nil
		case "secret edit existing-id --value secret --project-id project-123":
			return []byte(`{"id":"existing-id"}`), nil
		case "secret create app/API_TOKEN token project-123":
			return []byte(`{"id":"new-id","key":"app/API_TOKEN","value":"token"}`), nil
		default:
			return nil, fmt.Errorf("unexpected call: %s", strings.Join(args, " "))
		}
	}

	var buf bytes.Buffer
	log.SetWriters(&buf, &buf)
	defer log.ResetWriters()

	err := BackupDotfilesSecrets(DotfilesSecretsOptions{ManifestPath: manifestPath})
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Updated: app/DB_PASSWORD")
	assert.Contains(t, buf.String(), "Created: app/API_TOKEN")
	assert.Equal(t, []string{
		"secret list project-123",
		"secret edit existing-id --value secret --project-id project-123",
		"secret create app/API_TOKEN token project-123",
	}, calls)
}

func TestRestoreDotfilesSecretsWritesTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "bin", "secrets", "server.manifest")
	examplePath := filepath.Join(tmpDir, "bin", "containers", "app.env.example")
	require.NoError(t, os.MkdirAll(filepath.Dir(manifestPath), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Dir(examplePath), 0o755))
	require.NoError(t, os.WriteFile(manifestPath, []byte(`# Project UUID: project-123
bin/containers/app.env|app|DB_PASSWORD,API_TOKEN
`), 0o600))
	require.NoError(t, os.WriteFile(examplePath, []byte("DB_PASSWORD=__RESTORE__\nAPI_TOKEN=__RESTORE__\n"), 0o600))

	originalInvoke := bwsInvoke
	originalLookup := bwsLookUp
	defer func() {
		bwsInvoke = originalInvoke
		bwsLookUp = originalLookup
	}()

	bwsLookUp = func(file string) (string, error) { return "/usr/bin/bws", nil }
	bwsInvoke = func(args ...string) ([]byte, error) {
		if strings.Join(args, " ") == "secret list project-123" {
			return []byte(`[{"id":"1","key":"app/DB_PASSWORD","value":"restored-db"},{"id":"2","key":"app/API_TOKEN","value":"restored-token"}]`), nil
		}
		return nil, fmt.Errorf("unexpected call: %s", strings.Join(args, " "))
	}

	err := RestoreDotfilesSecrets(DotfilesSecretsOptions{ManifestPath: manifestPath})
	require.NoError(t, err)

	restoredPath := filepath.Join(tmpDir, "bin", "containers", "app.env")
	content, err := os.ReadFile(restoredPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "DB_PASSWORD=restored-db")
	assert.Contains(t, string(content), "API_TOKEN=restored-token")
}

func TestInvokeBWSWithRetryRetriesRateLimit(t *testing.T) {
	originalInvoke := bwsInvoke
	originalSleep := sleepFn
	defer func() {
		bwsInvoke = originalInvoke
		sleepFn = originalSleep
	}()

	attempts := 0
	bwsInvoke = func(args ...string) ([]byte, error) {
		attempts++
		if attempts == 1 {
			return []byte(`Error: 429 Too Many Requests. Slow down! Too many requests. Try again in 1s.`), fmt.Errorf("rate limited")
		}
		return []byte(`[]`), nil
	}

	var slept time.Duration
	sleepFn = func(delay time.Duration) { slept = delay }

	_, err := invokeBWSWithRetry("secret", "list", "project-123")
	require.NoError(t, err)
	assert.Equal(t, 2, attempts)
	assert.Equal(t, time.Second, slept)
}
