package logs

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/runlog"
	"github.com/eng618/eng/internal/ui/theme"
)

var (
	showTail   int
	showFollow bool
)

// ShowCmd prints a session log file.
var ShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show a session log (latest by default)",
	Long: `Prints a captured session log. With no name, shows the latest run.
Names accept exact matches or unique prefixes (see 'eng logs list').

Use --tail N to show only the last N lines, or --follow to stream new
lines like 'tail -f' until interrupted.`,
	Example: `  eng logs show
  eng logs show git-sync-all
  eng logs show --tail 50
  eng logs show system-update -f`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := ""
		if len(args) > 0 {
			name = args[0]
		}
		entry, err := runlog.Resolve(name)
		if err != nil {
			return err
		}
		if showFollow {
			return followLog(cmd, entry.Path)
		}
		return printLog(entry.Path, showTail)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		entries, err := runlog.List()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		var names []string
		for _, e := range entries {
			if strings.HasPrefix(e.Name, toComplete) {
				names = append(names, e.Name)
			}
		}
		return names, cobra.ShellCompDirectiveNoFileComp
	},
}

func init() {
	ShowCmd.Flags().IntVar(&showTail, "tail", 0, "Show only the last N lines (0 = whole file)")
	ShowCmd.Flags().BoolVarP(&showFollow, "follow", "f", false, "Stream new lines until interrupted")
}

func printLog(path string, tail int) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read log %s: %w", path, err)
	}
	text := string(data)
	if tail > 0 {
		lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
		if len(lines) > tail {
			lines = lines[len(lines)-tail:]
		}
		text = strings.Join(lines, "\n") + "\n"
		theme.InfoMessage(fmt.Sprintf("Showing last %d lines of %s", tail, path))
	} else {
		theme.InfoMessage(fmt.Sprintf("Showing %s", path))
	}
	_, _ = fmt.Fprint(log.Out, text)
	return nil
}

func followLog(cmd *cobra.Command, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open log %s: %w", path, err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat log %s: %w", path, err)
	}
	offset := info.Size()
	if showTail > 0 {
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read log %s: %w", path, err)
		}
		lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
		if len(lines) > showTail {
			lines = lines[len(lines)-showTail:]
		}
		_, _ = fmt.Fprintln(log.Out, strings.Join(lines, "\n"))
	}

	_, _ = fmt.Fprintln(log.Err, theme.MutedText.Render(fmt.Sprintf("Following %s (Ctrl+C to stop)…", path)))

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	buf := make([]byte, 32*1024)
	for {
		select {
		case <-cmd.Context().Done():
			return nil
		case <-ticker.C:
			for {
				n, err := f.ReadAt(buf, offset)
				if n > 0 {
					offset += int64(n)
					if _, werr := log.Out.Write(buf[:n]); werr != nil {
						return werr
					}
				}
				if err == io.EOF {
					break
				}
				if err != nil {
					return fmt.Errorf("failed to follow log %s: %w", path, err)
				}
			}
		}
	}
}
