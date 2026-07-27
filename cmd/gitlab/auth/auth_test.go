package auth

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/bitwarden"
	"github.com/eng618/eng/internal/log"
)

func setupTestViper(t *testing.T) string {
	t.Helper()
	viper.Reset()
	viper.SetConfigType("yaml")
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".eng.yaml")
	viper.SetConfigFile(configPath)
	_ = os.WriteFile(configPath, []byte("{}"), 0o644)
	return configPath
}

func resetFlags() {
	setTokenItem = ""
	setToken = ""
	setTokenStdin = false
	setHost = ""
	setProject = ""
	setNotes = ""
	docHostOpt = ""
	docProjectOpt = ""
	docQuiet = false
}

// ============================================================================
// T - Tests (Unit and Table-driven Tests)
// ============================================================================

func TestAuthCmd_Subcommands(t *testing.T) {
	resetFlags()
	subcommands := AuthCmd.Commands()
	expected := map[string]bool{
		"show":   false,
		"doctor": false,
		"set":    false,
	}

	for _, cmd := range subcommands {
		if _, exists := expected[cmd.Use]; exists {
			expected[cmd.Use] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("expected subcommand %s under AuthCmd", name)
		}
	}
}

func TestShowCmd(t *testing.T) {
	setupTestViper(t)
	resetFlags()

	// Backup and override bitwarden execCommand
	origBwExec := bitwarden.GetExecCommandForTest()
	defer bitwarden.SetExecCommandForTest(origBwExec)

	bitwarden.SetExecCommandForTest(func(name string, arg ...string) *exec.Cmd {
		if len(arg) > 0 && arg[0] == "status" {
			return exec.Command("echo", `{"status":"unlocked"}`)
		}
		if len(arg) > 2 && arg[0] == "get" && arg[1] == "item" {
			return exec.Command("echo", `{"id":"item-id","name":"gitlab-token","login":{"password":"glpat-mocktoken"}}`)
		}
		return exec.Command("true")
	})

	tests := []struct {
		name           string
		setupViper     func()
		setupEnv       func()
		expectContains []string
	}{
		{
			name: "token from env",
			setupViper: func() {
				viper.Set("gitlab.host", "gitlab.example.com")
				viper.Set("gitlab.project", "group/project")
			},
			setupEnv: func() {
				_ = os.Setenv("GITLAB_TOKEN", "glpat-envtoken")
			},
			expectContains: []string{"gitlab.example.com", "group/project", "env:GITLAB_TOKEN"},
		},
		{
			name: "token from bitwarden",
			setupViper: func() {
				viper.Set("gitlab.host", "gitlab.com")
				viper.Set("gitlab.project", "my/repo")
				viper.Set("gitlab.tokenItem", "gitlab-token")
			},
			setupEnv: func() {
				_ = os.Unsetenv("GITLAB_TOKEN")
				_ = os.Setenv("BW_SESSION", "mocksession")
			},
			expectContains: []string{"gitlab.com", "my/repo", "bitwarden:gitlab-token"},
		},
		{
			name: "token from config",
			setupViper: func() {
				viper.Set("gitlab.token", "glpat-configtoken")
				viper.Set("gitlab.host", "gitlab.com")
				viper.Set("gitlab.project", "my/repo")
			},
			setupEnv: func() {
				_ = os.Unsetenv("GITLAB_TOKEN")
				viper.Set("gitlab.tokenItem", "")
			},
			expectContains: []string{"config:gitlab.token"},
		},
		{
			name: "no token",
			setupViper: func() {
				viper.Set("gitlab.token", "")
				viper.Set("gitlab.tokenItem", "")
				viper.Set("gitlab.host", "gitlab.com")
				viper.Set("gitlab.project", "my/repo")
			},
			setupEnv: func() {
				_ = os.Unsetenv("GITLAB_TOKEN")
			},
			expectContains: []string{"token:   none"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupViper()
			tt.setupEnv()

			var outBuf bytes.Buffer
			log.SetWriters(&outBuf, &outBuf)
			defer log.ResetWriters()

			cmd := showCmd
			err := cmd.RunE(cmd, []string{})
			if err != nil {
				t.Fatalf("showCmd failed: %v", err)
			}

			output := outBuf.String()
			for _, sub := range tt.expectContains {
				if !strings.Contains(output, sub) {
					t.Errorf("expected output to contain %q, output: %q", sub, output)
				}
			}
		})
	}
}

