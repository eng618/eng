package runlog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
)

func testDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("ENG_LOG_DIR", dir)
	return dir
}

func TestStartWritesSessionLog(t *testing.T) {
	testDir(t)
	origDisable := ui.DisableProgress
	ui.DisableProgress = false
	defer func() { ui.DisableProgress = origDisable }()

	path, stop := Start("git sync all!")
	if path == "" {
		t.Fatal("expected a log path, got empty")
	}
	if !strings.HasSuffix(filepath.Base(path), ".log") {
		t.Errorf("expected .log suffix, got %q", path)
	}

	log.Info("hello %s", "world")
	log.Verbose(true, "verbose detail")
	stop()

	if log.FileLogActive() {
		t.Error("expected file tee to be disabled after stop")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, want := range []string{"eng git sync all!", "hello world", "verbose detail"} {
		if !strings.Contains(content, want) {
			t.Errorf("expected log to contain %q, got:\n%s", want, content)
		}
	}
	if strings.Contains(content, "\x1b[") {
		t.Error("expected no ANSI escapes in log file")
	}
}

func TestStartDisabledWhenProgressDisabled(t *testing.T) {
	testDir(t)
	origDisable := ui.DisableProgress
	ui.DisableProgress = true
	defer func() { ui.DisableProgress = origDisable }()

	path, stop := Start("git-sync-all")
	defer stop()
	if path != "" {
		t.Errorf("expected no log path in non-interactive mode, got %q", path)
	}
	entries, err := List()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("expected no log files, got %v", entries)
	}
}

func TestResolve(t *testing.T) {
	dir := testDir(t)

	if _, err := Resolve(""); err == nil {
		t.Error("expected error resolving with no logs")
	}

	base := time.Now().Add(-time.Hour)
	makeLog := func(name string, age time.Duration) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		ts := base.Add(age)
		if err := os.Chtimes(p, ts, ts); err != nil {
			t.Fatal(err)
		}
		return p
	}
	p1 := makeLog("eng-aaa-20240101-000000.log", 0)
	p2 := makeLog("eng-bbb-20240101-000001.log", time.Minute)

	latest, err := Resolve("")
	if err != nil {
		t.Fatal(err)
	}
	if latest.Path != p2 {
		t.Errorf("expected latest %q, got %q", p2, latest.Path)
	}

	entry, err := Resolve("eng-aaa-")
	if err != nil {
		t.Fatal(err)
	}
	if entry.Path != p1 {
		t.Errorf("expected prefix match %q, got %q", p1, entry.Path)
	}

	if _, err := Resolve("eng-"); err == nil {
		t.Error("expected ambiguity error for shared prefix")
	}
	if _, err := Resolve("nope"); err == nil {
		t.Error("expected error for unknown name")
	}
}

func TestParseName(t *testing.T) {
	cmd, ts := ParseName("eng-git-sync-all-20240102-150405.log")
	if cmd != "git-sync-all" {
		t.Errorf("expected command git-sync-all, got %q", cmd)
	}
	if ts.Format("20060102-150405") != "20240102-150405" {
		t.Errorf("expected timestamp 20240102-150405, got %s", ts.Format("20060102-150405"))
	}

	cmd, ts = ParseName("eng-run-20240102-150405-2.log")
	if cmd != "run" || ts.Format("20060102-150405") != "20240102-150405" {
		t.Errorf("expected collision suffix handled, got %q %v", cmd, ts)
	}

	cmd, _ = ParseName("eng-run-2-20240102-150405.log")
	if cmd != "run-2" {
		t.Errorf("expected slug run-2 preserved, got %q", cmd)
	}

	if cmd, _ := ParseName("other.log"); cmd != "other.log" {
		t.Errorf("expected raw name passthrough, got %q", cmd)
	}
}

func TestStartCollisionSuffix(t *testing.T) {
	testDir(t)
	origDisable := ui.DisableProgress
	ui.DisableProgress = false
	defer func() { ui.DisableProgress = origDisable }()

	p1, stop1 := Start("same-second")
	defer stop1()
	p2, stop2 := Start("same-second")
	defer stop2()

	if p1 == "" || p2 == "" {
		t.Fatalf("expected two log paths, got %q and %q", p1, p2)
	}
	if p1 == p2 {
		t.Error("expected distinct paths for same-second sessions")
	}
}

func TestPruneKeepsRecent(t *testing.T) {
	dir := testDir(t)

	base := time.Now().Add(-time.Hour)
	for i := 0; i < KeepRuns+5; i++ {
		name := filepath.Join(dir, fileName("run", base.Add(time.Duration(i)*time.Minute)))
		if err := os.WriteFile(name, []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		ts := base.Add(time.Duration(i) * time.Minute)
		if err := os.Chtimes(name, ts, ts); err != nil {
			t.Fatal(err)
		}
	}

	origDisable := ui.DisableProgress
	ui.DisableProgress = false
	defer func() { ui.DisableProgress = origDisable }()
	if _, stop := Start("newest"); stop != nil {
		stop()
	}

	entries, err := List()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != KeepRuns {
		t.Errorf("expected %d retained logs, got %d", KeepRuns, len(entries))
	}
	present := map[string]bool{}
	for _, e := range entries {
		present[e.Name] = true
	}
	oldest := "eng-run-" + base.Format("20060102-150405") + ".log"
	if present[oldest] {
		t.Errorf("expected oldest log %s to be pruned", oldest)
	}
	newestFixture := "eng-run-" + base.Add(time.Duration(KeepRuns+4)*time.Minute).Format("20060102-150405") + ".log"
	if !present[newestFixture] {
		t.Errorf("expected newest fixture %s to be retained", newestFixture)
	}
}
