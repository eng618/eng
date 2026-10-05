package gpg

import (
	"net/http"
	"net/http/httptest"
	"os/exec"
	"regexp"
	"testing"
)

func TestFetchAndImportGPGURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("-----BEGIN PGP PUBLIC KEY BLOCK-----\nTest Key\n-----END PGP PUBLIC KEY BLOCK-----\n"))
	}))
	defer server.Close()

	origExec := execCommand
	defer func() { execCommand = origExec }()

	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("echo", "mock")
	}

	err := fetchAndImportGPGURL(server.URL, false)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func BenchmarkValidKeyID_Inline(b *testing.B) {
	keyID := "1234567890ABCDEF"
	for i := 0; i < b.N; i++ {
		validKeyID := regexp.MustCompile("^[0-9A-Fa-f]{16,40}$")
		_ = validKeyID.MatchString(keyID)
	}
}

func BenchmarkValidKeyID_Package(b *testing.B) {
	keyID := "1234567890ABCDEF"
	for i := 0; i < b.N; i++ {
		_ = validKeyIDRegex.MatchString(keyID)
	}
}
