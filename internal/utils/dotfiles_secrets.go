package utils

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/eng618/eng/internal/utils/log"
)

const dotfilesSecretsRetries = 4

var (
	bwsInvoke = defaultBWSInvoke
	bwsLookUp = exec.LookPath
	sleepFn   = time.Sleep

	rateLimitDelayRE = regexp.MustCompile(`Try again in ([0-9]+)s`)
)

// DotfilesSecretsOptions configures manifest-driven backup and restore operations.
type DotfilesSecretsOptions struct {
	ManifestPath string
	RootPath     string
	ProjectID    string
	Verbose      bool
	UseSpinner   bool
}

// DotfilesSecretsManifest defines the tracked secret mappings for dotfiles env files.
type DotfilesSecretsManifest struct {
	ProjectID string
	Entries   []DotfilesSecretsEntry
}

// DotfilesSecretsEntry maps one env file to a bws prefix and the keys that should be managed.
type DotfilesSecretsEntry struct {
	RelativeFile string
	Prefix       string
	Keys         []string
}

type bwsSecret struct {
	ID    string `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// BackupDotfilesSecrets reads configured env files and saves their managed keys to bws.
func BackupDotfilesSecrets(opts DotfilesSecretsOptions) error {
	if err := ensureBWSAvailable(); err != nil {
		return err
	}
	if err := ensureBWSTokenConfigured(); err != nil {
		return err
	}

	sp := maybeStartSpinner(opts.UseSpinner, "Preparing dotfiles secrets backup...")
	defer stopSpinner(sp)

	manifest, err := LoadDotfilesSecretsManifest(opts.ManifestPath)
	if err != nil {
		return err
	}

	projectID, err := resolveDotfilesSecretsProjectID(opts.ProjectID, manifest.ProjectID)
	if err != nil {
		return err
	}

	rootPath := resolveDotfilesSecretsRoot(opts.RootPath, opts.ManifestPath)
	updateSpinner(sp, "Loading existing secrets from Bitwarden Secrets Manager...")
	secretsByKey, err := listBWSSecretsByKey(projectID)
	if err != nil {
		return err
	}

	for _, entry := range manifest.Entries {
		updateSpinner(sp, fmt.Sprintf("Backing up %s...", entry.RelativeFile))
		targetFile := resolveDotfilesSecretsFile(rootPath, entry.RelativeFile)
		if _, statErr := os.Stat(targetFile); errors.Is(statErr, os.ErrNotExist) {
			log.Warn("Skipping missing file: %s", entry.RelativeFile)
			continue
		} else if statErr != nil {
			return fmt.Errorf("failed to stat %s: %w", targetFile, statErr)
		}

		log.Message("Backing up %s ...", entry.RelativeFile)

		values, err := readEnvValues(targetFile)
		if err != nil {
			return err
		}

		for _, key := range entry.Keys {
			updateSpinner(sp, fmt.Sprintf("Backing up %s/%s...", entry.Prefix, key))
			value, ok := values[key]
			if !ok || strings.TrimSpace(value) == "" {
				return fmt.Errorf("missing key %q in %s", key, entry.RelativeFile)
			}

			secretName := entry.Prefix + "/" + key
			if secret, exists := secretsByKey[secretName]; exists && secret.ID != "" {
				if err := editBWSSecret(secret.ID, projectID, value); err != nil {
					return err
				}
				secret.Value = value
				secretsByKey[secretName] = secret
				log.Message("  Updated: %s", secretName)
				continue
			}

			created, err := createBWSSecret(secretName, value, projectID)
			if err != nil {
				return err
			}
			secretsByKey[secretName] = created
			log.Message("  Created: %s", secretName)
		}
	}

	updateSpinner(sp, "Backup complete")
	log.Success("Backup complete")
	return nil
}

// RestoreDotfilesSecrets recreates managed env files from templates and bws values.
func RestoreDotfilesSecrets(opts DotfilesSecretsOptions) error {
	if err := ensureBWSAvailable(); err != nil {
		return err
	}
	if err := ensureBWSTokenConfigured(); err != nil {
		return err
	}

	sp := maybeStartSpinner(opts.UseSpinner, "Preparing dotfiles secrets restore...")
	defer stopSpinner(sp)

	manifest, err := LoadDotfilesSecretsManifest(opts.ManifestPath)
	if err != nil {
		return err
	}

	projectID, err := resolveDotfilesSecretsProjectID(opts.ProjectID, manifest.ProjectID)
	if err != nil {
		return err
	}

	rootPath := resolveDotfilesSecretsRoot(opts.RootPath, opts.ManifestPath)
	updateSpinner(sp, "Loading secrets from Bitwarden Secrets Manager...")
	secretsByKey, err := listBWSSecretsByKey(projectID)
	if err != nil {
		return err
	}

	for _, entry := range manifest.Entries {
		updateSpinner(sp, fmt.Sprintf("Restoring %s...", entry.RelativeFile))
		targetFile := resolveDotfilesSecretsFile(rootPath, entry.RelativeFile)
		exampleFile := targetFile + ".example"

		if _, statErr := os.Stat(exampleFile); statErr != nil {
			return fmt.Errorf("missing example file for %s: %w", entry.RelativeFile, statErr)
		}

		log.Message("Restoring %s ...", entry.RelativeFile)

		replacements := make(map[string]string, len(entry.Keys))
		for _, key := range entry.Keys {
			updateSpinner(sp, fmt.Sprintf("Restoring %s/%s...", entry.Prefix, key))
			secretName := entry.Prefix + "/" + key
			secret, exists := secretsByKey[secretName]
			if !exists || strings.TrimSpace(secret.Value) == "" {
				return fmt.Errorf("missing bws secret %q in project %s", secretName, projectID)
			}
			replacements[key] = secret.Value
		}

		restored, err := renderEnvTemplate(exampleFile, replacements)
		if err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(targetFile), 0o755); err != nil {
			return fmt.Errorf("failed to create parent directory for %s: %w", targetFile, err)
		}

		if err := os.WriteFile(targetFile, []byte(restored), 0o600); err != nil {
			return fmt.Errorf("failed to write %s: %w", targetFile, err)
		}

		for _, key := range entry.Keys {
			log.Message("  Restored: %s", key)
		}
	}

	updateSpinner(sp, "Restore complete")
	log.Success("Restore complete")
	return nil
}

// DoctorDotfilesSecrets validates manifest-driven secret mappings and templates.
func DoctorDotfilesSecrets(opts DotfilesSecretsOptions) error {
	if err := ensureBWSAvailable(); err != nil {
		return err
	}
	if err := ensureBWSTokenConfigured(); err != nil {
		return err
	}

	sp := maybeStartSpinner(opts.UseSpinner, "Running dotfiles secrets doctor...")
	defer stopSpinner(sp)

	manifest, err := LoadDotfilesSecretsManifest(opts.ManifestPath)
	if err != nil {
		return err
	}

	projectID, err := resolveDotfilesSecretsProjectID(opts.ProjectID, manifest.ProjectID)
	if err != nil {
		return err
	}

	rootPath := resolveDotfilesSecretsRoot(opts.RootPath, opts.ManifestPath)
	updateSpinner(sp, "Loading secrets from Bitwarden Secrets Manager...")
	secretsByKey, err := listBWSSecretsByKey(projectID)
	if err != nil {
		return err
	}

	issues := []string{}
	checked := 0
	for _, entry := range manifest.Entries {
		updateSpinner(sp, fmt.Sprintf("Checking %s...", entry.RelativeFile))
		targetFile := resolveDotfilesSecretsFile(rootPath, entry.RelativeFile)
		exampleFile := targetFile + ".example"
		if _, err := os.Stat(exampleFile); err != nil {
			issues = append(issues, fmt.Sprintf("missing template: %s", exampleFile))
		}

		for _, key := range entry.Keys {
			checked++
			secretName := entry.Prefix + "/" + key
			secret, exists := secretsByKey[secretName]
			if !exists || strings.TrimSpace(secret.Value) == "" {
				issues = append(issues, fmt.Sprintf("missing secret: %s", secretName))
			}
		}
	}

	if len(issues) > 0 {
		updateSpinner(sp, "Doctor found issues")
		for _, issue := range issues {
			log.Error(issue)
		}
		return fmt.Errorf("dotfiles secrets doctor found %d issue(s)", len(issues))
	}

	updateSpinner(sp, "Doctor checks passed")
	log.Success("Dotfiles secrets doctor passed (%d managed keys checked)", checked)
	return nil
}

// LoadDotfilesSecretsManifest parses the tracked manifest used for dotfiles secrets backup and restore.
func LoadDotfilesSecretsManifest(manifestPath string) (*DotfilesSecretsManifest, error) {
	file, err := os.Open(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open manifest %s: %w", manifestPath, err)
	}
	defer func() {
		_ = file.Close()
	}()

	manifest := &DotfilesSecretsManifest{}
	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			const projectPrefix = "# Project UUID:"
			if strings.HasPrefix(line, projectPrefix) {
				manifest.ProjectID = strings.TrimSpace(strings.TrimPrefix(line, projectPrefix))
			}
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid manifest entry at %s:%d", manifestPath, lineNumber)
		}

		keys := strings.Split(parts[2], ",")
		entry := DotfilesSecretsEntry{
			RelativeFile: strings.TrimSpace(parts[0]),
			Prefix:       strings.TrimSpace(parts[1]),
			Keys:         make([]string, 0, len(keys)),
		}
		for _, key := range keys {
			trimmed := strings.TrimSpace(key)
			if trimmed != "" {
				entry.Keys = append(entry.Keys, trimmed)
			}
		}
		if entry.RelativeFile == "" || entry.Prefix == "" || len(entry.Keys) == 0 {
			return nil, fmt.Errorf("invalid manifest entry at %s:%d", manifestPath, lineNumber)
		}
		manifest.Entries = append(manifest.Entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed reading manifest %s: %w", manifestPath, err)
	}

	return manifest, nil
}

func ensureBWSAvailable() error {
	if _, err := bwsLookUp("bws"); err != nil {
		return fmt.Errorf("bws is required but was not found in PATH: %w", err)
	}
	return nil
}

func ensureBWSTokenConfigured() error {
	if strings.TrimSpace(os.Getenv("BWS_ACCESS_TOKEN")) == "" {
		return errors.New("BWS_ACCESS_TOKEN is not set; export your Bitwarden Secrets Manager access token first")
	}
	return nil
}

func resolveDotfilesSecretsProjectID(override, manifestProjectID string) (string, error) {
	if strings.TrimSpace(override) != "" {
		return strings.TrimSpace(override), nil
	}
	if envProjectID := strings.TrimSpace(os.Getenv("BWS_PROJECT_ID")); envProjectID != "" {
		return envProjectID, nil
	}
	if strings.TrimSpace(manifestProjectID) != "" {
		return strings.TrimSpace(manifestProjectID), nil
	}
	return "", errors.New(
		"no Bitwarden Secrets Manager project ID configured; use --project-id, set BWS_PROJECT_ID, or add '# Project UUID:' to the manifest",
	)
}

func resolveDotfilesSecretsRoot(rootPath, manifestPath string) string {
	if strings.TrimSpace(rootPath) != "" {
		return filepath.Clean(rootPath)
	}
	return filepath.Clean(filepath.Join(filepath.Dir(manifestPath), "..", ".."))
}

func resolveDotfilesSecretsFile(rootPath, relativeFile string) string {
	if filepath.IsAbs(relativeFile) {
		return filepath.Clean(relativeFile)
	}
	return filepath.Join(rootPath, filepath.FromSlash(relativeFile))
}

func readEnvValues(path string) (map[string]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read env file %s: %w", path, err)
	}

	values := make(map[string]string)
	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		values[strings.TrimSpace(key)] = value
	}

	return values, nil
}

func renderEnvTemplate(examplePath string, replacements map[string]string) (string, error) {
	content, err := os.ReadFile(examplePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template %s: %w", examplePath, err)
	}

	lines := strings.Split(string(content), "\n")
	seen := make(map[string]bool, len(replacements))
	for idx, line := range lines {
		key, _, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		trimmedKey := strings.TrimSpace(key)
		value, ok := replacements[trimmedKey]
		if !ok {
			continue
		}
		lines[idx] = trimmedKey + "=" + value
		seen[trimmedKey] = true
	}

	for key, value := range replacements {
		if !seen[key] {
			lines = append(lines, key+"="+value)
		}
	}

	return strings.Join(lines, "\n"), nil
}

func listBWSSecretsByKey(projectID string) (map[string]bwsSecret, error) {
	output, err := invokeBWSWithRetry("secret", "list", projectID)
	if err != nil {
		return nil, err
	}

	var secrets []bwsSecret
	if len(strings.TrimSpace(string(output))) > 0 {
		if err := json.Unmarshal(output, &secrets); err != nil {
			return nil, fmt.Errorf("failed to parse bws secret list output: %w", err)
		}
	}

	byKey := make(map[string]bwsSecret, len(secrets))
	for _, secret := range secrets {
		byKey[secret.Key] = secret
	}
	return byKey, nil
}

func createBWSSecret(secretName, value, projectID string) (bwsSecret, error) {
	output, err := invokeBWSWithRetry("secret", "create", secretName, value, projectID)
	if err != nil {
		return bwsSecret{}, err
	}

	var secret bwsSecret
	if err := json.Unmarshal(output, &secret); err != nil {
		return bwsSecret{}, fmt.Errorf("failed to parse bws secret create output: %w", err)
	}
	if secret.Key == "" {
		secret.Key = secretName
	}
	if secret.Value == "" {
		secret.Value = value
	}
	return secret, nil
}

func editBWSSecret(secretID, projectID, value string) error {
	_, err := invokeBWSWithRetry("secret", "edit", secretID, "--value", value, "--project-id", projectID)
	return err
}

func invokeBWSWithRetry(args ...string) ([]byte, error) {
	var lastOutput []byte
	var lastErr error

	for attempt := 0; attempt < dotfilesSecretsRetries; attempt++ {
		output, err := bwsInvoke(args...)
		if err == nil {
			return output, nil
		}

		lastOutput = output
		lastErr = err
		if !isBWSRateLimit(output) || attempt == dotfilesSecretsRetries-1 {
			break
		}

		delay := parseBWSRateLimitDelay(output)
		log.Warn("bws rate limited, retrying in %s", delay)
		sleepFn(delay)
	}

	trimmed := strings.TrimSpace(string(lastOutput))
	if trimmed != "" {
		return nil, fmt.Errorf("bws %s failed: %w: %s", strings.Join(args, " "), lastErr, trimmed)
	}
	return nil, fmt.Errorf("bws %s failed: %w", strings.Join(args, " "), lastErr)
}

func defaultBWSInvoke(args ...string) ([]byte, error) {
	cmd := exec.Command("bws", args...)
	cmd.Env = os.Environ()
	return cmd.CombinedOutput()
}

func isBWSRateLimit(output []byte) bool {
	message := string(output)
	return strings.Contains(message, "429 Too Many Requests") ||
		strings.Contains(message, "Slow down! Too many requests")
}

func parseBWSRateLimitDelay(output []byte) time.Duration {
	matches := rateLimitDelayRE.FindStringSubmatch(string(output))
	if len(matches) == 2 {
		seconds, err := time.ParseDuration(matches[1] + "s")
		if err == nil {
			return seconds
		}
	}
	return time.Second
}

func maybeStartSpinner(enabled bool, msg string) *Spinner {
	if !enabled {
		return nil
	}
	sp := NewSpinner(msg)
	sp.Start()
	return sp
}

func updateSpinner(sp *Spinner, msg string) {
	if sp != nil {
		sp.UpdateMessage(msg)
	}
}

func stopSpinner(sp *Spinner) {
	if sp != nil {
		sp.Stop()
	}
}
