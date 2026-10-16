package oxfmt

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubLookPath returns a LookPath func that resolves only bins on PATH.
func stubLookPath(t *testing.T, bins ...string) func(string) (string, error) {
	t.Helper()
	available := map[string]bool{}
	for _, b := range bins {
		available[b] = true
	}
	return func(file string) (string, error) {
		if available[file] {
			return "/bin/" + file, nil
		}
		return "", errors.New("not found: " + file)
	}
}

func TestResolveLauncher(t *testing.T) {
	tests := []struct {
		name       string
		runner     string
		config     string
		bins       []string
		bunLock    bool
		wantBin    string
		wantDirect bool
		wantErr    string
	}{
		{
			name:    "auto prefers bun wrapper in bun workspace",
			runner:  RunnerAuto,
			bins:    []string{"oxfmt", "bun", "npx"},
			bunLock: true,
			wantBin: "bun",
		},
		{
			name:    "auto skips bun without bun lock",
			runner:  RunnerAuto,
			bins:    []string{"oxfmt", "bun", "npx"},
			wantBin: "npx",
		},
		{
			name:    "auto prefers npx over local oxfmt for shared style",
			runner:  RunnerAuto,
			bins:    []string{"oxfmt", "npx"},
			wantBin: "npx",
		},
		{
			name:       "auto falls back to local oxfmt",
			runner:     RunnerAuto,
			bins:       []string{"oxfmt"},
			wantBin:    "oxfmt",
			wantDirect: true,
		},
		{
			name:    "auto errors with no runners",
			runner:  RunnerAuto,
			wantErr: "no JS runner found",
		},
		{
			name:       "explicit oxfmt",
			runner:     RunnerOxfmt,
			bins:       []string{"npx", "oxfmt"},
			wantBin:    "oxfmt",
			wantDirect: true,
		},
		{
			name:    "explicit oxfmt missing",
			runner:  RunnerOxfmt,
			bins:    []string{"npx"},
			wantErr: "oxfmt not found on PATH",
		},
		{
			name:    "explicit bun",
			runner:  RunnerBun,
			bins:    []string{"bun"},
			wantBin: "bun",
		},
		{
			name:    "explicit bun missing",
			runner:  RunnerBun,
			wantErr: "bun not found on PATH",
		},
		{
			name:    "explicit npx",
			runner:  RunnerNpx,
			bins:    []string{"npx"},
			wantBin: "npx",
		},
		{
			name:       "custom config prefers local oxfmt",
			runner:     RunnerAuto,
			config:     ".oxfmtrc.json",
			bins:       []string{"oxfmt", "npx"},
			wantBin:    "oxfmt",
			wantDirect: true,
		},
		{
			name:    "custom config via wrapper when oxfmt missing",
			runner:  RunnerAuto,
			config:  ".oxfmtrc.json",
			bins:    []string{"npx"},
			wantBin: "npx",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := Options{Runner: tt.runner, ConfigPath: tt.config}
			spec, err := ResolveLauncher(opts, stubLookPath(t, tt.bins...), tt.bunLock)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantBin, spec.Binary)
			assert.Equal(t, tt.wantDirect, spec.Direct)
			if !tt.wantDirect {
				assert.Contains(t, spec.Args, "gv-oxfmt")
				assert.Contains(t, spec.Args, ConfigPackage)
			}
		})
	}
}

func TestInvocationArgs(t *testing.T) {
	wrapper := LaunchSpec{Binary: "npx", Args: []string{"--yes", "gv-oxfmt"}}
	assert.Equal(t,
		[]string{"--yes", "gv-oxfmt", "--list-different", "web"},
		invocationArgs(wrapper, Options{}, "--list-different", "web"),
	)
	assert.Equal(t,
		[]string{"--yes", "gv-oxfmt", "-c", "custom.json", "--write", "web"},
		invocationArgs(wrapper, Options{ConfigPath: "custom.json"}, "--write", "web"),
	)
	direct := LaunchSpec{Binary: "oxfmt", Direct: true}
	assert.Equal(t,
		[]string{"--write", "."},
		invocationArgs(direct, Options{}, "--write", "."),
	)
}

func TestHasBunLockFor(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "bun.lock"), []byte("lock"), 0o644))
	nested := filepath.Join(root, "a", "b")
	require.NoError(t, os.MkdirAll(nested, 0o755))

	assert.True(t, hasBunLockFor(nested))
	assert.True(t, hasBunLockFor(filepath.Join(nested, "file.ts")))
	assert.False(t, hasBunLockFor(t.TempDir()))
	assert.False(t, hasBunLockFor(filepath.Join(t.TempDir(), "missing")))
}