func TestDoctorCmd_Success(t *testing.T) {
	setupTestViper(t)
	resetFlags()

	// Backup and override lookPath and execCommand
	oldLookPath := lookPath
	oldExecCommand := execCommand
	defer func() {
		lookPath = oldLookPath
		execCommand = oldExecCommand
	}()

	lookPath = func(file string) (string, error) {
		if file == "glab" {
			return "/usr/local/bin/glab", nil
		}
		return "", errors.New("not found")
	}

	execCommand = func(name string, arg ...string) *exec.Cmd {
		if name == "glab" && len(arg) >= 2 && arg[0] == "api" {
			if arg[1] == "user" {
				return exec.Command("echo", `{"username":"doctoruser","name":"Doctor User"}`)
			}
			if strings.HasPrefix(arg[1], "projects/") {
				return exec.Command("echo", `{"path_with_namespace":"mygroup/myproject"}`)
			}
		}
		return exec.Command("true")
	}

	viper.Set("gitlab.host", "gitlab.com")
	viper.Set("gitlab.project", "mygroup/myproject")
	_ = os.Setenv("GITLAB_TOKEN", "mock-token")

	var outBuf bytes.Buffer
	log.SetWriters(&outBuf, &outBuf)
	defer log.ResetWriters()

	cmd := doctorCmd
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("doctorCmd failed: %v", err)
	}

	output := outBuf.String()
	if !strings.Contains(output, "Token valid for user: doctoruser") {
		t.Errorf("expected validation success message, got %q", output)
	}
	if !strings.Contains(output, "Project access OK: mygroup/myproject") {
		t.Errorf("expected project access OK message, got %q", output)
	}
}

func TestDoctorCmd_GlabMissing(t *testing.T) {
	resetFlags()

	oldLookPath := lookPath
	defer func() { lookPath = oldLookPath }()

	lookPath = func(file string) (string, error) {
		return "", errors.New("glab not found")
	}

	cmd := doctorCmd
	err := cmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("expected doctorCmd to fail when glab is missing, got nil")
	}
	if !strings.Contains(err.Error(), "glab CLI not found in PATH") {
		t.Errorf("expected error message to mention glab CLI, got: %v", err)
	}
}

func TestSetCmd_SaveToken(t *testing.T) {
	setupTestViper(t)
	resetFlags()

	// Backup and override bitwarden execCommand
	origBwExec := bitwarden.GetExecCommandForTest()
	defer bitwarden.SetExecCommandForTest(origBwExec)

	bitwarden.SetExecCommandForTest(func(name string, arg ...string) *exec.Cmd {
		if name == "bw" {
			if arg[0] == "status" {
				return exec.Command("echo", `{"status":"unlocked"}`)
			}
			if arg[0] == "list" {
				return exec.Command("echo", `[]`) // No existing item
			}
			if arg[0] == "encode" {
				return exec.Command("echo", `{"id":"new-item"}`)
			}
			if arg[0] == "create" {
				return exec.Command("echo", `{"id":"created-id"}`)
			}
		}
		return exec.Command("true")
	})

	setToken = "glpat-tokensaving123"
	setTokenItem = "gitlab-token-key"
	setHost = "gitlab.enterprise.com"
	setProject = "enterprise/repo"

	cmd := setCmd
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("setCmd run failed: %v", err)
	}

	// Verify viper configuration was updated
	if viper.GetString("gitlab.tokenItem") != "gitlab-token-key" {
		t.Errorf("expected tokenItem gitlab-token-key, got %s", viper.GetString("gitlab.tokenItem"))
	}
	if viper.GetString("gitlab.host") != "gitlab.enterprise.com" {
		t.Errorf("expected host gitlab.enterprise.com, got %s", viper.GetString("gitlab.host"))
	}
	if viper.GetString("gitlab.project") != "enterprise/repo" {
		t.Errorf("expected project enterprise/repo, got %s", viper.GetString("gitlab.project"))
	}
}

// ============================================================================
// E - Examples (Executable Documentation)
// ============================================================================

func ExampleAuthCmd() {
	fmt.Println("Use:", AuthCmd.Use)
	fmt.Println("Short:", AuthCmd.Short)
	// Output:
	// Use: auth
	// Short: Manage GitLab authentication for eng
}

// ============================================================================
// B - Benchmarks (Performance Profiling)
// ============================================================================

func BenchmarkAuthCmd_Init(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AuthCmd.Commands()
	}
}
