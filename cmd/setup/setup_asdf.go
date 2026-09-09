package setup

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/cmdutil"
	"github.com/eng618/eng/internal/log"
)

var SetupASDFCmd = &cobra.Command{
	Use:   "asdf",
	Short: "Setup asdf plugins from $HOME/.tool-versions",
	Long:  `Reads $HOME/.tool-versions and installs asdf plugins listed there.`,
	Run: func(cmd *cobra.Command, args []string) {
		setupASDF(cmdutil.IsVerbose(cmd))
	},
}

// ensureASDFInstalled verifies that asdf is installed, installing it via Homebrew if available or via git clone.
func ensureASDFInstalled(verbose bool) error {
	if _, err := lookPath("asdf"); err == nil {
		log.Verbose(verbose, "asdf is already installed")
		return nil
	}

	// Check if Homebrew is available
	_, brewErr := lookPath("brew")
	if brewErr != nil {
		if _, err := stat("/home/linuxbrew/.linuxbrew/bin/brew"); err == nil {
			_ = os.Setenv("PATH", os.Getenv("PATH")+":/home/linuxbrew/.linuxbrew/bin")
			brewErr = nil
		}
	}

	if brewErr == nil {
		log.Start("Installing asdf via Homebrew...")
		cmd := execCommand("brew", "install", "asdf")
		cmd.Stdout = log.Writer()
		cmd.Stderr = log.ErrorWriter()
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install asdf via Homebrew: %w", err)
		}

		// Update PATH if brew installed asdf to a specific prefix
		if _, err := lookPath("asdf"); err != nil {
			cmdPrefix := execCommand("brew", "--prefix", "asdf")
			if out, err := cmdPrefix.Output(); err == nil {
				asdfPrefix := strings.TrimSpace(string(out))
				asdfBin := filepath.Join(asdfPrefix, "bin")
				asdfLibexecBin := filepath.Join(asdfPrefix, "libexec", "bin")
				if _, err := stat(asdfBin); err == nil {
					_ = os.Setenv("PATH", os.Getenv("PATH")+":"+asdfBin)
				} else if _, err := stat(asdfLibexecBin); err == nil {
					_ = os.Setenv("PATH", os.Getenv("PATH")+":"+asdfLibexecBin)
				}
			}
		}

		log.Success("asdf installed successfully via Homebrew")
		return nil
	}

	// Check for existing ~/.asdf
	homeDir, err := userHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	asdfDir := filepath.Join(homeDir, ".asdf")
	if _, err := stat(asdfDir); err == nil {
		_ = os.Setenv("PATH", os.Getenv("PATH")+":"+filepath.Join(asdfDir, "bin"))
		log.Verbose(verbose, "Found existing asdf directory at %s", asdfDir)
		return nil
	}

	if _, err := lookPath("git"); err != nil {
		return fmt.Errorf("neither Homebrew nor Git is available to install asdf")
	}

	log.Start("Installing asdf via Git clone into ~/.asdf...")
	cmd := execCommand("git", "clone", "https://github.com/asdf-vm/asdf.git", asdfDir, "--branch", "v0.14.0")
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.ErrorWriter()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to git clone asdf: %w", err)
	}

	_ = os.Setenv("PATH", os.Getenv("PATH")+":"+filepath.Join(asdfDir, "bin"))
	log.Success("asdf installed via Git to ~/.asdf")
	return nil
}

func setupASDF(verbose bool) {
	log.Verbose(verbose, "Starting ASDF setup...")

	if err := ensureASDFInstalled(verbose); err != nil {
		log.Error("Could not setup asdf: %v", err)
		return
	}

	homeDir, err := userHomeDir()
	if err != nil {
		log.Error("Could not determine home directory: %v", err)
		return
	}

	toolVersionsPath := filepath.Join(homeDir, ".tool-versions")
	if _, err := stat(toolVersionsPath); err != nil {
		if _, currErr := stat(".tool-versions"); currErr == nil {
			toolVersionsPath = ".tool-versions"
		} else {
			log.Message("No .tool-versions file found in %s; skipping asdf plugin setup", homeDir)
			return
		}
	}

	file, err := os.Open(toolVersionsPath)
	if err != nil {
		log.Error("Could not open %s: %v", toolVersionsPath, err)
		return
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Error("Error closing file %s: %v", toolVersionsPath, cerr)
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 1 {
			continue
		}
		plugin := fields[0]
		cmd := execCommand("asdf", "plugin", "add", plugin)
		out, err := cmd.CombinedOutput()
		outStr := strings.TrimSpace(string(out))
		if err != nil {
			if strings.Contains(strings.ToLower(outStr), "already") {
				log.Verbose(verbose, "asdf plugin '%s' already added", plugin)
			} else {
				log.Error("Failed to add asdf plugin '%s': %v (%s)", plugin, err, outStr)
			}
		} else {
			log.Success("Added asdf plugin: %s", plugin)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Error("Error reading %s: %v", toolVersionsPath, err)
		return
	}
	// Install all plugins
	installCmd := execCommand("asdf", "install")
	installCmd.Stdout = log.Writer()
	installCmd.Stderr = log.ErrorWriter()
	log.Start("Running 'asdf install' to install all plugins...")
	if err := installCmd.Run(); err != nil {
		log.Error("Failed to run 'asdf install': %v", err)
	} else {
		log.Success("All asdf plugins installed successfully.")
	}
}
