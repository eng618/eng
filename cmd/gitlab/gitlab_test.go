package gitlab

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"

	gitlabcfg "github.com/eng618/eng/internal/config/gitlab"
	"github.com/eng618/eng/internal/ui"
)

func resetFlags() {
	initOutputPath = ""
	initForce = false
	initYes = false
	rulesPath = ""
	projectOpt = ""
	hostOpt = ""
	dryRun = false
	tokenItemOpt = ""
}

// ============================================================================
// T - Tests (Unit and Table-driven Tests)
// ============================================================================

func TestGitLabCmd_Subcommands(t *testing.T) {
	resetFlags()
	subcommands := GitLabCmd.Commands()
	expected := map[string]bool{
		"auth":     false,
		"mr-rules": false,
	}

	for _, cmd := range subcommands {
		if _, exists := expected[cmd.Use]; exists {
			expected[cmd.Use] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("expected subcommand %s under GitLabCmd", name)
		}
	}

	mrSubcommands := mrRulesCmd.Commands()
	expectedMR := map[string]bool{
		"init":  false,
		"apply": false,
	}

	for _, cmd := range mrSubcommands {
		if _, exists := expectedMR[cmd.Use]; exists {
			expectedMR[cmd.Use] = true
		}
	}

	for name, found := range expectedMR {
		if !found {
			t.Errorf("expected subcommand %s under mrRulesCmd", name)
		}
	}
}

func TestMrRulesInitCmd_Yes(t *testing.T) {
	resetFlags()
	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "rules.json")

	// Set CLI flags directly (Cobra binds them to package variables)
	initOutputPath = outPath
	initYes = true

	cmd := mrRulesInitCmd
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error during init --yes: %v", err)
	}

	// Verify file was written
	b, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read generated rules: %v", err)
	}

	var rules gitlabcfg.MRRules
	if err := json.Unmarshal(b, &rules); err != nil {
		t.Fatalf("failed to parse output rules json: %v", err)
	}

	if !rules.DeleteSourceBranch || !rules.RequireSquash {
		t.Errorf("expected default rules to be set to true, got %+v", rules)
	}
}

func TestMrRulesInitCmd_Interactive(t *testing.T) {
	resetFlags()
	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "rules.json")

	initOutputPath = outPath
	initYes = false

	// Mock stdin reader
	oldReader := reader
	defer func() { reader = oldReader }()
	// Feed inputs: Schema version = "2"
	reader = bufio.NewReader(strings.NewReader("2\n"))

	// Mock ui.Select prompts
	oldSelect := ui.Select
	defer func() { ui.Select = oldSelect }()
	ui.Select = func(prompt string, options []string, def string) (string, error) {
		// Mock responses
		switch {
		case strings.Contains(prompt, "Merge method"):
			return "merge_commit", nil
		case strings.Contains(prompt, "Delete source branch"):
			return "No", nil
		case strings.Contains(prompt, "Require squash"):
			return "Yes", nil
		case strings.Contains(prompt, "Pipelines must succeed"):
			return "No", nil
		case strings.Contains(prompt, "Treat skipped"):
			return "Yes", nil
		case strings.Contains(prompt, "All threads"):
			return "No", nil
		default:
			return def, nil
		}
	}

	cmd := mrRulesInitCmd
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error during interactive init: %v", err)
	}

	// Verify file was written and values correctly matched mock inputs
	b, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read generated rules: %v", err)
	}

	var rules gitlabcfg.MRRules
	if err := json.Unmarshal(b, &rules); err != nil {
		t.Fatalf("failed to parse rules json: %v", err)
	}

	if rules.SchemaVersion != "2" {
		t.Errorf("expected SchemaVersion '2', got %q", rules.SchemaVersion)
	}
	if rules.MergeMethod != "merge_commit" {
		t.Errorf("expected MergeMethod 'merge_commit', got %q", rules.MergeMethod)
	}
	if rules.DeleteSourceBranch {
		t.Error("expected DeleteSourceBranch to be false")
	}
	if !rules.RequireSquash {
		t.Error("expected RequireSquash to be true")
	}
	if rules.PipelinesMustSucceed {
		t.Error("expected PipelinesMustSucceed to be false")
	}
	if !rules.AllowSkippedAsSuccess {
		t.Error("expected AllowSkippedAsSuccess to be true")
	}
	if rules.AllThreadsMustResolve {
		t.Error("expected AllThreadsMustResolve to be false")
	}
}

