package system

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/asdf"
	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

// UpdateCmd represents the system update command.
var UpdateCmd = &cobra.Command{
	Use:     "update",
	Aliases: []string{"update", "u"},
	Short:   "Update the system and perform maintenance",
	Long:    `This command updates the system, Homebrew packages, asdf plugins, and performs cleanup operations.`,
	Run: func(cmd *cobra.Command, _args []string) {
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary).
			MarginBottom(1)
		if !ui.DisableProgress {
			fmt.Println(headerStyle.Render("🚀 System Maintenance & Update"))
		}

		isVerbose := cmdutil.IsVerbose(cmd)
		autoApprove, _ := cmd.Flags().GetBool("yes")
		cleanupTimeout, _ := cmd.Flags().GetInt("cleanup-timeout")
		log.Verbose(isVerbose, "Checking system type...")

		checkCmd := execCommand("uname", "-a")
		output, err := checkCmd.Output()
		if err != nil {
			log.Error("Error checking system type: %s", err)
			return
		}

		uname := strings.ToLower(string(output))
		log.Verbose(isVerbose, "System type detected: %s", strings.TrimSpace(string(output)))

		if strings.Contains(uname, "ubuntu") || strings.Contains(uname, "linux") {
			log.Verbose(isVerbose, "Detected Ubuntu/Linux system, running system update...")
			updateUbuntu(isVerbose, autoApprove, cleanupTimeout)
		} else if strings.Contains(uname, "darwin") {
			log.Verbose(isVerbose, "Detected macOS system, running macOS update...")
			updateMacOS(isVerbose)
		} else if strings.Contains(uname, "raspberrypi") || strings.Contains(uname, "raspbian") {
			log.Verbose(isVerbose, "Detected Raspberry Pi system, running Raspberry Pi update...")
			updateRaspberryPi(isVerbose)
		} else {
			log.Warn("This system is not yet supported for updates.")
			log.Verbose(isVerbose, "Unsupported system type: %s", strings.TrimSpace(string(output)))
		}
	},
}

// BrewCmd represents the Homebrew update subcommand.
var BrewCmd = &cobra.Command{
	Use:   "brew",
	Short: "Update Homebrew packages only",
	Long:  `This command updates only Homebrew packages, skipping system updates.`,
	Run: func(cmd *cobra.Command, _args []string) {
		isVerbose := cmdutil.IsVerbose(cmd)
		updateBrew(isVerbose)
	},
}

func init() {
	UpdateCmd.Flags().BoolP("yes", "y", false, "Auto-approve cleanup operations without prompting")
	UpdateCmd.Flags().Int("cleanup-timeout", 60, "Timeout in seconds for cleanup confirmation prompt")
	UpdateCmd.AddCommand(BrewCmd)
}

func updateUbuntu(isVerbose, autoApprove bool, cleanupTimeout int) {
	log.Message("Running system update for Ubuntu/Linux...")
	log.Message("About to run a command with sudo. You may be prompted for your system password.")

	updateCmd := execCommand("bash", "-c", "sudo apt-get update && sudo apt-get upgrade -y")
	updateCmd.Stdout = log.Writer()
	updateCmd.Stderr = log.ErrorWriter()
	if err := updateCmd.Run(); err != nil {
		log.Error("Error updating system: %s", err)
		return
	}
	log.Success("System updated successfully.")

	runCleanup(isVerbose, autoApprove, cleanupTimeout)
	updateBrew(isVerbose)
	updateAsdf(isVerbose)
}

func updateMacOS(isVerbose bool) {
	updateBrew(isVerbose)
	updateAsdf(isVerbose)
}

func updateRaspberryPi(isVerbose bool) {
	updateBrew(isVerbose)
	updateAsdf(isVerbose)
}

func updateBrew(isVerbose bool) {
	_, err := lookPath("brew")
	if err != nil {
		log.Message("Homebrew (brew) is not installed on this system.")
		return
	}

	var spinner *ui.Spinner
	if !ui.DisableProgress {
		spinner = ui.NewSpinner("Running Homebrew update, upgrade, and cleanup...")
	}

	// Measure brew cache size before cleanup
	homeDir, _ := userHomeDir()
	brewCacheDir := filepath.Join(homeDir, "Library", "Caches", "Homebrew")
	if _, err := os.Stat(brewCacheDir); os.IsNotExist(err) {
		brewCacheDir = filepath.Join(homeDir, ".cache", "Homebrew")
	}

	sizeBefore := asdf.CalculateDirSize(brewCacheDir)

	updateCmd := execCommand("bash", "-c", "brew update && brew outdated && brew upgrade && brew cleanup")
	updateCmd.Stdout = log.Writer()
	updateCmd.Stderr = log.ErrorWriter()

	if err := updateCmd.Run(); err != nil {
		if spinner != nil {
			spinner.Stop()
		}
		log.Error("Error updating Homebrew packages: %s", err)
	} else {
		if spinner != nil {
			spinner.Stop()
		}
		sizeAfter := asdf.CalculateDirSize(brewCacheDir)
		freedBytes := sizeBefore - sizeAfter
		freedStr := ""
		if freedBytes > 0 {
			freedStr = fmt.Sprintf(" (freed %s cache)", humanize.Bytes(uint64(freedBytes)))
		}
		theme.SuccessMessage(fmt.Sprintf("Homebrew packages updated and cleaned successfully!%s", freedStr))
	}
}

