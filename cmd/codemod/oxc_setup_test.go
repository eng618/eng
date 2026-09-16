package codemod

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
)

func chdirTemp(t *testing.T) {
	t.Helper()
	tempDir := t.TempDir()
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, os.Chdir(oldWd))
	})
	require.NoError(t, os.Chdir(tempDir))
}

func mockExecCommand(t *testing.T, fn func(name string, arg ...string) *exec.Cmd) {
	t.Helper()
	oldCommand := execCommand
	t.Cleanup(func() { execCommand = oldCommand })
	execCommand = fn
}

func mockOxcPrompt(t *testing.T, answer bool) {
	t.Helper()
	oldPrompt := oxcConfirmPrompt
	t.Cleanup(func() { oxcConfirmPrompt = oldPrompt })
	oxcConfirmPrompt = func(_ string, _ bool) (bool, error) { return answer, nil }
}

func writePackageJSON(t *testing.T, pkg map[string]interface{}) {
	t.Helper()
	data, err := json.Marshal(pkg)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile("package.json", data, 0o644))
}

func TestResolveOxcPreset_Explicit(t *testing.T) {
	for _, preset := range []string{"recommended", "next", "vite", "react", "typescript", "base"} {
		got, err := resolveOxcPreset(preset)
		assert.NoError(t, err)
		assert.Equal(t, preset, got)
	}
}

func TestResolveOxcPreset_Invalid(t *testing.T) {
	_, err := resolveOxcPreset("bogus")
	assert.Error(t, err)
}

