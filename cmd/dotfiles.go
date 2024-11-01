/*
Copyright © 2023 Eric N. Garcia <eng618@garciaericn.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
)

// dotfilesCmd represents the dotfiles command
var dotfilesCmd = &cobra.Command{
	Use:   "dotfiles",
	Short: "Manage dotfiles",
	Long:  `This command is used to facilitate the management of private hidden dot files.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("dotfiles called")
	},
}

func init() {
	rootCmd.AddCommand(dotfilesCmd)

	dotfilesCmd.AddCommand(sync)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dotfilesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dotfilesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

var sync = &cobra.Command{
	Use:   "sync",
	Short: "sync your local bear repository",
	Long:  `This command fetches and pulls in remote changes to the local bare dot repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("sync called")

		// TODO: Use configs to specify config value... in my case that is `cfg`
		// TODO: fetch all changes `cfg fetch --all --prune --jobs=10`
		// TODO: pull latest changes `cfg pull --rebase --autostash`
		// TODO: if possible... reset the shell, if not prompt message for user to restart shell.

		syncLocal := exec.Command("echo", "hello world")

		utils.StartChildProcess(syncLocal)
	},
}
