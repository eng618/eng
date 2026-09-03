package system

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/asdf"
	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/sysinfo"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

var detectDistro = sysinfo.Detect

// UpdateCmd represents the system update command.
var UpdateCmd = &cobra.Command{
	Use:     "update",
	Aliases: []string{"upgrade", "u"},
	Short:   "Update the system and perform maintenance",
	Long:    `This command updates the system, Homebrew packages, asdf plugins, flatpak packages, and performs cleanup operations.`,
	Run: func(cmd *cobra.Command, _args []string) {
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary).
			MarginBottom(1)
		if !ui.DisableProgress {
			fmt.Fprintln(log.Out, headerStyle.Render("🚀 System Maintenance & Update"))
		}

		isVerbose := cmdutil.IsVerbose(cmd)
		autoApprove, _ := cmd.Flags().GetBool("yes")
		cleanupTimeout, _ := cmd.Flags().GetInt("cleanup-timeout")
		log.Verbose(isVerbose, "Checking system type...")

		distro := detectDistro()
		log.Verbose(isVerbose, "System detected: %s (OS: %s, ID: %s)", distro.PrettyName, distro.RawOS, distro.ID)

		if distro.IsFedora() {
			log.Verbose(isVerbose, "Detected Fedora/RHEL system, running DNF system update...")
			updateFedora(isVerbose, autoApprove, cleanupTimeout)
		} else if distro.IsDebianUbuntu() {
			log.Verbose(isVerbose, "Detected Debian/Ubuntu system, running APT system update...")
			updateDebianUbuntu(isVerbose, autoApprove, cleanupTimeout)
		} else if distro.IsMacOS() {
			log.Verbose(isVerbose, "Detected macOS system, running macOS update...")
			updateMacOS(isVerbose)
		} else if distro.IsRaspberryPi() {
			log.Verbose(isVerbose, "Detected Raspberry Pi system, running Raspberry Pi update...")
			updateRaspberryPi(isVerbose, autoApprove, cleanupTimeout)
		} else if distro.IsLinux() {
			log.Verbose(isVerbose, "Detected generic Linux system, running general maintenance...")
			updateGenericLinux(isVerbose, autoApprove, cleanupTimeout)
		} else {
			log.Warn("This system is not yet configured for automated OS updates.")
			log.Verbose(isVerbose, "System type: %s", distro.PrettyName)
			updateBrew(isVerbose)
			updateAsdf(isVerbose)
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
	UpdateCmd.AddCommand(UpdateIdeCmd)
}

func updateFedora(isVerbose, autoApprove bool, cleanupTimeout int) {
	log.Message("Running system update for Fedora/RHEL...")
	log.Message("About to run a command with sudo. You may be prompted for your system password.")

	updateCmd := execCommand("bash", "-c", "sudo dnf upgrade --refresh -y")
	updateCmd.Stdout = log.Writer()
	updateCmd.Stderr = log.ErrorWriter()
	if err := updateCmd.Run(); err != nil {
		log.Error("Error updating system with DNF: %s", err)
		return
	}
	log.Success("System updated successfully.")

	runDnfCleanup(isVerbose, autoApprove, cleanupTimeout)
	updateFlatpak(isVerbose)
	updateBrew(isVerbose)
	updateAsdf(isVerbose)
	updateIde(isVerbose, autoApprove)
}

func updateDebianUbuntu(isVerbose, autoApprove bool, cleanupTimeout int) {
	log.Message("Running system update for Ubuntu/Debian...")
	log.Message("About to run a command with sudo. You may be prompted for your system password.")

	updateCmd := execCommand("bash", "-c", "sudo apt-get update && sudo apt-get upgrade -y")
	updateCmd.Stdout = log.Writer()
	updateCmd.Stderr = log.ErrorWriter()
	if err := updateCmd.Run(); err != nil {
		log.Error("Error updating system with APT: %s", err)
		return
	}
	log.Success("System updated successfully.")

	runAptCleanup(isVerbose, autoApprove, cleanupTimeout)
	updateFlatpak(isVerbose)
	updateBrew(isVerbose)
	updateAsdf(isVerbose)
	updateIde(isVerbose, autoApprove)
}

func updateGenericLinux(isVerbose, autoApprove bool, cleanupTimeout int) {
	if _, err := lookPath("dnf"); err == nil {
		updateFedora(isVerbose, autoApprove, cleanupTimeout)
		return
	}
	if _, err := lookPath("apt-get"); err == nil {
		updateDebianUbuntu(isVerbose, autoApprove, cleanupTimeout)
		return
	}

	updateFlatpak(isVerbose)
	updateBrew(isVerbose)
	updateAsdf(isVerbose)
	updateIde(isVerbose, autoApprove)
}

func updateMacOS(isVerbose bool) {
	updateBrew(isVerbose)
	updateAsdf(isVerbose)
}

func updateRaspberryPi(isVerbose, autoApprove bool, cleanupTimeout int) {
	updateDebianUbuntu(isVerbose, autoApprove, cleanupTimeout)
}

func updateFlatpak(isVerbose bool) {
	_, err := lookPath("flatpak")
	if err != nil {
		return
	}

	log.Verbose(isVerbose, "Checking for Flatpak updates...")
	var spinner *ui.Spinner
	if !ui.DisableProgress {
		spinner = ui.NewSpinner("Updating Flatpak packages...")
		spinner.Start()
	}

	updateCmd := execCommand("flatpak", "update", "-y")
	if isVerbose {
		updateCmd.Stdout = log.Writer()
		updateCmd.Stderr = log.ErrorWriter()
	} else {
		var stderrBuf bytes.Buffer
		updateCmd.Stderr = &stderrBuf
	}

	runErr := updateCmd.Run()
	if spinner != nil {
		spinner.Stop()
	}

	if runErr != nil {
		log.Verbose(isVerbose, "Flatpak update notice: %v", runErr)
	} else {
		theme.SuccessMessage("Flatpak packages updated successfully.")
	}
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
		spinner.Start()
	}

	// Measure brew cache size before cleanup
	homeDir, _ := userHomeDir()
	brewCacheDir := filepath.Join(homeDir, "Library", "Caches", "Homebrew")
	if _, err := os.Stat(brewCacheDir); os.IsNotExist(err) {
		brewCacheDir = filepath.Join(homeDir, ".cache", "Homebrew")
	}

	sizeBefore := asdf.CalculateDirSize(brewCacheDir)

	updateCmd := execCommand("bash", "-c", "brew update && brew outdated && brew upgrade && brew cleanup")
	if isVerbose {
		updateCmd.Stdout = log.Writer()
		updateCmd.Stderr = log.ErrorWriter()
	} else {
		var stderrBuf bytes.Buffer
		updateCmd.Stderr = &stderrBuf
	}

	runErr := updateCmd.Run()
	if spinner != nil {
		spinner.Stop()
	}

	if runErr != nil {
		log.Error("Error updating Homebrew packages: %s", runErr)
	} else {
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

type cleanupOperation struct {
	Label   string
	Command string
	IsRaw   bool
}

func runDnfCleanup(isVerbose, autoApprove bool, cleanupTimeout int) {
	ops := []cleanupOperation{
		{Label: "Remove unneeded packages (dnf autoremove)", Command: "dnf autoremove"},
		{Label: "Clean package cache (dnf clean all)", Command: "dnf clean all"},
	}
	if _, dockerErr := lookPath("docker"); dockerErr == nil {
		ops = append(
			ops,
			cleanupOperation{
				Label:   "Prune Docker system (docker system prune)",
				Command: "docker system prune",
				IsRaw:   true,
			},
		)
	}
	executeCleanupOperations(ops, isVerbose, autoApprove, cleanupTimeout)
}

func runAptCleanup(isVerbose, autoApprove bool, cleanupTimeout int) {
	ops := []cleanupOperation{
		{Label: "Remove unneeded packages (apt-get autoremove --purge)", Command: "apt-get autoremove --purge"},
		{Label: "Clean package cache (apt-get autoclean)", Command: "apt-get autoclean"},
	}
	if _, dockerErr := lookPath("docker"); dockerErr == nil {
		ops = append(
			ops,
			cleanupOperation{
				Label:   "Prune Docker system (docker system prune)",
				Command: "docker system prune",
				IsRaw:   true,
			},
		)
	}
	executeCleanupOperations(ops, isVerbose, autoApprove, cleanupTimeout)
}

func executeCleanupOperations(availableOps []cleanupOperation, isVerbose, autoApprove bool, cleanupTimeout int) {
	var selectedOps []cleanupOperation
	if autoApprove {
		selectedOps = availableOps
	} else {
		resultCh := make(chan []cleanupOperation, 1)

		go func() {
			var options []string
			opMap := make(map[string]cleanupOperation)
			for _, op := range availableOps {
				options = append(options, op.Label)
				opMap[op.Label] = op
			}

			selected, err := ui.MultiSelect("Select cleanup operations to run:", options, options)
			if err != nil {
				resultCh <- nil
				return
			}

			var mapped []cleanupOperation
			for _, s := range selected {
				if op, ok := opMap[s]; ok {
					mapped = append(mapped, op)
				}
			}
			resultCh <- mapped
		}()

		select {
		case res := <-resultCh:
			selectedOps = res
		case <-time.After(time.Duration(cleanupTimeout) * time.Second):
			log.Message("\nTimeout reached (%d seconds). Auto-selecting all cleanup operations...", cleanupTimeout)
			selectedOps = availableOps
		}
	}

	if len(selectedOps) == 0 {
		log.Message("No cleanup operations selected.")
		return
	}

	multiSpinner, _ := ui.NewMultiSpinner()
	defer func() {
		if multiSpinner != nil {
			multiSpinner.Stop()
		}
	}()

	for _, op := range selectedOps {
		var opSpinner ui.ProgressSpinner
		if multiSpinner != nil {
			opSpinner = multiSpinner.AddSpinner(fmt.Sprintf("Running %s...", op.Command))
		}

		if op.Command == "docker system prune" {
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

		args := append([]string{"sudo"}, strings.Fields(op.Command)...)
		args = append(args, "-y")
		cleanupCmd := execCommand(args[0], args[1:]...)
		cleanupCmd.Stdout = log.Writer()
		cleanupCmd.Stderr = log.ErrorWriter()

		if err := cleanupCmd.Run(); err != nil {
			if opSpinner != nil {
				opSpinner.Fail(fmt.Sprintf("Failed %s: %v", op.Command, err))
			} else {
				log.Error("Failed %s: %v", op.Command, err)
			}
		} else {
			if opSpinner != nil {
				opSpinner.Success(fmt.Sprintf("Completed %s", op.Command))
			} else {
				log.Success("Completed %s", op.Command)
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

func updateIde(isVerbose, autoApprove bool) {
	homeDir, err := userHomeDir()
	if err != nil {
		return
	}
	installDir := filepath.Join(homeDir, ".local", "opt", "antigravity-ide")
	_, errStat := stat(installDir)
	_, errPath := lookPath("agy-ide")

	if errStat != nil && errPath != nil {
		log.Verbose(isVerbose, "Antigravity IDE is not installed, skipping IDE update.")
		return
	}

	downloadsDir := filepath.Join(homeDir, "Downloads")
	archive := findLatestIdeArchive(downloadsDir)
	configURL := strings.TrimSpace(viper.GetString("antigravity.ide_download_url"))

	if archive == "" && configURL == "" {
		log.Verbose(isVerbose, "No pending Antigravity IDE archive or URL found, skipping IDE update.")
		return
	}

	log.Message("Checking for Antigravity IDE updates...")
	if err := RunIdeUpdate(context.Background(), "", isVerbose, autoApprove); err != nil {
		log.Warn("Antigravity IDE update check: %v", err)
	}
}
