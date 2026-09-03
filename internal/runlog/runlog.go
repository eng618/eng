// Package runlog manages per-command session log files.
//
// Verbose commands (git bulk syncs, system updates) produce far more output
// than fits a clean terminal summary. runlog captures the full log stream to
// a timestamped file under the OS cache dir while the terminal shows only the
// summary. Use `eng logs show` to inspect a past run.
//
// File logging is skipped when ui.DisableProgress is set. That flag is only
// ever set by tests, so production runs always log while `go test` never
// touches the real cache directory.
package runlog

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

// KeepRuns is how many recent session logs are retained. Older files are
// pruned whenever a new session starts.
const KeepRuns = 20

// Prefix for session log files.
const filePrefix = "eng-"

// Dir returns the session log directory, creating it if needed.
// ENG_LOG_DIR overrides the default (OS cache dir), which also keeps tests hermetic.
func Dir() (string, error) {
	if override := strings.TrimSpace(os.Getenv("ENG_LOG_DIR")); override != "" {
		if err := os.MkdirAll(override, 0o755); err != nil {
			return "", fmt.Errorf("failed to create log directory %s: %w", override, err)
		}
		return override, nil
	}
	base, err := os.UserCacheDir()
	if err != nil || base == "" {
		base = os.TempDir()
	}
	dir := filepath.Join(base, "eng", "logs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create log directory %s: %w", dir, err)
	}
	return dir, nil
}

// Start begins a file-logging session for the named command (e.g.
// "git-sync-all"). It returns the log file path and a stop function that
// must be deferred by the caller. When file logging is unavailable
// (tests, unwritable cache dir) it returns "" and a no-op stop.
func Start(command string) (string, func()) {
	noop := func() {}
	if ui.DisableProgress {
		return "", noop
	}

	dir, err := Dir()
	if err != nil {
		return "", noop
	}
	prune(dir)

	path := filepath.Join(dir, fileName(command, time.Now()))
	var f *os.File
	for i := 0; i < 100; i++ {
		var err error
		f, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
		if err == nil {
			break
		}
		if !os.IsExist(err) {
			return "", noop
		}
		// Same-second collision (parallel runs): suffix and retry.
		path = strings.TrimSuffix(path, ".log") + fmt.Sprintf("-%d.log", i+2)
		f = nil
	}
	if f == nil {
		return "", noop
	}

	header := fmt.Sprintf("eng %s — %s", command, time.Now().Format(time.RFC3339))
	_, _ = fmt.Fprintln(f, header)
	_, _ = fmt.Fprintln(f, strings.Repeat("=", len(header)))

	log.SetFileLog(f)
	theme.SetFileLog(f)

	return path, func() {
		log.SetFileLog(nil)
		theme.SetFileLog(nil)
		_ = f.Close()
	}
}

// Finish prints the "full log" pointer line. It is a no-op for empty paths,
// so commands can `defer runlog.Finish(path)` unconditionally: the pointer
// always lands after the command summary.
func Finish(path string) {
	if path == "" {
		return
	}
	log.Info("Full log: %s (view with `eng logs show`)", path)
}

func fileName(command string, t time.Time) string {
	var b strings.Builder
	for _, r := range strings.ToLower(command) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	name := strings.Trim(b.String(), "-")
	if name == "" {
		name = "run"
	}
	return fmt.Sprintf("%s%s-%s.log", filePrefix, name, t.Format("20060102-150405"))
}

// prune removes the oldest session logs beyond KeepRuns.
func prune(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	var logs []os.DirEntry
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), filePrefix) && strings.HasSuffix(e.Name(), ".log") {
			logs = append(logs, e)
		}
	}
	if len(logs) < KeepRuns {
		return
	}
	// Oldest first by modtime (names share a timestamp suffix but differ by
	// command, so plain name sorting would group by command, not age).
	infos := make([]os.FileInfo, 0, len(logs))
	for _, e := range logs {
		if info, err := e.Info(); err == nil {
			infos = append(infos, info)
		}
	}
	sort.Slice(infos, func(i, j int) bool {
		if infos[i].ModTime().Equal(infos[j].ModTime()) {
			return infos[i].Name() < infos[j].Name()
		}
		return infos[i].ModTime().Before(infos[j].ModTime())
	})
	for _, info := range infos[:len(infos)-KeepRuns+1] {
		_ = os.Remove(filepath.Join(dir, info.Name()))
	}
}

// Entry describes one session log file.
type Entry struct {
	Name    string
	Path    string
	ModTime time.Time
	Size    int64
}

// List returns session log entries, newest first. Missing dir → empty list.
func List() ([]Entry, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []Entry
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), filePrefix) || !strings.HasSuffix(e.Name(), ".log") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, Entry{
			Name:    e.Name(),
			Path:    filepath.Join(dir, e.Name()),
			ModTime: info.ModTime(),
			Size:    info.Size(),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ModTime.Equal(out[j].ModTime) {
			return out[i].Name > out[j].Name
		}
		return out[i].ModTime.After(out[j].ModTime)
	})
	return out, nil
}

// Resolve finds a log by exact name, unique prefix, or "" for the latest.
func Resolve(name string) (Entry, error) {
	entries, err := List()
	if err != nil {
		return Entry{}, err
	}
	if len(entries) == 0 {
		return Entry{}, fmt.Errorf("no session logs found")
	}
	if name == "" {
		return entries[0], nil
	}
	for _, e := range entries {
		if e.Name == name {
			return e, nil
		}
	}
	var match *Entry
	for i := range entries {
		if strings.HasPrefix(entries[i].Name, name) {
			if match != nil {
				return Entry{}, fmt.Errorf(
					"ambiguous log name %q, matches %q and %q",
					name,
					match.Name,
					entries[i].Name,
				)
			}
			m := entries[i]
			match = &m
		}
	}
	if match == nil {
		return Entry{}, fmt.Errorf("no session log matches %q (see `eng logs list`)", name)
	}
	return *match, nil
}

// ParseName splits a session log filename into its command slug and timestamp.
// Unknown shapes return the raw base name and the zero time.
func ParseName(name string) (string, time.Time) {
	rest, ok := strings.CutPrefix(name, filePrefix)
	if !ok {
		return name, time.Time{}
	}
	rest = strings.TrimSuffix(rest, ".log")
	// Timestamp is the trailing YYYYMMDD-HHMMSS, optionally followed by a -N
	// same-second collision suffix. Try direct parse first so command slugs
	// ending in numbers (e.g. "run-2") are not mistaken for the suffix.
	const tsLen = len("20060102-150405")
	if cmd, ts, ok := splitTimestamp(rest, tsLen); ok {
		return cmd, ts
	}
	if i := strings.LastIndex(rest, "-"); i > 0 {
		if _, err := strconv.Atoi(rest[i+1:]); err == nil {
			if cmd, ts, ok := splitTimestamp(rest[:i], tsLen); ok {
				return cmd, ts
			}
		}
	}
	return rest, time.Time{}
}

func splitTimestamp(rest string, tsLen int) (string, time.Time, bool) {
	if len(rest) <= tsLen {
		return "", time.Time{}, false
	}
	ts, err := time.Parse("20060102-150405", rest[len(rest)-tsLen:])
	if err != nil {
		return "", time.Time{}, false
	}
	return strings.Trim(rest[:len(rest)-tsLen], "-"), ts, true
}
