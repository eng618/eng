package sysinfo

import (
	"errors"
	"testing"
)

func TestDetect_MacOS(t *testing.T) {
	origGOOS := RuntimeGOOS
	defer func() { RuntimeGOOS = origGOOS }()

	RuntimeGOOS = "darwin"
	info := Detect()

	if !info.IsMacOS() {
		t.Error("Expected IsMacOS() to be true")
	}
	if info.IsLinux() {
		t.Error("Expected IsLinux() to be false")
	}
	if info.IsFedora() {
		t.Error("Expected IsFedora() to be false")
	}
	if info.ID != "macos" {
		t.Errorf("Expected ID 'macos', got %q", info.ID)
	}
}

func TestDetect_Fedora(t *testing.T) {
	origGOOS := RuntimeGOOS
	origReadFile := ReadFile
	origLookPath := LookPath
	defer func() {
		RuntimeGOOS = origGOOS
		ReadFile = origReadFile
		LookPath = origLookPath
	}()

	RuntimeGOOS = "linux"
	fedoraOSRelease := `NAME="Fedora Linux"
VERSION="41 (Workstation Edition)"
ID=fedora
VERSION_ID=41
PLATFORM_ID="platform:f41"
PRETTY_NAME="Fedora Linux 41 (Workstation Edition)"
ANSI_COLOR="0;38;2;60;110;180"
LOGO=fedora-logo-icon
CPE_NAME="cpe:/o:fedoraproject:fedora:41"
HOME_URL="https://fedoraproject.org/"
`
	ReadFile = func(name string) ([]byte, error) {
		return []byte(fedoraOSRelease), nil
	}
	LookPath = func(file string) (string, error) {
		if file == "dnf" {
			return "/usr/bin/dnf", nil
		}
		return "", errors.New("not found")
	}

	info := Detect()

	if !info.IsLinux() {
		t.Error("Expected IsLinux() to be true")
	}
	if !info.IsFedora() {
		t.Error("Expected IsFedora() to be true")
	}
	if info.IsDebianUbuntu() {
		t.Error("Expected IsDebianUbuntu() to be false")
	}
	if info.IsMacOS() {
		t.Error("Expected IsMacOS() to be false")
	}
	if info.ID != "fedora" {
		t.Errorf("Expected ID 'fedora', got %q", info.ID)
	}
	if info.VersionID != "41" {
		t.Errorf("Expected VersionID '41', got %q", info.VersionID)
	}
}

func TestDetect_RHEL(t *testing.T) {
	origGOOS := RuntimeGOOS
	origReadFile := ReadFile
	origLookPath := LookPath
	defer func() {
		RuntimeGOOS = origGOOS
		ReadFile = origReadFile
		LookPath = origLookPath
	}()

	RuntimeGOOS = "linux"
	rhelRelease := `NAME="Red Hat Enterprise Linux"
VERSION="9.4 (Plow)"
ID="rhel"
ID_LIKE="fedora"
VERSION_ID="9.4"
PRETTY_NAME="Red Hat Enterprise Linux 9.4 (Plow)"
`
	ReadFile = func(name string) ([]byte, error) {
		return []byte(rhelRelease), nil
	}
	LookPath = func(file string) (string, error) {
		if file == "dnf" {
			return "/usr/bin/dnf", nil
		}
		return "", errors.New("not found")
	}

	info := Detect()

	if !info.IsFedora() {
		t.Error("Expected IsFedora() to be true for RHEL (ID_LIKE fedora)")
	}
	if info.IsDebianUbuntu() {
		t.Error("Expected IsDebianUbuntu() to be false for RHEL")
	}
}

func TestDetect_Ubuntu(t *testing.T) {
	origGOOS := RuntimeGOOS
	origReadFile := ReadFile
	origLookPath := LookPath
	defer func() {
		RuntimeGOOS = origGOOS
		ReadFile = origReadFile
		LookPath = origLookPath
	}()

	RuntimeGOOS = "linux"
	ubuntuRelease := `NAME="Ubuntu"
VERSION="24.04 LTS (Noble Numbat)"
ID=ubuntu
ID_LIKE=debian
PRETTY_NAME="Ubuntu 24.04 LTS"
VERSION_ID="24.04"
`
	ReadFile = func(name string) ([]byte, error) {
		return []byte(ubuntuRelease), nil
	}
	LookPath = func(file string) (string, error) {
		if file == "apt-get" {
			return "/usr/bin/apt-get", nil
		}
		return "", errors.New("not found")
	}

	info := Detect()

	if !info.IsDebianUbuntu() {
		t.Error("Expected IsDebianUbuntu() to be true")
	}
	if info.IsFedora() {
		t.Error("Expected IsFedora() to be false")
	}
}

func TestDetect_RaspberryPi(t *testing.T) {
	origGOOS := RuntimeGOOS
	origReadFile := ReadFile
	origLookPath := LookPath
	defer func() {
		RuntimeGOOS = origGOOS
		ReadFile = origReadFile
		LookPath = origLookPath
	}()

	RuntimeGOOS = "linux"
	raspbianRelease := `PRETTY_NAME="Raspbian GNU/Linux 12 (bookworm)"
NAME="Raspbian GNU/Linux"
VERSION_ID="12"
VERSION="12 (bookworm)"
VERSION_CODENAME=bookworm
ID=raspbian
ID_LIKE=debian
`
	ReadFile = func(name string) ([]byte, error) {
		return []byte(raspbianRelease), nil
	}
	LookPath = func(file string) (string, error) {
		return "", errors.New("not found")
	}

	info := Detect()

	if !info.IsRaspberryPi() {
		t.Error("Expected IsRaspberryPi() to be true")
	}
	if !info.IsDebianUbuntu() {
		t.Error("Expected IsDebianUbuntu() to be true due to ID_LIKE=debian")
	}
}

func TestDetect_PackageManagers(t *testing.T) {
	origLookPath := LookPath
	defer func() { LookPath = origLookPath }()

	available := map[string]bool{
		"brew":    true,
		"flatpak": true,
		"docker":  true,
		"dnf":     true,
		"apt-get": false,
	}

	LookPath = func(file string) (string, error) {
		if available[file] {
			return "/usr/bin/" + file, nil
		}
		return "", errors.New("not found")
	}

	if !HasBrew() {
		t.Error("Expected HasBrew() to be true")
	}
	if !HasFlatpak() {
		t.Error("Expected HasFlatpak() to be true")
	}
	if !HasDocker() {
		t.Error("Expected HasDocker() to be true")
	}
	if !HasDNF() {
		t.Error("Expected HasDNF() to be true")
	}
	if HasApt() {
		t.Error("Expected HasApt() to be false")
	}
}
