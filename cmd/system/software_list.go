package system

import (
	"context"
	"os"
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

func checkByBundleIDOrPath(bundleID, linuxPath string) func() bool {
	return func() bool {
		if runtime.GOOS == "darwin" {
			return execCommand("mdfind", "kMDItemCFBundleIdentifier == '"+bundleID+"'").Run() == nil
		}
		_, err := lookPath(linuxPath)
		return err == nil
	}
}

func checkObsidian() bool {
	if runtime.GOOS == "darwin" {
		return execCommand(
			"mdfind",
			"kMDItemCFBundleIdentifier == 'com.obsidian.md' || kMDItemCFBundleIdentifier == 'md.obsidian'",
		).Run() == nil
	}
	_, err := lookPath("obsidian")
	return err == nil
}

func checkAndroidStudio() bool {
	if runtime.GOOS == "darwin" {
		_, err := os.Stat("/Applications/Android Studio.app")
		return err == nil
	}
	if runtime.GOOS == "linux" {
		_, err := lookPath("studio")
		if err == nil {
			return true
		}
		_, err = lookPath("studio.sh")
		return err == nil
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
			Check:       checkByPath("code"),
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
				// Checked in setup.go usually, but good to have here
				cmd := execCommand("sh", "-c", "[ -d \"$HOME/.oh-my-zsh\" ]")
				return cmd.Run() == nil
			},
			Install: func() error {
				// This is handled by setupOhMyZsh in setup.go, but we could unify.
				// For now, let's keep it consistent with the list.
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
			Check:       checkFalse,
			Install:     installByURL("https://ente.io/download"),
		},
		{
			Name:        "Ente Auth",
			Description: "Ente Authenticator (2FA)",
			Optional:    true,
			URL:         "https://github.com/ente-io/ente/releases?q=tag%3Aauth-v4",
			Check:       checkFalse,
			Install:     installByURL("https://github.com/ente-io/ente/releases?q=tag%3Aauth-v4"),
		},
		{
			Name:        "Ente Locker",
			Description: "Ente Password Manager",
			Optional:    true,
			URL:         "https://ente.io/locker/",
			Check:       checkFalse,
			Install:     installByURL("https://ente.io/locker/"),
		},
		{
			Name:        "GPGTools",
			Description: "OpenPGP Suite",
			Optional:    false, // Seem important for dotfiles
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
			Check:       checkFalse,
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
			Check:       checkFalse,
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
			Check:       checkByBundleIDOrPath("com.google.Chrome", "google-chrome"),
			Install:     installByURL("https://www.google.com/chrome/"),
		},
		{
			Name:        "Brave Browser",
			Description: "Privacy Browser",
			Optional:    true,
			URL:         "https://brave.com/download/",
			Check:       checkByBundleIDOrPath("com.brave.Browser", "brave-browser"),
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
			Check:       checkFalse,
			Install:     installByURL("https://www.realvnc.com/en/connect/download/viewer/"),
		},
		{
			Name:        "Rancher Desktop",
			Description: "Container Management",
			Optional:    true,
			URL:         "https://rancherdesktop.io/",
			Check:       checkFalse,
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
			Check:       checkFalse,
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
			Check:       checkFalse,
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
			Check:       checkFalse,
			Install:     installByURL("https://obsproject.com/"),
		},
		{
			Name:        "HandBrake",
			Description: "Video Transcoder",
			Optional:    true,
			URL:         "https://handbrake.fr/downloads.php",
			Check:       checkFalse,
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
			Check:       checkFalse,
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
				log.Info("Installing bitwarden-cli via brew...")
				cmd := execCommand("brew", "install", "bitwarden-cli")
				cmd.Stdout = log.Writer()
				cmd.Stderr = log.ErrorWriter()
				return cmd.Run()
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
