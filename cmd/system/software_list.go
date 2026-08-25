package system

import (
	"context"
	"path/filepath"
	"runtime"

	"github.com/eng618/eng/internal/log"
)

type Software struct {
	Name        string
	Description string
	Optional    bool
	URL         string      // For manual downloads
	Check       func() bool // Returns true if already installed
	// Install returns an error if installation fails.
	// For manual installs, this typically involves opening a URL.
	Install func() error
	// OS restriction (empty means both, otherwise "linux" or "darwin")
	OS string
}

func openURL(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // linux, freebsd, openbsd, netbsd
		cmd = "xdg-open"
	}
	args = append(args, url)
	return execCommand(cmd, args...).Start()
}

// Helpers for checks and installs

func checkFalse() bool {
	return false
}

func installByURL(url string) func() error {
	return func() error {
		return openURL(url)
	}
}

func checkByPath(path string) func() bool {
	return func() bool {
		_, err := lookPath(path)
		return err == nil
	}
}

func checkByBundleID(bundleID string) func() bool {
	return func() bool {
		if runtime.GOOS == "darwin" {
			return execCommand("mdfind", "kMDItemCFBundleIdentifier == '"+bundleID+"'").Run() == nil
		}
		return false
	}
}

func checkByBundleIDOrPath(bundleID string, linuxPaths ...string) func() bool {
	return func() bool {
		if runtime.GOOS == "darwin" {
			return execCommand("mdfind", "kMDItemCFBundleIdentifier == '"+bundleID+"'").Run() == nil
		}
		for _, path := range linuxPaths {
			if _, err := lookPath(path); err == nil {
				return true
			}
		}
		// Check Flatpak system and user application paths
		if _, err := stat("/var/lib/flatpak/app/" + bundleID); err == nil {
			return true
		}
		homeDir, err := userHomeDir()
		if err == nil {
			if _, err := stat(filepath.Join(homeDir, ".local", "share", "flatpak", "app", bundleID)); err == nil {
				return true
			}
		}
		return false
	}
}

func checkObsidian() bool {
	if runtime.GOOS == "darwin" {
		return execCommand(
			"mdfind",
			"kMDItemCFBundleIdentifier == 'com.obsidian.md' || kMDItemCFBundleIdentifier == 'md.obsidian'",
		).Run() == nil
	}
	if _, err := lookPath("obsidian"); err == nil {
		return true
	}
	if _, err := stat("/var/lib/flatpak/app/md.obsidian.Obsidian"); err == nil {
		return true
	}
	homeDir, err := userHomeDir()
	if err == nil {
		if _, err := stat(
			filepath.Join(homeDir, ".local", "share", "flatpak", "app", "md.obsidian.Obsidian"),
		); err == nil {
			return true
		}
	}
	return false
}

func checkAndroidStudio() bool {
	if runtime.GOOS == "darwin" {
		_, err := stat("/Applications/Android Studio.app")
		return err == nil
	}
	if runtime.GOOS == "linux" {
		if _, err := lookPath("studio"); err == nil {
			return true
		}
		if _, err := lookPath("studio.sh"); err == nil {
			return true
		}
		if _, err := stat("/opt/android-studio/bin/studio.sh"); err == nil {
			return true
		}
		if _, err := stat("/var/lib/flatpak/app/com.google.AndroidStudio"); err == nil {
			return true
		}
		homeDir, err := userHomeDir()
		if err == nil {
			if _, err := stat(
				filepath.Join(homeDir, ".local", "share", "flatpak", "app", "com.google.AndroidStudio"),
			); err == nil {
				return true
			}
		}
	}
	return false
}

func getCoreSoftware() []Software {
	return []Software{
		// Critical / Core Items
		// VS Code is needed before Brew Bundle for extensions
		{
			Name:        "VS Code",
			Description: "Code Editor",
			Optional:    true,
			URL:         "https://code.visualstudio.com/Download",
			Check:       checkByBundleIDOrPath("com.microsoft.VSCode", "code"),
			Install:     installByURL("https://code.visualstudio.com/Download"),
		},
		{
			Name:        "Brew Bundle",
			Description: "Install software from Brewfile",
			Optional:    false,
			Check: func() bool {
				// Brew bundle check returns 0 if satisfied, 1 if not
				cmd := execCommand("brew", "bundle", "check")
				return cmd.Run() == nil
			},
			Install: func() error {
				log.Info("Running brew bundle install...")
				cmd := execCommand("brew", "bundle", "install")
				cmd.Stdout = log.Writer()
				cmd.Stderr = log.ErrorWriter()
				return cmd.Run()
			},
			OS: "darwin", // Brewfile is mostly Mac formatted in this repo
		},
		{
			Name:        "Oh My Zsh",
			Description: "Zsh configuration framework",
			Optional:    false,
			Check: func() bool {
				homeDir, err := userHomeDir()
				if err != nil {
					return false
				}
				_, err = stat(filepath.Join(homeDir, ".oh-my-zsh"))
				return err == nil
			},
			Install: func() error {
				return nil
			},
		},
	}
}

