/*
Copyright © 2024 Eric N. Garcia <eng618@garciaericn.com>

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
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
)

// systemCmd represents the system command
var systemCmd = &cobra.Command{
	Use:   "system",
	Short: "A command for managing the system",
	Long:  `This command will help manage various aspects of MacOS.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("system called")
	},
}

func init() {
	rootCmd.AddCommand(systemCmd)

	systemCmd.AddCommand(killPort)

	// TODO: Ubuntu update command

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// systemCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// systemCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

var killPort = &cobra.Command{
	Use:   "killPort",
	Short: "Kill a supplied port",
	Long:  `This will find what process is running on a supplied port, then kill that process.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			KillString := fmt.Sprintf("kill -9 $(lsof -ti:%s)", strings.Join(args, ","))
			log.Message("This should run: %s", KillString)

			killPortCmd := exec.Command(KillString)
			utils.StartChildProcess(killPortCmd)
		} else {
			log.Warn("You need to supply the port to kill, which will run the following:")
		}
	},
}
