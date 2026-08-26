package cleanup

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dustin/go-humanize"

	"github.com/eng618/eng/internal/asdf"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/sysinfo"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

var detectDistro = sysinfo.Detect
var userHomeDir = os.UserHomeDir

// CleanupTask represents a discrete maintenance task.
type CleanupTask struct {
	ID          string
	Label       string
	Category    string
	Execute     func(opts SystemCleanOptions) (int64, string, error)
	IsAvailable func() bool
}

// RunSystemCleanup orchestrates full host maintenance across OS types.
func RunSystemCleanup(opts SystemCleanOptions) (*Report, error) {
	report := &Report{}
	distro := detectDistro()

	log.Verbose(opts.Verbose, "System detected: %s (OS: %s, ID: %s)", distro.PrettyName, distro.RawOS, distro.ID)

	tasks := buildAvailableTasks(distro, opts)
	if len(tasks) == 0 {
		theme.WarningMessage("No cleanup tasks available for this system configuration.")
		return report, nil
	}

	selectedTasks := selectTasksToRun(tasks, opts)
	if len(selectedTasks) == 0 {
		log.Message("No cleanup operations selected.")
		return report, nil
	}

	var multiSpinner *ui.MultiSpinner
	if !ui.DisableProgress && !opts.DryRun {
		multiSpinner, _ = ui.NewMultiSpinner()
		defer func() {
			if multiSpinner != nil {
				multiSpinner.Stop()
			}
		}()
	}

	for _, task := range selectedTasks {
		if opts.DryRun {
			report.Add(ItemResult{
				Name:        task.Label,
				Category:    task.Category,
				Status:      StatusSuccess,
				ReclaimText: "Simulated",
			})
			continue
		}

		var sp ui.ProgressSpinner
		if multiSpinner != nil {
			sp = multiSpinner.AddSpinner(fmt.Sprintf("Running %s...", task.Label))
		}

		freedBytes, outputMsg, err := task.Execute(opts)
		if err != nil {
			if sp != nil {
				sp.Fail(fmt.Sprintf("%s failed: %v", task.Label, err))
			}
			report.Add(ItemResult{
				Name:     task.Label,
				Category: task.Category,
				Status:   StatusFailed,
				Message:  err.Error(),
			})
		} else {
			freedStr := "completed"
			if freedBytes > 0 {
				freedStr = fmt.Sprintf("freed %s", humanize.Bytes(uint64(freedBytes)))
			}
			if outputMsg != "" {
				freedStr = outputMsg
			}

			if sp != nil {
				sp.Success(fmt.Sprintf("%s (%s)", task.Label, freedStr))
			}
			report.Add(ItemResult{
				Name:       task.Label,
				Category:   task.Category,
				Status:     StatusSuccess,
				FreedBytes: freedBytes,
			})
		}
	}

	// If Docker was selected, run Docker pruning and merge into report
	if opts.Docker && IsDockerAvailable() {
		dockerReport, err := RunDockerCleanup(opts.DockerOpts)
		if err == nil && dockerReport != nil {
			for _, item := range dockerReport.Items {
				report.Add(item)
			}
		}
	}

	return report, nil
}

