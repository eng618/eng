package oxfmt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptionsValidate(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		wantErr string
		want    Options
	}{
		{
			name: "defaults",
			opts: Options{},
			want: Options{Path: ".", Runner: RunnerAuto},
		},
		{
			name: "write mode",
			opts: Options{Path: "web", Write: true, Runner: "npx"},
			want: Options{Path: "web", Write: true, Runner: "npx"},
		},
		{
			name:    "check and write conflict",
			opts:    Options{Write: true, Check: true},
			wantErr: "cannot combine --check and --write",
		},
		{
			name:    "diff and write conflict",
			opts:    Options{Write: true, Diff: true},
			wantErr: "cannot combine --diff and --write",
		},
		{
			name:    "invalid runner",
			opts:    Options{Runner: "yarn"},
			wantErr: `invalid --runner "yarn"`,
		},
		{
			name: "runner case-insensitive",
			opts: Options{Runner: "BUN"},
			want: Options{Path: ".", Runner: RunnerBun},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, tt.opts)
		})
	}
}
