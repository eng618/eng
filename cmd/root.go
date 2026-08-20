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
	"context"
	"errors"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/cmd/asdf"
	"github.com/eng618/eng/cmd/codemod"
	"github.com/eng618/eng/cmd/compose"
	"github.com/eng618/eng/cmd/config"
	"github.com/eng618/eng/cmd/dotfiles"
	"github.com/eng618/eng/cmd/files"
	"github.com/eng618/eng/cmd/git"
	"github.com/eng618/eng/cmd/gitlab"
	"github.com/eng618/eng/cmd/immich"
	"github.com/eng618/eng/cmd/project"
	"github.com/eng618/eng/cmd/system"
	"github.com/eng618/eng/cmd/ts"
	"github.com/eng618/eng/cmd/version"
	"github.com/eng618/eng/internal/cmdutil"
	configUtils "github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands.
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

// GetRootCommand returns the root cobra.Command instance (used by tools/gendocs).
func GetRootCommand() *cobra.Command {
	return rootCmd
}

// ExecuteContext adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func ExecuteContext(ctx context.Context) {
	// Silence default error printing so we can use our custom error handler
	rootCmd.SilenceErrors = true
	// Silence usage printing on errors to avoid noisy output
	rootCmd.SilenceUsage = true

	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		theme.HandleError(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Custom Lip Gloss Help & Usage styling
	rootCmd.SetHelpFunc(ui.HelpFunc)
	rootCmd.SetUsageFunc(ui.UsageFunc)

	// Define command groups
	rootCmd.AddGroup(
		&cobra.Group{ID: "devtools", Title: "Developer Tools"},
		&cobra.Group{ID: "envops", Title: "Environment & Ops"},
		&cobra.Group{ID: "mgmt", Title: "Management & Workflows"},
		&cobra.Group{ID: "meta", Title: "Meta"},
	)

	// Set the version string for the root command's --version flag
	rootCmd.Version = version.Version

	// Persistent flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.eng.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")

	// Bind the verbose flag to viper config
	err := viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	cobra.CheckErr(err)

	// Subcommands & Group Assignment
	asdf.AsdfCmd.GroupID = "devtools"
	codemod.CodemodCmd.GroupID = "devtools"
	git.GitCmd.GroupID = "devtools"
	gitlab.GitLabCmd.GroupID = "devtools"
	ts.TailscaleCmd.GroupID = "devtools"

	compose.ComposeCmd.GroupID = "envops"
	dotfiles.DotfilesCmd.GroupID = "envops"
	files.FilesCmd.GroupID = "envops"
	immich.ImmichCmd.GroupID = "envops"
	system.SystemCmd.GroupID = "envops"

	config.ConfigCmd.GroupID = "mgmt"
	project.ProjectCmd.GroupID = "mgmt"
	dashboardCmd.GroupID = "mgmt"

	version.VersionCmd.GroupID = "meta"

	// Add subcommands
	rootCmd.AddCommand(asdf.AsdfCmd)
	rootCmd.AddCommand(codemod.CodemodCmd)
	rootCmd.AddCommand(compose.ComposeCmd)
	rootCmd.AddCommand(config.ConfigCmd)
	rootCmd.AddCommand(dotfiles.DotfilesCmd)
	rootCmd.AddCommand(files.FilesCmd)
	rootCmd.AddCommand(gitlab.GitLabCmd)
	rootCmd.AddCommand(git.GitCmd)
	rootCmd.AddCommand(immich.ImmichCmd)
	rootCmd.AddCommand(project.ProjectCmd)
	rootCmd.AddCommand(system.SystemCmd)
	rootCmd.AddCommand(ts.TailscaleCmd)
	rootCmd.AddCommand(version.VersionCmd)
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
		log.Verbose(cmdutil.IsVerbose(rootCmd), "Using config file: %s", viper.ConfigFileUsed())
	} else if errors.As(err, &viper.ConfigFileNotFoundError{}) {
		// Config file not found, create it
		configFilePath := viper.ConfigFileUsed()
		if configFilePath == "" {
			// Construct the default config file path if not already set by viper
			home, err := os.UserHomeDir()
			cobra.CheckErr(err)
			configFilePath = home + string(os.PathSeparator) + ".eng.yaml"
		}

		if err := viper.SafeWriteConfigAs(configFilePath); err != nil {
			log.Warn("Error creating config file %s: %v", configFilePath, err)
		} else {
			log.Verbose(cmdutil.IsVerbose(rootCmd), "Created new config file: %s", configFilePath)
		}
	} else {
		// Config file was found but another error was produced
		log.Warn("Error reading config file %s: %v", viper.ConfigFileUsed(), err)
	}

	// Run migration to ensure keys are standardized
	configUtils.MigrateConfig()
}