func updateAsdf(isVerbose bool) {
	_, err := lookPath("asdf")
	if err != nil {
		log.Message("asdf version manager is not installed on this system.")
		return
	}

	var spinner *ui.Spinner
	if !ui.DisableProgress {
		spinner = ui.NewSpinner("Updating asdf plugins...")
	}

	updateCmd := execCommand("asdf", "plugin", "update", "--all")
	updateCmd.Stdout = log.Writer()
	updateCmd.Stderr = log.ErrorWriter()
	if err := updateCmd.Run(); err != nil {
		if spinner != nil {
			spinner.Stop()
		}
		log.Error("Error updating asdf plugins: %s", err)
	} else {
		if spinner != nil {
			spinner.Stop()
		}
		theme.SuccessMessage("asdf plugins updated successfully.")
	}
}

func runCleanup(isVerbose, autoApprove bool, cleanupTimeout int) {
	operations := []string{
		"apt-get autoremove --purge",
		"apt-get autoclean",
	}

	_, dockerErr := lookPath("docker")
	if dockerErr == nil {
		operations = append(operations, "docker system prune")
	}

	var selectedOperations []string
	if autoApprove {
		selectedOperations = operations
	} else {
		resultCh := make(chan []string, 1)

		go func() {
			options := []string{
				"Remove unneeded packages (apt-get autoremove --purge)",
				"Clean package cache (apt-get autoclean)",
			}
			if dockerErr == nil {
				options = append(options, "Prune Docker system (docker system prune)")
			}

			selected, err := ui.MultiSelect("Select cleanup operations to run:", options, options)
			if err != nil {
				resultCh <- nil
				return
			}

			var mapped []string
			for _, s := range selected {
				if strings.Contains(s, "autoremove") {
					mapped = append(mapped, "apt-get autoremove --purge")
				}
				if strings.Contains(s, "autoclean") {
					mapped = append(mapped, "apt-get autoclean")
				}
				if strings.Contains(s, "Docker") {
					mapped = append(mapped, "docker system prune")
				}
			}
			resultCh <- mapped
		}()

		select {
		case res := <-resultCh:
			selectedOperations = res
		case <-time.After(time.Duration(cleanupTimeout) * time.Second):
			log.Message("\nTimeout reached (%d seconds). Auto-selecting all cleanup operations...", cleanupTimeout)
			selectedOperations = operations
		}
	}

	if len(selectedOperations) == 0 {
		log.Message("No cleanup operations selected.")
		return
	}

	multiSpinner, _ := ui.NewMultiSpinner()
	defer func() {
		if multiSpinner != nil {
			multiSpinner.Stop()
		}
	}()

	for _, op := range selectedOperations {
		var opSpinner ui.ProgressSpinner
		if multiSpinner != nil {
			opSpinner = multiSpinner.AddSpinner(fmt.Sprintf("Running %s...", op))
		}

		if op == "docker system prune" {
			pruneCmd := execCommand("docker", "system", "prune", "-f")
			out, err := pruneCmd.CombinedOutput()
			if err != nil {
				if opSpinner != nil {
					opSpinner.Fail(fmt.Sprintf("Docker prune failed: %v", err))
				} else {
					log.Error("Docker prune failed: %v", err)
				}
			} else {
				reclaimedStr := parseDockerPruneSpace(string(out))
				msg := "Docker system prune completed."
				if reclaimedStr != "" {
					msg = fmt.Sprintf("Docker system prune completed (freed %s).", reclaimedStr)
				}
				if opSpinner != nil {
					opSpinner.Success(msg)
				} else {
					theme.SuccessMessage(msg)
				}
			}
			continue
		}

		cmdStr := fmt.Sprintf("sudo %s -y", op)
		cleanupCmd := execCommand("bash", "-c", cmdStr)
		cleanupCmd.Stdout = log.Writer()
		cleanupCmd.Stderr = log.ErrorWriter()

		if err := cleanupCmd.Run(); err != nil {
			if opSpinner != nil {
				opSpinner.Fail(fmt.Sprintf("Failed %s: %v", op, err))
			} else {
				log.Error("Failed %s: %v", op, err)
			}
		} else {
			if opSpinner != nil {
				opSpinner.Success(fmt.Sprintf("Completed %s", op))
			} else {
				log.Success("Completed %s", op)
			}
		}
	}
}

var dockerReclaimRe = regexp.MustCompile(`Total reclaimed space:\s*([\d\.]+\s*[KMGT]?B)`)

func parseDockerPruneSpace(output string) string {
	match := dockerReclaimRe.FindStringSubmatch(output)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

func runCleanupOperationWithSpinner(isVerbose bool, command, operationName string, spinner ui.ProgressSpinner) {
	log.Verbose(isVerbose, "Running: %s", command)

	cleanupCmd := execCommand("bash", "-c", command)
	cleanupCmd.Stdout = log.Writer()
	cleanupCmd.Stderr = log.ErrorWriter()

	if err := cleanupCmd.Run(); err != nil {
		spinner.Fail(fmt.Sprintf("Error running %s: %s", operationName, err))
	} else {
		spinner.Success(fmt.Sprintf("%s completed", operationName))
	}
}

func runCleanupOperation(isVerbose bool, command, operationName string) {
	log.Verbose(isVerbose, "Running: %s", command)

	progress := ui.NewProgressSpinner(fmt.Sprintf("Running %s...", operationName))
	cleanupCmd := execCommand("bash", "-c", command)
	cleanupCmd.Stdout = log.Writer()
	cleanupCmd.Stderr = log.ErrorWriter()

	if err := cleanupCmd.Run(); err != nil {
		progress.Stop()
		log.Error("Error running %s: %s", operationName, err)
	} else {
		progress.SetProgressBar(1.0, fmt.Sprintf("%s completed", operationName))
		progress.Stop()
		log.Success("%s completed.", operationName)
	}
}
