package system

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
