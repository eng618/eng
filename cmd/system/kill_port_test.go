package system

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

func BenchmarkParsePortOutputLsof(b *testing.B) {
	output := `COMMAND  PID USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
node    1234 user   23u  IPv4 0x1234567890      0t0  TCP *:3000 (LISTEN)
python  5678 user   24u  IPv4 0x1234567891      0t0  TCP *:8000 (LISTEN)`
	for i := 0; i < 100; i++ {
		output += "\nnode    1234 user   23u  IPv4 0x1234567890      0t0  TCP *:3000 (LISTEN)"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parsePortOutput(output, "lsof", "")
	}
}

func BenchmarkParsePortOutputSS(b *testing.B) {
	output := `Netid State Recv-Q Send-Q Local Address:Port Peer Address:Port Process
tcp   LISTEN 0      128          0.0.0.0:22             0.0.0.0:*      users:(("sshd",pid=1003,fd=3))`
	for i := 0; i < 100; i++ {
		output += "\ntcp   LISTEN 0      128          0.0.0.0:22             0.0.0.0:*      users:((\"sshd\",pid=1003,fd=3))"
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parsePortOutput(output, "ss", "")
	}
}
