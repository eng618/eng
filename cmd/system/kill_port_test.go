package system

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eng618/eng/internal/log"
)

func TestFindPortTool(t *testing.T) {
	// This test assumes lsof is available, as in the environment
	tool := findPortTool()
	assert.NotEmpty(t, tool)
	assert.Contains(t, []string{"lsof", "ss", "netstat"}, tool)
}

func TestParsePortOutput(t *testing.T) {
	// Mock lsof output
	output := `COMMAND  PID USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
node    1234 user   23u  IPv4 0x1234567890      0t0  TCP *:3000 (LISTEN)
python  5678 user   24u  IPv4 0x1234567891      0t0  TCP *:8000 (LISTEN)`

	ports, err := parsePortOutput(output, "lsof", "")
	assert.NoError(t, err)
	assert.Len(t, ports, 2)
	assert.Equal(t, "node", ports[0].Command)
	assert.Equal(t, "1234", ports[0].PID)
	assert.Equal(t, "3000", ports[0].Port)
	assert.Equal(t, "python", ports[1].Command)
	assert.Equal(t, "5678", ports[1].PID)
	assert.Equal(t, "8000", ports[1].Port)
}

func TestParsePortOutputWithFilter(t *testing.T) {
	output := `COMMAND  PID USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
node    1234 user   23u  IPv4 0x1234567890      0t0  TCP *:3000 (LISTEN)
python  5678 user   24u  IPv4 0x1234567891      0t0  TCP *:8000 (LISTEN)`

	ports, err := parsePortOutput(output, "lsof", "node")
	assert.NoError(t, err)
	assert.Len(t, ports, 1)
	assert.Equal(t, "node", ports[0].Command)
}

func TestParsePortList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     []string
		wantErrs []string
	}{
		{"single port", "3000", []string{"3000"}, nil},
		{"multiple ports", "3000,8080,9000", []string{"3000", "8080", "9000"}, nil},
		{"ports with whitespace", " 3000 , 8080 ", []string{"3000", "8080"}, nil},
		{"min port", "1", []string{"1"}, nil},
		{"max port", "65535", []string{"65535"}, nil},
		{"empty string", "", nil, []string{"port list cannot be empty"}},
		{"whitespace only", "   ", nil, []string{"port list cannot be empty"}},
		{"leading comma", ",3000", []string{"3000"}, []string{"empty port in list"}},
		{"trailing comma", "3000,", []string{"3000"}, []string{"empty port in list"}},
		{"double comma", "3000,,8080", []string{"3000", "8080"}, []string{"empty port in list"}},
		{"non integer", "abc", nil, []string{`invalid port number "abc"`}},
		{"port zero", "0", nil, []string{"port must be between 1 and 65535"}},
		{"negative port", "-1", nil, []string{"port must be between 1 and 65535"}},
		{"port too high", "65536", nil, []string{"port must be between 1 and 65535"}},
		{"multiple problems", "3000,abc,8080,99999", []string{"3000", "8080"}, []string{
			`invalid port number "abc"`,
			"port must be between 1 and 65535",
		}},
		{"all problems", "abc,def,", nil, []string{
			`invalid port number "abc"`,
			`invalid port number "def"`,
			"empty port in list",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, errs := parsePortList(tt.input)
			if tt.wantErrs != nil {
				assert.Len(t, errs, len(tt.wantErrs))
				for i, want := range tt.wantErrs {
					assert.Contains(t, errs[i].Error(), want)
				}
				assert.Equal(t, tt.want, got)
				return
			}
			assert.Empty(t, errs)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestKillPortCmd_InvalidPort_ExitsEarly(t *testing.T) {
	tests := []struct {
		name     string
		args     string
		wantErrs []string
	}{
		{"non integer", "abc", []string{`invalid port number "abc": port must be an integer`}},
		{"empty entry", "3000,,8080", []string{"empty port in list"}},
		{"out of range", "99999", []string{"port must be between 1 and 65535"}},
		{"multiple problems", "3000,abc,99999,", []string{
			"Found 3 problem(s) in port list",
			`invalid port number "abc"`,
			"port must be between 1 and 65535",
			"empty port in list",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log.SetWriters(&buf, &buf)
			defer log.ResetWriters()

			KillPortCmd.SetOut(&buf)
			KillPortCmd.SetErr(&buf)

			KillPortCmd.Run(KillPortCmd, []string{tt.args})

			out := buf.String()
			for _, want := range tt.wantErrs {
				assert.Contains(t, out, want)
			}
			assert.NotContains(t, out, "Attempting to find process on port")
		})
	}
}
