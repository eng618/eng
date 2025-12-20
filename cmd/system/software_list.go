package system

import (
	"runtime"

	"github.com/eng618/eng/utils/log"
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

func getSoftwareList() []Software {
	return []Software{
		// Critical / Core Items
		// VS Code is needed before Brew Bundle for extensions
		{
			Name:        "VS Code",
			Description: "Code Editor",
			Optional:    true,
			URL:         "https://code.visualstudio.com/",
			Check: func() bool {
				_, err := lookPath("code")
				return err == nil
			},
			Install: func() error { return openURL("https://code.visualstudio.com/") },
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

		// Manual Installs
		{
			Name:        "Google Chrome",
			Description: "Web Browser",
			Optional:    true,
			URL:         "https://www.google.com/chrome/",
			Check: func() bool {
				if runtime.GOOS == "darwin" {
					return execCommand("mdfind", "kMDItemCFBundleIdentifier == 'com.google.Chrome'").Run() == nil
				}
				_, err := lookPath("google-chrome")
				return err == nil
			},
			Install: func() error { return openURL("https://www.google.com/chrome/") },
		},
		{
			Name:        "Brave Browser",
			Description: "Privacy Browser",
			Optional:    true,
			URL:         "https://brave.com/download/",
			Check: func() bool {
				if runtime.GOOS == "darwin" {
					return execCommand("mdfind", "kMDItemCFBundleIdentifier == 'com.brave.Browser'").Run() == nil
				}
				_, err := lookPath("brave-browser")
				return err == nil
			},
			Install: func() error { return openURL("https://brave.com/download/") },
		},
		{
			Name:        "iTerm2",
			Description: "Terminal Emulator",
			Optional:    true,
			URL:         "https://iterm2.com/",
			OS:          "darwin",
			Check: func() bool {
				return execCommand("mdfind", "kMDItemCFBundleIdentifier == 'com.googlecode.iterm2'").Run() == nil
			},
			Install: func() error { return openURL("https://iterm2.com/") },
		},
		{
			Name:        "Alfred",
			Description: "Productivity App",
			Optional:    true,
			URL:         "https://www.alfredapp.com/",
			OS:          "darwin",
			Check: func() bool {
				return execCommand("mdfind", "kMDItemCFBundleIdentifier == 'com.runningwithcrayons.Alfred'").Run() == nil
			},
			Install: func() error { return openURL("https://www.alfredapp.com/") },
		},
		{
			Name:        "LICEcap",
			Description: "GIF Recorder",
			Optional:    true,
			URL:         "https://www.cockos.com/licecap/",
			OS:          "darwin",
			Check: func() bool {
				// Simple check, might not accept all paths
				return execCommand("mdfind", "kMDItemCFBundleIdentifier == 'com.cockos.LICEcap'").Run() == nil
			},
			Install: func() error { return openURL("https://www.cockos.com/licecap/") },
		},
		{
			Name:        "Signal",
			Description: "Secure Messaging",
			Optional:    true,
			URL:         "https://signal.org/download/",
			Check:       func() bool { return false },
			Install:     func() error { return openURL("https://signal.org/download/") },
		},
		{
			Name:        "VNC Viewer",
			Description: "Remote Desktop",
			Optional:    true,
			URL:         "https://www.realvnc.com/en/connect/download/viewer/",
			Check:       func() bool { return false },
			Install:     func() error { return openURL("https://www.realvnc.com/en/connect/download/viewer/") },
		},
		{
			Name:        "VLC",
			Description: "Video Player",
			Optional:    true,
			URL:         "https://www.videolan.org/vlc/",
			Check: func() bool {
				if runtime.GOOS == "darwin" {
					return execCommand("mdfind", "kMDItemCFBundleIdentifier == 'org.videolan.vlc'").Run() == nil
				}
				_, err := lookPath("vlc")
				return err == nil
			},
			Install: func() error { return openURL("https://www.videolan.org/vlc/") },
		},
		{
			Name:        "GPGTools",
			Description: "OpenPGP Suite",
			Optional:    false, // Seem important for dotfiles
			URL:         "https://gpgtools.org/",
			OS:          "darwin",
			Check: func() bool {
				return execCommand("mdfind", "kMDItemCFBundleIdentifier == 'org.gpgtools.gpgkeychain'").Run() == nil
			},
			Install: func() error { return openURL("https://gpgtools.org/") },
		},
		{
			Name:        "YubiKey Manager",
			Description: "YubiKey Configuration",
			Optional:    true,
			URL:         "https://www.yubico.com/support/download/yubikey-manager/",
			Check:       func() bool { return false },
			Install:     func() error { return openURL("https://www.yubico.com/support/download/yubikey-manager/") },
		},
		{
			Name:        "Rancher Desktop",
			Description: "Container Management",
			Optional:    true,
			URL:         "https://rancherdesktop.io/",
			Check:       func() bool { return false },
			Install:     func() error { return openURL("https://rancherdesktop.io/") },
		},
		{
			Name:        "Jabra Direct",
			Description: "Headset Software",
			Optional:    true,
			URL:         "https://www.jabra.com/software-and-services/jabra-direct",
			Check:       func() bool { return false },
			Install:     func() error { return openURL("https://www.jabra.com/software-and-services/jabra-direct") },
		},
		{
			Name:        "OBS Studio",
			Description: "Screen Recorder",
			Optional:    true,
			URL:         "https://obsproject.com/",
			Check:       func() bool { return false },
			Install:     func() error { return openURL("https://obsproject.com/") },
		},
		{
			Name:        "HandBrake",
			Description: "Video Transcoder",
			Optional:    true,
			URL:         "https://handbrake.fr/downloads.php",
			Check:       func() bool { return false },
			Install:     func() error { return openURL("https://handbrake.fr/downloads.php") },
		},
		{
			Name:        "Notion",
			Description: "Notes & Collaboration",
			Optional:    true,
			URL:         "https://www.notion.so/desktop",
			Check:       func() bool { return false },
			Install:     func() error { return openURL("https://www.notion.so/desktop") },
		},
		{
			Name:        "NextDNS",
			Description: "DNS Security",
			Optional:    true,
			URL:         "https://nextdns.io/",
			Check:       func() bool { return false },
			Install:     func() error { return openURL("https://nextdns.io/") },
		},
		{
			Name:        "Antigravity",
			Description: "Antigravity Tool",
			Optional:    false,
			URL:         "https://antigravity.google/download",
			Check:       func() bool { return false },
			Install:     func() error { return openURL("https://antigravity.google/download") },
		},
		{
			Name:        "Spotify",
			Description: "Music Streaming",
			Optional:    true,
			URL:         "https://open.spotify.com/download",
			Check: func() bool {
				if runtime.GOOS == "darwin" {
					return execCommand("mdfind", "kMDItemCFBundleIdentifier == 'com.spotify.client'").Run() == nil
				}
				_, err := lookPath("spotify")
				return err == nil
			},
			Install: func() error { return openURL("https://open.spotify.com/download") },
		},
	}
}
