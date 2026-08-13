package system

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eng618/eng/internal/log"
)

func TestParseProcessOutput(t *testing.T) {
	// Mock ps aux output
	output := `USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
user      1234  0.0  0.1  12345  6789 pts/0    S    10:00   0:00 node app.js
user      5678  0.0  0.2  23456  7890 pts/1    S    11:00   0:01 python server.py`

	processes, err := parseProcessOutput(output, "")
	assert.NoError(t, err)
	assert.Len(t, processes, 2)
	assert.Equal(t, "node app.js", processes[0].Command)
	assert.Equal(t, "1234", processes[0].PID)
	assert.Equal(t, "user", processes[0].User)
	assert.Equal(t, "python server.py", processes[1].Command)
	assert.Equal(t, "5678", processes[1].PID)
}

func TestParseProcessOutputWithFilter(t *testing.T) {
	output := `USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
user      1234  0.0  0.1  12345  6789 pts/0    S    10:00   0:00 node app.js
user      5678  0.0  0.2  23456  7890 pts/1    S    11:00   0:01 python server.py`

	processes, err := parseProcessOutput(output, "node")
	assert.NoError(t, err)
	assert.Len(t, processes, 1)
	assert.Equal(t, "node app.js", processes[0].Command)
}

func TestParseProcessList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     []string
		wantErrs []string
	}{
		{"single PID", "1234", []string{"1234"}, nil},
		{"multiple PIDs", "1234,5678,9012", []string{"1234", "5678", "9012"}, nil},
		{"PIDs with whitespace", " 1234 , 5678 ", []string{"1234", "5678"}, nil},
		{"valid PID 1", "1", []string{"1"}, nil},
		{"empty string", "", nil, []string{"process PID list cannot be empty"}},
		{"whitespace only", "   ", nil, []string{"process PID list cannot be empty"}},
		{"leading comma", ",1234", []string{"1234"}, []string{"empty PID in list"}},
		{"trailing comma", "1234,", []string{"1234"}, []string{"empty PID in list"}},
		{"double comma", "1234,,5678", []string{"1234", "5678"}, []string{"empty PID in list"}},
		{"non integer", "abc", nil, []string{`invalid PID "abc": PID must be an integer`}},
		{"PID zero", "0", nil, []string{"PID must be greater than 0"}},
		{"negative PID", "-1234", nil, []string{"PID must be greater than 0"}},
		{"multiple problems", "1234,abc,5678,0", []string{"1234", "5678"}, []string{
			`invalid PID "abc": PID must be an integer`,
			"PID must be greater than 0",
		}},
		{"all problems", "abc,def,", nil, []string{
			`invalid PID "abc": PID must be an integer`,
			`invalid PID "def": PID must be an integer`,
			"empty PID in list",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, errs := parseProcessList(tt.input)
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

func TestKillProcessCmd_InvalidPID_ExitsEarly(t *testing.T) {
	tests := []struct {
		name     string
		args     string
		wantErrs []string
	}{
		{"non integer", "abc", []string{`invalid PID "abc": PID must be an integer`}},
		{"empty entry", "1234,,5678", []string{"empty PID in list"}},
		{"negative pid", "-10", []string{"PID must be greater than 0"}},
		{"multiple problems", "1234,abc,-5,", []string{
			"Found 3 problem(s) in process PID list",
			`invalid PID "abc": PID must be an integer`,
			"PID must be greater than 0",
			"empty PID in list",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log.SetWriters(&buf, &buf)
			defer log.ResetWriters()

			KillProcessCmd.SetOut(&buf)
			KillProcessCmd.SetErr(&buf)

			KillProcessCmd.Run(KillProcessCmd, []string{tt.args})

			out := buf.String()
			for _, want := range tt.wantErrs {
				assert.Contains(t, out, want)
			}
			assert.NotContains(t, out, "Attempting to kill process with PID")
		})
	}
}
