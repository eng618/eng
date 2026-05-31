package files_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eng618/eng/cmd/files"
)

func TestFilesCmd(t *testing.T) {
	cmd := files.FilesCmd

	assert.NotNil(t, cmd)
	assert.Equal(t, "files", cmd.Use)
	assert.Equal(t, "A command for managing files", cmd.Short)
	assert.Equal(
		t,
		"This command will help manage various aspects of file operations on MacOS and Linux systems.",
		cmd.Long,
	)

	assert.True(t, cmd.HasSubCommands())

	var subCommands []string
	for _, c := range cmd.Commands() {
		subCommands = append(subCommands, c.Name())
	}

	assert.Contains(t, subCommands, "findAndDelete")
	assert.Contains(t, subCommands, "findNonMovieFolders")
}

func TestFilesCmd_Run(t *testing.T) {
	cmd := files.FilesCmd

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	cmd.Run(cmd, []string{})

	assert.Contains(t, buf.String(), "Usage:")
	assert.Contains(t, buf.String(), "files [command]")
}

func TestFilesCmd_Init_Flags(t *testing.T) {
	// FindAndDeleteCmd flags
	fadCmd := files.FindAndDeleteCmd
	assert.NotNil(t, fadCmd.Flags().Lookup("glob"))
	assert.Equal(t, "g", fadCmd.Flags().Lookup("glob").Shorthand)

	assert.NotNil(t, fadCmd.Flags().Lookup("ext"))
	assert.Equal(t, "e", fadCmd.Flags().Lookup("ext").Shorthand)

	assert.NotNil(t, fadCmd.Flags().Lookup("filename"))
	assert.Equal(t, "f", fadCmd.Flags().Lookup("filename").Shorthand)

	assert.NotNil(t, fadCmd.Flags().Lookup("list-extensions"))
	assert.Equal(t, "l", fadCmd.Flags().Lookup("list-extensions").Shorthand)

	// FindNonMovieFoldersCmd flags
	fnmfCmd := files.FindNonMovieFoldersCmd
	assert.NotNil(t, fnmfCmd.Flags().Lookup("dry-run"))
	assert.Equal(t, "true", fnmfCmd.Flags().Lookup("dry-run").DefValue)
}
