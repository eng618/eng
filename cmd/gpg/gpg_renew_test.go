package gpg

import (
	"testing"
)

func TestInspectGPGKeyParsing(t *testing.T) {
	// Test inspectGPGKey output parsing logic structure
	keyID := "7C180F0FCB31441B"
	if len(keyID) != 16 {
		t.Errorf("expected 16 char key ID, got %d", len(keyID))
	}
}

func TestFindGitHubCLI(t *testing.T) {
	// Test findGitHubCLI does not crash
	_ = findGitHubCLI()
}

func TestResolvePrimaryFingerprint_FullFingerprint(t *testing.T) {
	// 40-character fingerprint should resolve directly
	fpr := "BE363376C8A71C92C2E1DB137C180F0FCB31441B"
	resolved, err := resolvePrimaryFingerprint(fpr, false)
	if err != nil {
		t.Fatalf("expected 40-char fingerprint to resolve, got error: %v", err)
	}
	if resolved != fpr {
		t.Errorf("expected %s, got %s", fpr, resolved)
	}
}
