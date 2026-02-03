package main

import (
	"strings"
	"testing"
)

func TestGenerateFormula(t *testing.T) {
	config := &Config{
		Version: "1.3.4",
		TagName: "v1.3.4",
		Checksums: Checksums{
			"eng_1.3.4_Darwin_x86_64.tar.gz": "d_amd64_sha",
			"eng_1.3.4_Darwin_arm64.tar.gz":  "d_arm64_sha",
			"eng_1.3.4_Linux_x86_64.tar.gz":  "l_amd64_sha",
			"eng_1.3.4_Linux_arm64.tar.gz":   "l_arm64_sha",
		},
		DarwinAMD64File: "eng_1.3.4_Darwin_x86_64.tar.gz",
		DarwinARM64File: "eng_1.3.4_Darwin_arm64.tar.gz",
		LinuxAMD64File:  "eng_1.3.4_Linux_x86_64.tar.gz",
		LinuxARM64File:  "eng_1.3.4_Linux_arm64.tar.gz",
	}

	formula, err := generateFormula(config)
	if err != nil {
		t.Fatalf("Failed to generate formula: %v", err)
	}

	// Basic checks
	expectedLines := []string{
		"class Eng < Formula",
		"version '1.3.4'",
		"url 'https://github.com/eng618/eng/releases/download/v1.3.4/eng_1.3.4_Darwin_x86_64.tar.gz'",
		"sha256 'd_amd64_sha'",
		"sha256 'd_arm64_sha'",
		"sha256 'l_amd64_sha'",
		"sha256 'l_arm64_sha'",
		"bin.install 'eng'",
	}

	for _, line := range expectedLines {
		if !strings.Contains(formula, line) {
			t.Errorf("Formula missing expected content: %s", line)
		}
	}

	// Regression test for the literal backslash issue
	// The generated Ruby code should NOT contain literal \"
	if strings.Contains(formula, `\"`) {
		t.Errorf("Formula contains literal backslashes which will cause Ruby syntax errors")
	}

	// Verify the File.stat interpolation is correct
	expectedStat := `File.stat("#{bin}/eng").mode.to_s(8)`
	if !strings.Contains(formula, expectedStat) {
		t.Errorf("Formula missing correctly formatted File.stat block. Expected: %s", expectedStat)
	}
}

func TestGenerateFormula_MissingChecksum(t *testing.T) {
	config := &Config{
		Version: "1.3.4",
		TagName: "v1.3.4",
		Checksums: Checksums{
			"eng_1.3.4_Darwin_x86_64.tar.gz": "d_amd64_sha",
		},
		DarwinAMD64File: "eng_1.3.4_Darwin_x86_64.tar.gz",
		DarwinARM64File: "eng_1.3.4_Darwin_arm64.tar.gz", // This is missing from Checksums
	}

	_, err := generateFormula(config)
	if err == nil {
		t.Fatal("Expected error due to missing checksum, but got nil")
	}

	if !strings.Contains(err.Error(), "checksum not found") {
		t.Errorf("Expected 'checksum not found' error, but got: %v", err)
	}
}
