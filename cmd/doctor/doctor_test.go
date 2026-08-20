package doctor

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/stretchr/testify/assert"
)

type mockFileInfo struct {
	isDir bool
}

func (m mockFileInfo) Name() string       { return "mock" }
func (m mockFileInfo) Size() int64        { return 0 }
func (m mockFileInfo) Mode() fs.FileMode  { return 0755 }
func (m mockFileInfo) ModTime() time.Time { return time.Now() }
func (m mockFileInfo) IsDir() bool        { return m.isDir }
func (m mockFileInfo) Sys() any           { return nil }

func TestDoctorCommand(t *testing.T) {
	// Redirect log output
	var buf bytes.Buffer
	origOut := log.Out
	log.Out = &buf
	defer func() { log.Out = origOut }()

	ui.DisableProgress = true
	defer func() { ui.DisableProgress = false }()

	// Mock execLookPath and osStat
	origLookPath := execLookPath
	origOsStat := osStat
	defer func() {
		execLookPath = origLookPath
		osStat = origOsStat
	}()

	execLookPath = func(file string) (string, error) {
		if file == "git" || file == "brew" || file == "bash" {
			return "/usr/bin/" + file, nil
		}
		return "", errors.New("not found")
	}

	osStat = func(name string) (os.FileInfo, error) {
		return mockFileInfo{isDir: true}, nil
	}

	err := DoctorCmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "eng Workstation Diagnostics")
	assert.Contains(t, output, "CLI Tools & Dependencies:")
	assert.Contains(t, output, "Git")
}
