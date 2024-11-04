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

	"github.com/spf13/cobra"

	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
)

// tailscaleCMD represents the system command
var tailscaleCMD = &cobra.Command{
	Use:   "tailscale",
	Short: "A helper for the tailscale command",
	Long:  `This command will help manage various aspects of MacOS.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("system called")
	},
	Aliases: []string{"ts"},
}

func init() {
	rootCmd.AddCommand(tailscaleCMD)

	tailscaleCMD.AddCommand(ts_up)
	tailscaleCMD.AddCommand(ts_down)

	// TODO: Ubuntu update command

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// tailScaleCMD.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// tailScaleCMD.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

var ts_up = &cobra.Command{
	Use:   "up",
	Short: "bring up the tailscale service",
	Long:  `This call 'sudo tailscale up' under the hood..`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Bringing up the tailscale service")

		// TODO: build utility to verify command exists before trying.

		tsUpCmd := exec.Command("/bin/sh", "-c", "sudo tailscale up")
		utils.StartChildProcess(tsUpCmd)
	},
}

var ts_down = &cobra.Command{
	Use:   "down",
	Short: "take down the tailscale service",
	Long:  `This call 'sudo tailscale down' under the hood..`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Start("Taking down the tailscale service")

		tsDownCmd := exec.Command("/bin/sh", "-c", "sudo tailscale down")
		utils.StartChildProcess(tsDownCmd)
	},
}
