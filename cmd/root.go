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
	"os"

	"github.com/eng618/eng/cmd/codemod"
	"github.com/eng618/eng/cmd/config"
	"github.com/eng618/eng/cmd/dotfiles"
	"github.com/eng618/eng/cmd/git"
	"github.com/eng618/eng/cmd/system"
	"github.com/eng618/eng/cmd/ts"
	"github.com/eng618/eng/cmd/version"
	"github.com/eng618/eng/utils"
	"github.com/eng618/eng/utils/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "eng",
	Short: "A personal cli to facilitate my workflow.",
	Long: `
                                          __ __
                                         |  \  \
  ______  _______   ______        _______| ▓▓\▓▓
 /      \|       \ /      \      /       \ ▓▓  \
|  ▓▓▓▓▓▓\ ▓▓▓▓▓▓▓\  ▓▓▓▓▓▓\    |  ▓▓▓▓▓▓▓ ▓▓ ▓▓
| ▓▓    ▓▓ ▓▓  | ▓▓ ▓▓  | ▓▓    | ▓▓     | ▓▓ ▓▓
| ▓▓▓▓▓▓▓▓ ▓▓  | ▓▓ ▓▓__| ▓▓    | ▓▓_____| ▓▓ ▓▓
 \▓▓     \ ▓▓  | ▓▓\▓▓    ▓▓     \▓▓     \ ▓▓ ▓▓
  \▓▓▓▓▓▓▓\▓▓   \▓▓_\▓▓▓▓▓▓▓      \▓▓▓▓▓▓▓\▓▓\▓▓
                  |  \__| ▓▓
                   \▓▓    ▓▓
                    \▓▓▓▓▓▓

This is personal cli to facilitate my workflow. An maintain my development machine.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	cobra.CheckErr(err)
}

func init() {
	cobra.OnInitialize(initConfig)

	// Set the version string for the root command's --version flag
	// Use the exported Version variable from the version package
	rootCmd.Version = version.Version

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.eng.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")

	// Bind the verbose flag to viper config
	err := viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	cobra.CheckErr(err)

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

	// Add subcommands
	rootCmd.AddCommand(system.SystemCmd)
	rootCmd.AddCommand(dotfiles.DotfilesCmd)
	rootCmd.AddCommand(git.GitCmd)
	rootCmd.AddCommand(config.ConfigCmd)
	rootCmd.AddCommand(ts.TailscaleCmd)
	rootCmd.AddCommand(version.VersionCmd)
	rootCmd.AddCommand(codemod.CodemodCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".eng" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".eng")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Verbose(utils.IsVerbose(rootCmd), "Using config file: %s", viper.ConfigFileUsed())
	} else if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
		// Config file was found but another error was produced
		log.Warn("Error reading config file %s: %v", viper.ConfigFileUsed(), err)
	} else {
		log.Verbose(utils.IsVerbose(rootCmd), "No config file found, using defaults.")
	}
}
