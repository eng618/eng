package sysinfo

import (
	"bufio"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// DistroInfo contains parsed distribution information from /etc/os-release.
type DistroInfo struct {
	ID         string
	IDLike     []string
	VersionID  string
	Name       string
	PrettyName string
	RawOS      string // runtime.GOOS
}

var (
	// LookPath is mockable for testing.
	LookPath = exec.LookPath
	// ReadFile is mockable for testing.
	ReadFile = os.ReadFile
	// RuntimeGOOS is mockable for testing.
	RuntimeGOOS = runtime.GOOS
	// OSReleasePaths are the standard paths checked for distro identification.
	OSReleasePaths = []string{"/etc/os-release", "/usr/lib/os-release"}
)

// Detect probes the operating system and Linux distribution metadata.
func Detect() DistroInfo {
	info := DistroInfo{
		RawOS: RuntimeGOOS,
	}

	if info.RawOS == "darwin" {
		info.ID = "macos"
		info.Name = "macOS"
		info.PrettyName = "macOS"
		return info
	}

	if info.RawOS != "linux" {
		info.ID = info.RawOS
		info.Name = info.RawOS
		info.PrettyName = info.RawOS
		return info
	}

	// Read /etc/os-release
	var data []byte
	for _, p := range OSReleasePaths {
		content, err := ReadFile(p)
		if err == nil {
			data = content
			break
		}
	}

	if len(data) > 0 {
		parseOSRelease(string(data), &info)
	}

	if info.ID == "" {
		// Fallbacks based on package managers present
		if _, err := LookPath("dnf"); err == nil {
			info.ID = "fedora"
			info.Name = "Fedora Linux"
		} else if _, err := LookPath("apt-get"); err == nil {
			info.ID = "debian"
			info.Name = "Debian / Ubuntu"
		} else {
			info.ID = "linux"
			info.Name = "Linux"
		}
	}

	if info.PrettyName == "" {
		info.PrettyName = info.Name
	}
	if info.PrettyName == "" {
		info.PrettyName = info.ID
	}

	return info
}

func parseOSRelease(content string, info *DistroInfo) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, `"'`)

		switch key {
		case "ID":
			info.ID = strings.ToLower(val)
		case "ID_LIKE":
			fields := strings.Fields(strings.ToLower(val))
			info.IDLike = append(info.IDLike, fields...)
		case "VERSION_ID":
			info.VersionID = val
		case "NAME":
			info.Name = val
		case "PRETTY_NAME":
			info.PrettyName = val
		}
	}
}

// IsFedora returns true if running on Fedora or RHEL/CentOS family.
func (d DistroInfo) IsFedora() bool {
	if d.RawOS != "linux" {
		return false
	}
	if d.ID == "fedora" || d.ID == "rhel" || d.ID == "centos" || d.ID == "rocky" || d.ID == "alma" {
		return true
	}
	for _, like := range d.IDLike {
		if like == "fedora" || like == "rhel" || like == "centos" {
			return true
		}
	}
	// Fallback to DNF presence
	if _, err := LookPath("dnf"); err == nil {
		return true
	}
	if _, err := LookPath("dnf5"); err == nil {
		return true
	}
	return false
}

// IsDebianUbuntu returns true if running on Ubuntu, Debian, or derivative.
func (d DistroInfo) IsDebianUbuntu() bool {
	if d.RawOS != "linux" {
		return false
	}
	if d.ID == "ubuntu" || d.ID == "debian" || d.ID == "pop" || d.ID == "linuxmint" || d.ID == "elementary" {
		return true
	}
	for _, like := range d.IDLike {
		if like == "ubuntu" || like == "debian" {
			return true
		}
	}
	if _, err := LookPath("apt-get"); err == nil {
		return true
	}
	return false
}

// IsRaspberryPi returns true if running on Raspbian / Raspberry Pi OS.
func (d DistroInfo) IsRaspberryPi() bool {
	if d.RawOS != "linux" {
		return false
	}
	if d.ID == "raspbian" {
		return true
	}
	for _, like := range d.IDLike {
		if like == "raspbian" {
			return true
		}
	}
	return false
}

// IsMacOS returns true if running on Darwin/macOS.
func (d DistroInfo) IsMacOS() bool {
	return d.RawOS == "darwin"
}

// IsLinux returns true if running on Linux.
func (d DistroInfo) IsLinux() bool {
	return d.RawOS == "linux"
}

// HasBrew returns true if the 'brew' executable is on PATH.
func HasBrew() bool {
	_, err := LookPath("brew")
	return err == nil
}

// HasFlatpak returns true if the 'flatpak' executable is on PATH.
func HasFlatpak() bool {
	_, err := LookPath("flatpak")
	return err == nil
}

// HasDocker returns true if the 'docker' executable is on PATH.
func HasDocker() bool {
	_, err := LookPath("docker")
	return err == nil
}

// HasDNF returns true if 'dnf' or 'dnf5' is on PATH.
func HasDNF() bool {
	if _, err := LookPath("dnf"); err == nil {
		return true
	}
	if _, err := LookPath("dnf5"); err == nil {
		return true
	}
	return false
}

// HasApt returns true if 'apt-get' or 'apt' is on PATH.
func HasApt() bool {
	if _, err := LookPath("apt-get"); err == nil {
		return true
	}
	if _, err := LookPath("apt"); err == nil {
		return true
	}
	return false
}