func TestMrRulesInitCmd_ExistsConflict(t *testing.T) {
	resetFlags()
	tempDir := t.TempDir()
	outPath := filepath.Join(tempDir, "rules.json")

	// Pre-create the file
	err := os.WriteFile(outPath, []byte("{}"), 0o644)
	if err != nil {
		t.Fatalf("failed to write initial file: %v", err)
	}

	initOutputPath = outPath
	initYes = false
	initForce = false

	// Mock ui.Select to return "No" on overwrite prompt
	oldSelect := ui.Select
	defer func() { ui.Select = oldSelect }()
	ui.Select = func(prompt string, options []string, def string) (string, error) {
		if strings.Contains(prompt, "Overwrite") {
			return "No", nil
		}
		return "Yes", nil
	}

	cmd := mrRulesInitCmd
	err = cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file is still empty (not overwritten)
	b, _ := os.ReadFile(outPath)
	if string(b) != "{}" {
		t.Errorf("expected file content to remain unchanged, got %q", string(b))
	}
}

func TestMrRulesApplyCmd_DryRun(t *testing.T) {
	resetFlags()
	tempDir := t.TempDir()
	rulesFile := filepath.Join(tempDir, "rules.json")

	rulesContent := `{
		"schemaVersion": "1",
		"mergeMethod": "rebase_merge"
	}`
	_ = os.WriteFile(rulesFile, []byte(rulesContent), 0o644)

	rulesPath = rulesFile
	dryRun = true
	projectOpt = "mygroup/myproject"
	hostOpt = "gitlab.example.com"

	cmd := mrRulesApplyCmd
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("dry-run failed with error: %v", err)
	}
}

func TestMrRulesApplyCmd_MockExecution(t *testing.T) {
	resetFlags()
	tempDir := t.TempDir()
	rulesFile := filepath.Join(tempDir, "rules.json")

	rulesContent := `{
		"schemaVersion": "1",
		"mergeMethod": "merge_commit",
		"pipelinesMustSucceed": true
	}`
	_ = os.WriteFile(rulesFile, []byte(rulesContent), 0o644)

	// Mock execCommand
	original := execCommand
	defer func() { execCommand = original }()

	var invokedName string
	var invokedArgs []string
	execCommand = func(name string, arg ...string) *exec.Cmd {
		invokedName = name
		invokedArgs = arg
		return exec.Command("true")
	}

	// Setup parameters
	rulesPath = rulesFile
	dryRun = false
	projectOpt = "mygroup/myproject"
	hostOpt = "gitlab.example.com"
	viper.Set("gitlab.token", "glpat-token123")

	cmd := mrRulesApplyCmd
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("apply command failed: %v", err)
	}

	if invokedName != "glab" {
		t.Errorf("expected glab command, got %q", invokedName)
	}

	// Verify glab args
	argStr := strings.Join(invokedArgs, " ")
	if !strings.Contains(argStr, "api projects/mygroup%2Fmyproject -X PUT") {
		t.Errorf("glab command args incorrect: %v", invokedArgs)
	}
	if !strings.Contains(argStr, "-F merge_method=merge_commit") {
		t.Errorf("expected merge method flag in glab command: %v", invokedArgs)
	}
	if !strings.Contains(argStr, "-F only_allow_merge_if_pipeline_succeeds=true") {
		t.Errorf("expected pipeline succeeds flag in glab command: %v", invokedArgs)
	}
}

// ============================================================================
// E - Examples (Executable Documentation)
// ============================================================================

func ExampleGitLabCmd() {
	fmt.Println("Use:", GitLabCmd.Use)
	fmt.Println("Short:", GitLabCmd.Short)
	// Output:
	// Use: gitlab
	// Short: Interact with GitLab via glab
}

// ============================================================================
// B - Benchmarks (Performance Profiling)
// ============================================================================

func BenchmarkGitLabCmd_Init(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GitLabCmd.Commands()
	}
}