func buildAvailableTasks(distro sysinfo.DistroInfo, opts SystemCleanOptions) []CleanupTask {
	var tasks []CleanupTask

	// 1. Package Manager Cleanup
	if opts.Packages {
		if distro.IsDebianUbuntu() || distro.IsRaspberryPi() {
			tasks = append(tasks, CleanupTask{
				ID:       "apt-cleanup",
				Label:    "APT Package Cache & Autoremove",
				Category: "system",
				IsAvailable: func() bool {
					_, err := lookPath("apt-get")
					return err == nil
				},
				Execute: func(_ SystemCleanOptions) (int64, string, error) {
					cmd := execCommand("sudo", "apt-get", "autoremove", "--purge", "-y")
					_ = cmd.Run()
					cmdClean := execCommand("sudo", "apt-get", "autoclean", "-y")
					_ = cmdClean.Run()
					return 0, "cleaned", nil
				},
			})
		} else if distro.IsFedora() {
			tasks = append(tasks, CleanupTask{
				ID:       "dnf-cleanup",
				Label:    "DNF Package Cache & Autoremove",
				Category: "system",
				IsAvailable: func() bool {
					_, err := lookPath("dnf")
					return err == nil
				},
				Execute: func(_ SystemCleanOptions) (int64, string, error) {
					cmd := execCommand("sudo", "dnf", "autoremove", "-y")
					_ = cmd.Run()
					cmdClean := execCommand("sudo", "dnf", "clean", "all")
					_ = cmdClean.Run()
					return 0, "cleaned", nil
				},
			})
		}
	}

	// 2. Systemd Journal Vacuuming
	if opts.Journal {
		tasks = append(tasks, CleanupTask{
			ID:       "journal-vacuum",
			Label:    "Systemd Journal Log Vacuum",
			Category: "system",
			IsAvailable: func() bool {
				_, err := lookPath("journalctl")
				return err == nil
			},
			Execute: func(o SystemCleanOptions) (int64, string, error) {
				size := o.JournalSize
				if size == "" {
					size = "500M"
				}
				cmd := execCommand("sudo", "journalctl", fmt.Sprintf("--vacuum-size=%s", size))
				var outBuf bytes.Buffer
				cmd.Stdout = &outBuf
				if err := cmd.Run(); err != nil {
					return 0, "", fmt.Errorf("journalctl vacuum failed: %w", err)
				}
				freedBytes := ParseReclaimedBytes(outBuf.String())
				return freedBytes, "", nil
			},
		})
	}

	// 3. Homebrew Cache Cleanup
	if opts.Brew {
		tasks = append(tasks, CleanupTask{
			ID:       "brew-cleanup",
			Label:    "Homebrew Cache & Outdated Downloads",
			Category: "brew",
			IsAvailable: func() bool {
				_, err := lookPath("brew")
				return err == nil
			},
			Execute: func(_ SystemCleanOptions) (int64, string, error) {
				home, _ := userHomeDir()
				brewCache := filepath.Join(home, "Library", "Caches", "Homebrew")
				if _, err := os.Stat(brewCache); os.IsNotExist(err) {
					brewCache = filepath.Join(home, ".cache", "Homebrew")
				}
				sizeBefore := asdf.CalculateDirSize(brewCache)

				cmd := execCommand("brew", "cleanup", "-s")
				_ = cmd.Run()

				sizeAfter := asdf.CalculateDirSize(brewCache)
				freed := sizeBefore - sizeAfter
				if freed < 0 {
					freed = 0
				}
				return freed, "", nil
			},
		})
	}

	// 4. Asdf Plugin / Version Cleanup
	if opts.Asdf {
		tasks = append(tasks, CleanupTask{
			ID:       "asdf-cleanup",
			Label:    "Asdf Version Manager Old Versions",
			Category: "asdf",
			IsAvailable: func() bool {
				_, err := lookPath("asdf")
				return err == nil
			},
			Execute: func(_ SystemCleanOptions) (int64, string, error) {
				home, err := userHomeDir()
				if err != nil {
					return 0, "", err
				}
				listCmd := execCommand("asdf", "list")
				out, err := listCmd.Output()
				if err != nil {
					return 0, "no installed tools", nil
				}
				installed, err := asdf.ParseASDFListOutput(string(out))
				if err != nil || len(installed) == 0 {
					return 0, "no installed tools", nil
				}
				asdfDir := asdf.GetASDFDataDir(home)
				searchRoots := []string{filepath.Join(home, ".tool-versions"), filepath.Join(home, "Development")}
				discoveredFiles, _ := asdf.FindToolVersionFiles(searchRoots)
				protected, _, _ := asdf.ParseAndMergeToolVersions(discoveredFiles)
				removable := asdf.FilterRemovableVersions(installed, protected, "", asdfDir)

				var freedBytes int64
				for _, target := range removable {
					cmd := execCommand("asdf", "uninstall", target.Plugin, target.Version)
					if err := cmd.Run(); err == nil {
						freedBytes += target.SizeBytes
					}
				}
				return freedBytes, "", nil
			},
		})
	}

	var available []CleanupTask
	for _, t := range tasks {
		if t.IsAvailable() {
			available = append(available, t)
		}
	}
	return available
}

func selectTasksToRun(tasks []CleanupTask, opts SystemCleanOptions) []CleanupTask {
	if opts.AutoApprove {
		return tasks
	}

	timeout := opts.CleanupTimeout
	if timeout <= 0 {
		timeout = 30
	}

	resultCh := make(chan []CleanupTask, 1)

	go func() {
		var options []string
		taskMap := make(map[string]CleanupTask)
		for _, t := range tasks {
			options = append(options, t.Label)
			taskMap[t.Label] = t
		}

		selected, err := ui.MultiSelect("Select cleanup operations to run:", options, options)
		if err != nil {
			resultCh <- nil
			return
		}

		var chosen []CleanupTask
		for _, s := range selected {
			if t, ok := taskMap[s]; ok {
				chosen = append(chosen, t)
			}
		}
		resultCh <- chosen
	}()

	select {
	case res := <-resultCh:
		return res
	case <-time.After(time.Duration(timeout) * time.Second):
		log.Message("\nTimeout reached (%d seconds). Auto-selecting all available cleanup operations...", timeout)
		return tasks
	}
}
