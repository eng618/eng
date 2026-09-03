package logs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/eng618/eng/internal/log"
)

func testEnv(t *testing.T) (logBuf *strings.Builder) {
	t.Helper()
	t.Setenv("ENG_LOG_DIR", t.TempDir())
	var buf strings.Builder
	log.SetWriters(&buf, &buf)
	t.Cleanup(log.ResetWriters)
	return &buf
}

func writeFixture(t *testing.T, dir, name, content string, age time.Duration) {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	ts := time.Now().Add(-age)
	if err := os.Chtimes(p, ts, ts); err != nil {
		t.Fatal(err)
	}
}

func TestListEmpty(t *testing.T) {
	buf := testEnv(t)

	ListCmd.Run(ListCmd, nil)

	if !strings.Contains(buf.String(), "No session logs yet") {
		t.Errorf("expected empty hint, got %q", buf.String())
	}
}

func TestListShowClean(t *testing.T) {
	buf := testEnv(t)
	dir := t.TempDir()
	t.Setenv("ENG_LOG_DIR", dir)
	writeFixture(t, dir, "eng-git-fetch-all-20240101-000000.log", "line1\nline2\nline3\n", 2*time.Hour)
	writeFixture(t, dir, "eng-system-update-20240102-000000.log", "sys1\nsys2\n", time.Hour)

	ListCmd.Run(ListCmd, nil)
	out := buf.String()
	if !strings.Contains(out, "git-fetch-all") || !strings.Contains(out, "system-update") {
		t.Errorf("expected both slugs listed, got %q", out)
	}

	// Show latest (system-update) by default.
	buf.Reset()
	if err := ShowCmd.RunE(ShowCmd, nil); err != nil {
		t.Fatal(err)
	}
	if out := buf.String(); !strings.Contains(out, "sys1") || strings.Contains(out, "line1") {
		t.Errorf("expected latest log only, got %q", out)
	}

	// Show by prefix with tail.
	buf.Reset()
	if err := ShowCmd.Flags().Set("tail", "2"); err != nil {
		t.Fatal(err)
	}
	defer ShowCmd.Flags().Set("tail", "0")
	if err := ShowCmd.RunE(ShowCmd, []string{"eng-git-fetch"}); err != nil {
		t.Fatal(err)
	}
	if out := buf.String(); !strings.Contains(out, "line3") || strings.Contains(out, "line1") {
		t.Errorf("expected last 2 lines only, got %q", out)
	}

	// Unknown name errors.
	if err := ShowCmd.RunE(ShowCmd, []string{"nope"}); err == nil {
		t.Error("expected error for unknown log name")
	}

	// Clean removes everything.
	buf.Reset()
	if err := CleanCmd.RunE(CleanCmd, nil); err != nil {
		t.Fatal(err)
	}
	if out := buf.String(); !strings.Contains(out, "Removed 2 session log(s)") {
		t.Errorf("expected removal count, got %q", out)
	}
	if err := CleanCmd.RunE(CleanCmd, nil); err != nil {
		t.Fatal(err)
	}
	if out := buf.String(); !strings.Contains(out, "No session logs to clean") {
		t.Errorf("expected empty message, got %q", out)
	}
}
