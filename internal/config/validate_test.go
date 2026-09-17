package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidate_Table(t *testing.T) {
	tmp := t.TempDir()
	tests := []struct {
		name   string
		mutate func(*ResolvedConfig)
		fields []string
	}{
		{"valid minimal", func(c *ResolvedConfig) {}, nil},
		{
			"bad email",
			func(c *ResolvedConfig) { c.Email = "not-an-email" },
			[]string{"email"},
		},
		{
			"bad repo url",
			func(c *ResolvedConfig) { c.DotfilesRepo = ":// bad url" },
			[]string{"dotfiles.repo_url"},
		},
		{
			"scp repo url ok",
			func(c *ResolvedConfig) { c.DotfilesRepo = "git@github.com:eng618/dotfiles.git" },
			nil,
		},
		{
			"https repo url ok",
			func(c *ResolvedConfig) { c.DotfilesRepo = "https://github.com/eng618/dotfiles.git" },
			nil,
		},
		{
			"bad branch",
			func(c *ResolvedConfig) { c.DotfilesBranch = "bad branch" },
			[]string{"dotfiles.branch"},
		},
		{
			"dev path is file",
			func(c *ResolvedConfig) { c.GitDevPath = makeTempFile(t) },
			[]string{"git.dev_path"},
		},
		{
			"missing path ok when parent exists",
			func(c *ResolvedConfig) { c.GitDevPath = tmp + "/not-yet-created" },
			nil,
		},
		{
			"multiple errors",
			func(c *ResolvedConfig) { c.Email = "bad"; c.DotfilesBranch = "has space" },
			[]string{"email", "dotfiles.branch"},
		},
		{
			"bad proxy url",
			func(c *ResolvedConfig) {
				c.Proxies = []ProxyConfig{{Title: "corp", Value: "not a url"}}
			},
			[]string{"proxies[0].value"},
		},
		{
			"proxy missing port",
			func(c *ResolvedConfig) {
				c.Proxies = []ProxyConfig{{Title: "corp", Value: "http://proxy"}}
			},
			[]string{"proxies[0].value"},
		},
		{
			"valid proxies",
			func(c *ResolvedConfig) {
				c.Proxies = []ProxyConfig{
					{Title: "corp", Value: "http://proxy:8080", Enabled: true},
					{Title: "home", Value: "socks5://127.0.0.1:1080"},
				}
			},
			nil,
		},
		{
			"duplicate proxy titles",
			func(c *ResolvedConfig) {
				c.Proxies = []ProxyConfig{
					{Title: "corp", Value: "http://a:8080"},
					{Title: "corp", Value: "http://b:8080"},
				}
			},
			[]string{"proxies[1].title"},
		},
		{
			"multiple enabled proxies",
			func(c *ResolvedConfig) {
				c.Proxies = []ProxyConfig{
					{Title: "a", Value: "http://a:8080", Enabled: true},
					{Title: "b", Value: "http://b:8080", Enabled: true},
				}
			},
			[]string{"proxies"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ResolvedConfig{DotfilesBranch: "main", GitDevPath: tmp}
			tt.mutate(c)
			errs := c.Validate()
			if len(tt.fields) == 0 {
				require.Empty(t, errs)
				return
			}
			got := make([]string, 0, len(errs))
			for _, e := range errs {
				got = append(got, e.Field)
				require.NotEmpty(t, e.Hint)
			}
			require.ElementsMatch(t, tt.fields, got)
		})
	}
}

func TestFieldError_Message(t *testing.T) {
	e := FieldError{Field: "email", Value: "bad", Hint: "fix it"}
	require.Contains(t, e.Error(), "email")
	require.Contains(t, e.Error(), "fix it")
}

func makeTempFile(t *testing.T) string {
	t.Helper()
	f, err := os.Create(filepath.Join(t.TempDir(), "f"))
	require.NoError(t, err)
	require.NoError(t, f.Close())
	return f.Name()
}
