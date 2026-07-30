package containers

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverStacks(t *testing.T) {
	tempDir := t.TempDir()
	stacksDir := filepath.Join(tempDir, "stacks")
	mediaDir := filepath.Join(stacksDir, "media")
	arrDir := filepath.Join(stacksDir, "arrsenal")

	_ = os.MkdirAll(mediaDir, 0o755)
	_ = os.MkdirAll(arrDir, 0o755)

	mediaCompose := `
services:
  plex:
    image: plexinc/pms-docker
  jellyfin:
    image: jellyfin/jellyfin
`
	arrCompose := `
services:
  gluetun:
    image: qmcgaw/gluetun
`

	if err := os.WriteFile(filepath.Join(mediaDir, "docker-compose.yml"), []byte(mediaCompose), 0o644); err != nil {
		t.Fatalf("failed to write mock media compose: %v", err)
	}
	if err := os.WriteFile(filepath.Join(arrDir, "docker-compose.yml"), []byte(arrCompose), 0o644); err != nil {
		t.Fatalf("failed to write mock arr compose: %v", err)
	}

	mgr := NewManager(tempDir)
	stacks, err := mgr.DiscoverStacks()
	if err != nil {
		t.Fatalf("unexpected error discovering stacks: %v", err)
	}

	if len(stacks) != 2 {
		t.Errorf("expected 2 stacks, got %d", len(stacks))
	}

	foundMedia := false
	foundArr := false
	for _, s := range stacks {
		if s.Name == "media" {
			foundMedia = true
			if len(s.Services) != 2 {
				t.Errorf("expected 2 services in media, got %d", len(s.Services))
			}
		}
		if s.Name == "arrsenal" {
			foundArr = true
			if len(s.Services) != 1 {
				t.Errorf("expected 1 service in arrsenal, got %d", len(s.Services))
			}
		}
	}

	if !foundMedia || !foundArr {
		t.Errorf("failed to discover media or arrsenal stack")
	}
}

func TestParseDockerPsJSON(t *testing.T) {
	sampleJSON := `{"State":"running","Name":"plex"}
{"State":"running","Name":"jellyfin"}`

	count, status := parseDockerPsJSON([]byte(sampleJSON))
	if count != 2 || status != "Running" {
		t.Errorf("expected count 2 and status Running, got count %d and status %s", count, status)
	}

	partialJSON := `{"State":"running","Name":"plex"}
{"State":"exited","Name":"jellyfin"}`

	count, status = parseDockerPsJSON([]byte(partialJSON))
	if count != 2 || status != "Partial (1/2)" {
		t.Errorf("expected count 2 and status Partial (1/2), got count %d and status %s", count, status)
	}
}

func TestParseContainerDetailsJSON(t *testing.T) {
	sampleJSON := `{"ID":"123","Name":"web-api","Service":"api","State":"running","Health":"healthy","Image":"api:latest","Publishers":[{"TargetPort":80,"PublishedPort":8080,"Protocol":"tcp"}]}
{"ID":"456","Name":"web-db","Service":"db","State":"running","Health":"","Image":"postgres:16","Publishers":[]}`

	details := parseContainerDetailsJSON([]byte(sampleJSON))
	if len(details) != 2 {
		t.Fatalf("expected 2 container details, got %d", len(details))
	}
	if details[0].Name != "web-api" || details[0].Service != "api" || details[0].Health != "healthy" {
		t.Errorf("unexpected first container detail: %+v", details[0])
	}
	if len(details[0].Publishers) != 1 || details[0].Publishers[0].PublishedPort != 8080 {
		t.Errorf("unexpected publisher mapping: %+v", details[0].Publishers)
	}
}