func getSecurityAndPrivacyApps() []Software {
	return []Software{
		{
			Name:        "Ente",
			Description: "Encrypted Photo Backup",
			Optional:    true,
			URL:         "https://ente.io/download",
			Check:       checkByBundleIDOrPath("io.ente.photos", "ente"),
			Install:     installByURL("https://ente.io/download"),
		},
		{
			Name:        "Ente Auth",
			Description: "Ente Authenticator (2FA)",
			Optional:    true,
			URL:         "https://github.com/ente-io/ente/releases?q=tag%3Aauth-v4",
			Check:       checkByBundleIDOrPath("io.ente.auth", "ente-auth"),
			Install:     installByURL("https://github.com/ente-io/ente/releases?q=tag%3Aauth-v4"),
		},
		{
			Name:        "Ente Locker",
			Description: "Ente Password Manager",
			Optional:    true,
			URL:         "https://ente.io/locker/",
			Check:       checkByBundleIDOrPath("io.ente.locker", "ente-locker"),
			Install:     installByURL("https://ente.io/locker/"),
		},
		{
			Name:        "GPGTools",
			Description: "OpenPGP Suite",
			Optional:    false,
			URL:         "https://gpgtools.org/",
			OS:          "darwin",
			Check:       checkByBundleID("org.gpgtools.gpgkeychain"),
			Install:     installByURL("https://gpgtools.org/"),
		},
		{
			Name:        "YubiKey Manager",
			Description: "YubiKey Configuration",
			Optional:    true,
			URL:         "https://www.yubico.com/support/download/yubikey-manager/",
			Check:       checkByBundleIDOrPath("com.yubico.ykman", "ykman-gui", "ykman"),
			Install:     installByURL("https://www.yubico.com/support/download/yubikey-manager/"),
		},
		{
			Name:        "SafeInCloud",
			Description: "Password manager",
			Optional:    true,
			URL:         "https://www.safe-in-cloud.com/en/download/",
			Check:       checkByBundleID("com.safeinscloud.SafeInCloud"),
			Install:     installByURL("https://www.safe-in-cloud.com/en/download/"),
		},
		{
			Name:        "NextDNS",
			Description: "DNS Security",
			Optional:    true,
			URL:         "https://nextdns.io/",
			Check:       checkByPath("nextdns"),
			Install:     installByURL("https://nextdns.io/"),
		},
	}
}

func getBrowserApps() []Software {
	return []Software{
		{
			Name:        "Google Chrome",
			Description: "Web Browser",
			Optional:    true,
			URL:         "https://www.google.com/chrome/",
			Check:       checkByBundleIDOrPath("com.google.Chrome", "google-chrome", "google-chrome-stable"),
			Install:     installByURL("https://www.google.com/chrome/"),
		},
		{
			Name:        "Brave Browser",
			Description: "Privacy Browser",
			Optional:    true,
			URL:         "https://brave.com/download/",
			Check:       checkByBundleIDOrPath("com.brave.Browser", "brave-browser", "brave"),
			Install:     installByURL("https://brave.com/download/"),
		},
	}
}

func getDeveloperAndTerminalApps() []Software {
	return []Software{
		{
			Name:        "iTerm",
			Description: "Terminal Emulator",
			Optional:    true,
			URL:         "https://iterm2.com/downloads.html",
			OS:          "darwin",
			Check:       checkByBundleID("com.googlecode.iterm2"),
			Install:     installByURL("https://iterm2.com/downloads.html"),
		},
		{
			Name:        "VNC Viewer",
			Description: "Remote Desktop",
			Optional:    true,
			URL:         "https://www.realvnc.com/en/connect/download/viewer/",
			Check:       checkByBundleIDOrPath("com.realvnc.vncviewer", "vncviewer"),
			Install:     installByURL("https://www.realvnc.com/en/connect/download/viewer/"),
		},
		{
			Name:        "Rancher Desktop",
			Description: "Container Management",
			Optional:    true,
			URL:         "https://rancherdesktop.io/",
			Check:       checkByBundleIDOrPath("io.rancherdesktop.app", "rancher-desktop"),
			Install:     installByURL("https://rancherdesktop.io/"),
		},
		{
			Name:        "Android Studio",
			Description: "Android app development IDE",
			Optional:    true,
			URL:         "https://developer.android.com/studio",
			Check:       checkAndroidStudio,
			Install:     installByURL("https://developer.android.com/studio"),
		},
	}
}

func getProductivityApps() []Software {
	return []Software{
		{
			Name:        "Alfred",
			Description: "Productivity App",
			Optional:    true,
			URL:         "https://www.alfredapp.com/",
			OS:          "darwin",
			Check:       checkByBundleID("com.runningwithcrayons.Alfred"),
			Install:     installByURL("https://www.alfredapp.com/"),
		},
		{
			Name:        "Notion",
			Description: "Notes & Collaboration",
			Optional:    true,
			URL:         "https://www.notion.so/desktop",
			Check:       checkByBundleIDOrPath("notion.id", "notion-app", "notion"),
			Install:     installByURL("https://www.notion.so/desktop"),
		},
		{
			Name:        "Obsidian",
			Description: "Markdown notes & knowledge base",
			Optional:    true,
			URL:         "https://obsidian.md/download",
			Check:       checkObsidian,
			Install:     installByURL("https://obsidian.md/download"),
		},
	}
}

