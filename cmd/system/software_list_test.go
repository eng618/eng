package system

import (
	"errors"
	"os/exec"
	"testing"
)

func TestGetSoftwareList_Checks(t *testing.T) {
	origLookPath := lookPath
	origExec := execCommand
	defer func() {
		lookPath = origLookPath
		execCommand = origExec
	}()

	list := getSoftwareList()
	if len(list) == 0 {
		t.Fatal("Software list is empty")
	}

	// Mock VS Code check
	lookPath = func(path string) (string, error) {
		if path == "code" {
			return "/usr/local/bin/code", nil
		}
		return "", exec.ErrNotFound
	}

	for _, sw := range list {
		if sw.Name == "VS Code" {
			if !sw.Check() {
				t.Error("VS Code check should return true when executable is found")
			}
		}
	}
}

func TestCheckByBundleIDOrPath_Linux(t *testing.T) {
	origLookPath := lookPath
	defer func() { lookPath = origLookPath }()

	lookPath = func(path string) (string, error) {
		if path == "signal-desktop" {
			return "/usr/bin/signal-desktop", nil
		}
		return "", errors.New("not found")
	}

	checker := checkByBundleIDOrPath("org.whispersystems.signal-desktop", "signal-desktop", "signal")
	if !checker() {
		t.Error("expected checkByBundleIDOrPath to return true when linux binary exists")
	}
}

func TestBitwardenCLI_Install_Brew(t *testing.T) {
	origLookPath := lookPath
	origExec := execCommand
	defer func() {
		lookPath = origLookPath
		execCommand = origExec
	}()

	lookPath = func(path string) (string, error) {
		if path == "brew" {
			return "/usr/bin/brew", nil
		}
		return "", errors.New("not found")
	}

	calledBrew := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "brew" && len(args) >= 2 && args[0] == "install" && args[1] == "bitwarden-cli" {
			calledBrew = true
		}
		return exec.Command("echo", "success")
	}

	cliTools := getCLITools()
	var bwTool *Software
	for i := range cliTools {
		if cliTools[i].Name == "Bitwarden CLI" {
			bwTool = &cliTools[i]
			break
		}
	}

	if bwTool == nil {
		t.Fatal("Bitwarden CLI not found in getCLITools()")
	}

	if err := bwTool.Install(); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if !calledBrew {
		t.Error("expected bitwarden-cli to install via brew when brew is available")
	}
}

func TestBitwardenCLI_Install_NPM(t *testing.T) {
	origLookPath := lookPath
	origExec := execCommand
	defer func() {
		lookPath = origLookPath
		execCommand = origExec
	}()

	lookPath = func(path string) (string, error) {
		if path == "npm" {
			return "/usr/bin/npm", nil
		}
		return "", errors.New("not found")
	}

	calledNPM := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "npm" && len(args) >= 3 && args[0] == "install" && args[1] == "-g" && args[2] == "@bitwarden/cli" {
			calledNPM = true
		}
		return exec.Command("echo", "success")
	}

	cliTools := getCLITools()
	var bwTool *Software
	for i := range cliTools {
		if cliTools[i].Name == "Bitwarden CLI" {
			bwTool = &cliTools[i]
			break
		}
	}

	if bwTool == nil {
		t.Fatal("Bitwarden CLI not found in getCLITools()")
	}

	if err := bwTool.Install(); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if !calledNPM {
		t.Error("expected bitwarden-cli to install via npm when brew is absent and npm is available")
	}
}

func TestOpenURL(t *testing.T) {
	origExec := execCommand
	defer func() { execCommand = origExec }()

	called := false
	execCommand = func(name string, args ...string) *exec.Cmd {
		if name == "open" || name == "xdg-open" || name == "cmd" {
			called = true
		}
		return exec.Command("echo", "success")
	}

	_ = openURL("https://example.com")

	if !called {
		t.Error("openURL did not call any system command")
	}
}

func TestCategorizedSoftwareLists(t *testing.T) {
	security := getSecurityAndPrivacyApps()
	if len(security) == 0 {
		t.Error("getSecurityAndPrivacyApps returned empty list")
	}

	foundSurfshark := false
	for _, s := range security {
		if s.Name == "Surfshark" {
			foundSurfshark = true
			if s.URL == "" {
				t.Error("Surfshark missing URL")
			}
			break
		}
	}
	if !foundSurfshark {
		t.Error("Surfshark not found in getSecurityAndPrivacyApps()")
	}

	browsers := getBrowserApps()
	if len(browsers) == 0 {
		t.Error("getBrowserApps returned empty list")
	}

	dev := getDeveloperAndTerminalApps()
	if len(dev) == 0 {
		t.Error("getDeveloperAndTerminalApps returned empty list")
	}

	prod := getProductivityApps()
	if len(prod) == 0 {
		t.Error("getProductivityApps returned empty list")
	}

	media := getMediaAndCommunicationApps()
	if len(media) == 0 {
		t.Error("getMediaAndCommunicationApps returned empty list")
	}

	utils := getUtilityAndOtherApps()
	if len(utils) == 0 {
		t.Error("getUtilityAndOtherApps returned empty list")
	}
}