func TestResolveOxcPreset_Auto(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T)
		expected string
	}{
		{
			name: "next project",
			setup: func(t *testing.T) {
				writePackageJSON(t, map[string]interface{}{
					"dependencies": map[string]interface{}{"next": "latest"},
				})
			},
			expected: "next",
		},
		{
			name: "vite project",
			setup: func(t *testing.T) {
				writePackageJSON(t, map[string]interface{}{
					"devDependencies": map[string]interface{}{"vite": "^6.0.0"},
				})
			},
			expected: "vite",
		},
		{
			name: "react without vite",
			setup: func(t *testing.T) {
				writePackageJSON(t, map[string]interface{}{
					"dependencies": map[string]interface{}{"react": "^19.0.0"},
				})
			},
			expected: "react",
		},
		{
			name: "typescript library",
			setup: func(t *testing.T) {
				writePackageJSON(t, map[string]interface{}{
					"devDependencies": map[string]interface{}{"typescript": "^5.0.0"},
				})
			},
			expected: "typescript",
		},
		{
			name: "plain javascript",
			setup: func(t *testing.T) {
				writePackageJSON(t, map[string]interface{}{"name": "test"})
				require.NoError(t, os.WriteFile("index.js", []byte("console.log(1)"), 0o644))
			},
			expected: "recommended",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chdirTemp(t)
			tt.setup(t)
			got, err := resolveOxcPreset("auto")
			require.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestWriteOxcConfigs_Presets(t *testing.T) {
	tests := []struct {
		preset   string
		contains string
	}{
		{"recommended", "@gv-tech/oxc-config/recommended"},
		{"next", "@gv-tech/oxc-config/next"},
		{"vite", "@gv-tech/oxc-config/vite"},
		{"react", "@gv-tech/oxc-config/react"},
		{"typescript", "@gv-tech/oxc-config/typescript"},
		{"base", "@gv-tech/oxc-config/base"},
	}
	for _, tt := range tests {
		t.Run(tt.preset, func(t *testing.T) {
			chdirTemp(t)
			require.NoError(t, writeOxcConfigs(tt.preset, false))
			lintData, err := os.ReadFile("oxlint.config.ts")
			require.NoError(t, err)
			assert.Contains(t, string(lintData), tt.contains)
			fmtData, err := os.ReadFile("oxfmt.config.ts")
			require.NoError(t, err)
			assert.Contains(t, string(fmtData), "@gv-tech/oxc-config/oxfmt")
		})
	}
}

func TestWriteOxcConfigs_TypeAware(t *testing.T) {
	chdirTemp(t)
	require.NoError(t, writeOxcConfigs("vite", true))
	data, err := os.ReadFile("oxlint.config.ts")
	require.NoError(t, err)
	assert.Contains(t, string(data), "typeAware")
	assert.Contains(t, string(data), "@gv-tech/oxc-config/type-aware")
}

func TestDetectEslintPrettier(t *testing.T) {
	t.Run("detects eslint dep", func(t *testing.T) {
		chdirTemp(t)
		writePackageJSON(t, map[string]interface{}{
			"devDependencies": map[string]interface{}{"eslint": "^9.0.0"},
		})
		assert.True(t, detectEslintPrettier())
	})
	t.Run("detects prettier key", func(t *testing.T) {
		chdirTemp(t)
		writePackageJSON(t, map[string]interface{}{
			"name":     "test",
			"prettier": "@eng618/prettier-config",
		})
		assert.True(t, detectEslintPrettier())
	})
	t.Run("detects eslint config file", func(t *testing.T) {
		chdirTemp(t)
		writePackageJSON(t, map[string]interface{}{"name": "test"})
		require.NoError(t, os.WriteFile("eslint.config.mjs", []byte("export default []"), 0o644))
		assert.True(t, detectEslintPrettier())
	})
	t.Run("clean oxc project", func(t *testing.T) {
		chdirTemp(t)
		writePackageJSON(t, map[string]interface{}{"name": "test"})
		assert.False(t, detectEslintPrettier())
	})
}

func TestRemoveEslintPrettier(t *testing.T) {
	chdirTemp(t)
	writePackageJSON(t, map[string]interface{}{
		"name": "test",
		"devDependencies": map[string]interface{}{
			"eslint":   "^9.0.0",
			"prettier": "^3.0.0",
		},
	})
	require.NoError(t, os.WriteFile("eslint.config.mjs", []byte("export default []"), 0o644))

	var removed []string
	mockExecCommand(t, func(name string, arg ...string) *exec.Cmd {
		removed = append(removed, append([]string{name}, arg...)...)
		return exec.Command("echo", append([]string{name}, arg...)...)
	})

	require.NoError(t, removeEslintPrettier())
	joined := strings.Join(removed, " ")
	assert.Contains(t, joined, "eslint")
	assert.Contains(t, joined, "prettier")
	_, err := os.Stat("eslint.config.mjs")
	assert.True(t, os.IsNotExist(err))
}

func TestUpdatePackageJSONForOxc(t *testing.T) {
	chdirTemp(t)
	writePackageJSON(t, map[string]interface{}{
		"name":     "testpkg",
		"prettier": "@eng618/prettier-config",
		"foo":      "bar",
	})
	require.NoError(t, updatePackageJSONForOxc())
	out, err := os.ReadFile("package.json")
	require.NoError(t, err)
	content := string(out)
	assert.Contains(t, content, `"lint": "oxlint ."`)
	assert.Contains(t, content, `"format": "oxfmt --write ."`)
	assert.Contains(t, content, `"format:ci": "oxfmt --check ."`)
	assert.Contains(t, content, `"lint-staged":`)
	assert.NotContains(t, content, `"prettier":`)
	assert.Contains(t, content, "foo")
}

func TestInstallOxcDependencies_TypeAware(t *testing.T) {
	chdirTemp(t)
	writePackageJSON(t, map[string]interface{}{"name": "test"})
	var gotArgs []string
	mockExecCommand(t, func(name string, arg ...string) *exec.Cmd {
		gotArgs = append([]string{name}, arg...)
		return exec.Command("echo", append([]string{name}, arg...)...)
	})
	require.NoError(t, installOxcDependencies("recommended", true))
	joined := strings.Join(gotArgs, " ")
	assert.Contains(t, joined, "@gv-tech/oxc-config@latest")
	assert.Contains(t, joined, "oxlint@latest")
	assert.Contains(t, joined, "oxfmt@latest")
	assert.Contains(t, joined, "oxlint-tsgolint@latest")
}

func TestOxcSetupCmd_FullRun_RemovesEslint(t *testing.T) {
	chdirTemp(t)
	log.SetWriters(&strings.Builder{}, &strings.Builder{})
	t.Cleanup(log.ResetWriters)
	ui.DisableProgress = true
	t.Cleanup(func() { ui.DisableProgress = false })

	writePackageJSON(t, map[string]interface{}{
		"name": "testpkg",
		"devDependencies": map[string]interface{}{
			"eslint":   "^9.0.0",
			"prettier": "^3.0.0",
		},
	})
	require.NoError(t, os.WriteFile("eslint.config.mjs", []byte("export default []"), 0o644))
	require.NoError(t, os.WriteFile("index.js", []byte("console.log(1)"), 0o644))

	mockExecCommand(t, func(name string, arg ...string) *exec.Cmd {
		if name == "npx" {
			_ = os.MkdirAll(".husky", 0o755)
		}
		if name == "node" {
			return exec.Command("echo", "v22.18.0")
		}
		return exec.Command("echo", append([]string{name}, arg...)...)
	})

	oldPreset, oldTypeAware, oldRemove, oldYes := oxcPreset, oxcTypeAware, oxcRemoveEslint, oxcYes
	t.Cleanup(func() {
		oxcPreset, oxcTypeAware, oxcRemoveEslint, oxcYes = oldPreset, oldTypeAware, oldRemove, oldYes
	})
	oxcPreset, oxcTypeAware, oxcRemoveEslint, oxcYes = "auto", false, true, true

	OxcSetupCmd.Run(OxcSetupCmd, []string{})

	lintData, err := os.ReadFile("oxlint.config.ts")
	require.NoError(t, err)
	assert.Contains(t, string(lintData), "@gv-tech/oxc-config")
	_, err = os.ReadFile("oxfmt.config.ts")
	require.NoError(t, err)
	_, err = os.Stat("eslint.config.mjs")
	assert.True(t, os.IsNotExist(err))

	pkgData, err := os.ReadFile("package.json")
	require.NoError(t, err)
	assert.Contains(t, string(pkgData), "oxlint .")
	var pkg map[string]interface{}
	require.NoError(t, json.Unmarshal(pkgData, &pkg))
	_, hasPrettierKey := pkg["prettier"]
	assert.False(t, hasPrettierKey, "top-level prettier key should be removed")
}

func TestOxcSetupCmd_PromptsOnDetect(t *testing.T) {
	chdirTemp(t)
	log.SetWriters(&strings.Builder{}, &strings.Builder{})
	t.Cleanup(log.ResetWriters)
	ui.DisableProgress = true
	t.Cleanup(func() { ui.DisableProgress = false })

	writePackageJSON(t, map[string]interface{}{
		"name": "testpkg",
		"devDependencies": map[string]interface{}{
			"eslint": "^9.0.0",
		},
	})
	require.NoError(t, os.WriteFile("index.js", []byte("console.log(1)"), 0o644))

	mockExecCommand(t, func(name string, arg ...string) *exec.Cmd {
		if name == "npx" {
			_ = os.MkdirAll(".husky", 0o755)
		}
		if name == "node" {
			return exec.Command("echo", "v22.18.0")
		}
		return exec.Command("echo", append([]string{name}, arg...)...)
	})
	// Decline the removal prompt; legacy setup must survive.
	mockOxcPrompt(t, false)

	oldPreset, oldTypeAware, oldRemove, oldYes := oxcPreset, oxcTypeAware, oxcRemoveEslint, oxcYes
	t.Cleanup(func() {
		oxcPreset, oxcTypeAware, oxcRemoveEslint, oxcYes = oldPreset, oldTypeAware, oldRemove, oldYes
	})
	oxcPreset, oxcTypeAware, oxcRemoveEslint, oxcYes = "base", false, false, false

	OxcSetupCmd.Run(OxcSetupCmd, []string{})

	_, err := os.ReadFile("oxlint.config.ts")
	require.NoError(t, err)
	pkgData, err := os.ReadFile("package.json")
	require.NoError(t, err)
	// Declined removal keeps the eslint dep.
	assert.Contains(t, string(pkgData), "eslint")
}

func TestDetectReactUsage(t *testing.T) {
	t.Run("react dep", func(t *testing.T) {
		chdirTemp(t)
		writePackageJSON(t, map[string]interface{}{
			"dependencies": map[string]interface{}{"react": "^19.0.0"},
		})
		assert.True(t, detectReactUsage())
	})
	t.Run("tsx file", func(t *testing.T) {
		chdirTemp(t)
		writePackageJSON(t, map[string]interface{}{"name": "test"})
		require.NoError(t, os.WriteFile("app.tsx", []byte("export default 1"), 0o644))
		assert.True(t, detectReactUsage())
	})
	t.Run("no react", func(t *testing.T) {
		chdirTemp(t)
		writePackageJSON(t, map[string]interface{}{"name": "test"})
		assert.False(t, detectReactUsage())
	})
}

func TestDetectViteUsage(t *testing.T) {
	t.Run("vite dep", func(t *testing.T) {
		chdirTemp(t)
		writePackageJSON(t, map[string]interface{}{
			"devDependencies": map[string]interface{}{"vite": "^6.0.0"},
		})
		assert.True(t, detectViteUsage())
	})
	t.Run("vite config", func(t *testing.T) {
		chdirTemp(t)
		writePackageJSON(t, map[string]interface{}{"name": "test"})
		require.NoError(t, os.WriteFile("vite.config.ts", []byte("export default {}"), 0o644))
		assert.True(t, detectViteUsage())
	})
	t.Run("no vite", func(t *testing.T) {
		chdirTemp(t)
		writePackageJSON(t, map[string]interface{}{"name": "test"})
		assert.False(t, detectViteUsage())
	})
}