func getMediaAndCommunicationApps() []Software {
	return []Software{
		{
			Name:        "LICEcap",
			Description: "GIF Recorder",
			Optional:    true,
			URL:         "https://www.cockos.com/licecap/",
			OS:          "darwin",
			Check:       checkByBundleID("com.cockos.LICEcap"),
			Install:     installByURL("https://www.cockos.com/licecap/"),
		},
		{
			Name:        "Signal",
			Description: "Secure Messaging",
			Optional:    true,
			URL:         "https://signal.org/download/",
			Check:       checkByBundleIDOrPath("org.whispersystems.signal-desktop", "signal-desktop", "signal"),
			Install:     installByURL("https://signal.org/download/"),
		},
		{
			Name:        "VLC",
			Description: "Video Player",
			Optional:    true,
			URL:         "https://www.videolan.org/vlc/",
			Check:       checkByBundleIDOrPath("org.videolan.vlc", "vlc"),
			Install:     installByURL("https://www.videolan.org/vlc/"),
		},
		{
			Name:        "OBS Studio",
			Description: "Screen Recorder",
			Optional:    true,
			URL:         "https://obsproject.com/",
			Check:       checkByBundleIDOrPath("com.obsproject.Studio", "obs", "obs-studio"),
			Install:     installByURL("https://obsproject.com/"),
		},
		{
			Name:        "HandBrake",
			Description: "Video Transcoder",
			Optional:    true,
			URL:         "https://handbrake.fr/downloads.php",
			Check:       checkByBundleIDOrPath("fr.handbrake.ghb", "ghb", "handbrake", "HandBrake"),
			Install:     installByURL("https://handbrake.fr/downloads.php"),
		},
		{
			Name:        "Spotify",
			Description: "Music Streaming",
			Optional:    true,
			URL:         "https://open.spotify.com/download",
			Check:       checkByBundleIDOrPath("com.spotify.client", "spotify"),
			Install:     installByURL("https://open.spotify.com/download"),
		},
	}
}

func getUtilityAndOtherApps() []Software {
	return []Software{
		{
			Name:        "Jabra Direct",
			Description: "Headset Software",
			Optional:    true,
			URL:         "https://www.jabra.com/software-and-services/jabra-direct",
			Check:       checkByBundleIDOrPath("com.jabra.directonline", "jabra-direct"),
			Install:     installByURL("https://www.jabra.com/software-and-services/jabra-direct"),
		},
		{
			Name:        "Antigravity IDE",
			Description: "AI-First IDE",
			Optional:    false,
			URL:         "https://antigravity.google/download",
			Check: func() bool {
				if checkByPath("agy-ide")() || checkByPath("antigravity-ide")() {
					return true
				}
				homeDir, err := userHomeDir()
				if err != nil {
					return false
				}
				_, err = stat(filepath.Join(homeDir, ".local", "opt", "antigravity-ide"))
				return err == nil
			},
			Install: func() error {
				return RunIdeUpdate(context.Background(), "", false, false)
			},
		},
	}
}

func getManualInstalls() []Software {
	var list []Software
	list = append(list, getSecurityAndPrivacyApps()...)
	list = append(list, getBrowserApps()...)
	list = append(list, getDeveloperAndTerminalApps()...)
	list = append(list, getProductivityApps()...)
	list = append(list, getMediaAndCommunicationApps()...)
	list = append(list, getUtilityAndOtherApps()...)
	return list
}

func getCLITools() []Software {
	return []Software{
		{
			Name:        "Bitwarden CLI",
			Description: "Password Manager CLI",
			Optional:    false,
			Check:       checkByPath("bw"),
			Install: func() error {
				if _, err := lookPath("brew"); err == nil {
					log.Info("Installing bitwarden-cli via Homebrew...")
					cmd := execCommand("brew", "install", "bitwarden-cli")
					cmd.Stdout = log.Writer()
					cmd.Stderr = log.ErrorWriter()
					return cmd.Run()
				}
				if _, err := lookPath("npm"); err == nil {
					log.Info("Installing bitwarden-cli via npm...")
					cmd := execCommand("npm", "install", "-g", "@bitwarden/cli")
					cmd.Stdout = log.Writer()
					cmd.Stderr = log.ErrorWriter()
					return cmd.Run()
				}
				log.Warn("Neither brew nor npm was found to install bitwarden-cli automatically.")
				log.Message("Please install bitwarden-cli manually or install Homebrew/npm.")
				return openURL("https://bitwarden.com/help/cli/")
			},
		},
	}
}

func getSoftwareList() []Software {
	core := getCoreSoftware()
	manual := getManualInstalls()
	cli := getCLITools()

	list := make([]Software, 0, len(core)+len(manual)+len(cli))
	list = append(list, core...)
	list = append(list, manual...)
	list = append(list, cli...)
	return list
}
