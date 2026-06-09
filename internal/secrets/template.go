package secrets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/eng618/eng/internal/log"
)

// RenderTemplates scans the given root directory for .tmpl files and renders them
// by removing the .tmpl extension and injecting BWS secrets.
func RenderTemplates(rootPath string, projectID string, verbose bool, useSpinner bool) error {
	if err := ensureBWSAvailable(); err != nil {
		return err
	}
	if err := ensureBWSTokenConfigured(); err != nil {
		return err
	}

	sp := maybeStartSpinner(useSpinner, "Loading secrets from Bitwarden Secrets Manager...")
	defer stopSpinner(sp)

	projectID, err := resolveDotfilesSecretsProjectID(projectID, "")
	if err != nil {
		return err
	}

	secretsByKey, err := listBWSSecretsByKey(projectID)
	if err != nil {
		return err
	}

	updateSpinner(sp, "Scanning for .tmpl files...")

	var tmplFiles []string
	err = filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(d.Name(), ".tmpl") {
			tmplFiles = append(tmplFiles, path)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to scan for templates: %w", err)
	}

	funcMap := template.FuncMap{
		"bws": func(key string) (string, error) {
			secret, exists := secretsByKey[key]
			if !exists {
				// If not found in the initial list, attempt to fetch directly (fallback)
				// This might happen if list only returns a subset or the user created it just now
				val, fetchErr := fetchSingleBWSSecret(key, projectID)
				if fetchErr == nil {
					return val, nil
				}
				return "", fmt.Errorf("secret not found: %s", key)
			}
			return secret.Value, nil
		},
	}

	updateSpinner(sp, "Rendering templates...")

	for _, tmplPath := range tmplFiles {
		log.Verbose(verbose, "Rendering %s", tmplPath)
		
		content, err := os.ReadFile(tmplPath)
		if err != nil {
			return fmt.Errorf("failed to read template %s: %w", tmplPath, err)
		}

		t, err := template.New(filepath.Base(tmplPath)).Funcs(funcMap).Parse(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", tmplPath, err)
		}

		var buf bytes.Buffer
		if err := t.Execute(&buf, nil); err != nil {
			return fmt.Errorf("failed to render template %s: %w", tmplPath, err)
		}

		destPath := strings.TrimSuffix(tmplPath, ".tmpl")
		
		// Preserve file permissions of the template, or use 0600 if it contains secrets
		info, err := os.Stat(tmplPath)
		mode := os.FileMode(0600)
		if err == nil {
			// Copy permissions from template, but ensure it's not world readable
			mode = info.Mode() & 0700
			if mode == 0 {
				mode = 0600
			}
		}

		if err := os.WriteFile(destPath, buf.Bytes(), mode); err != nil {
			return fmt.Errorf("failed to write rendered file %s: %w", destPath, err)
		}
		
		log.Verbose(verbose, "Successfully rendered %s", destPath)
	}

	updateSpinner(sp, "Templates rendered successfully")
	log.Success("Rendered %d templates", len(tmplFiles))
	return nil
}

func fetchSingleBWSSecret(key string, projectID string) (string, error) {
	output, err := invokeBWSWithRetry("secret", "get", key)
	if err != nil {
		return "", err
	}
	var secret bwsSecret
	if err := json.Unmarshal(output, &secret); err != nil {
		return "", fmt.Errorf("failed to parse bws secret %q: %w", key, err)
	}
	return secret.Value, nil
}
