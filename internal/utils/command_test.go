// Package utils_test contains unit tests for the utils package.
package utils_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/utils"
)

func TestIsVerbose_FlagSet(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("verbose", false, "verbose output")
	_ = cmd.Flags().Set("verbose", "true")

	if !utils.IsVerbose(cmd) {
		t.Error("IsVerbose should return true when flag is set to true")
	}
}

func TestIsVerbose_FlagSetFalse(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("verbose", true, "verbose output") // Default true
	_ = cmd.Flags().Set("verbose", "false")             // Explicitly set to false

	if utils.IsVerbose(cmd) {
		t.Error("IsVerbose should return false when flag is set to false")
	}
}

func TestIsVerbose_FlagNotSet_UsesViper(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("verbose", false, "verbose output")
	viper.Set("verbose", true)
	defer viper.Reset()

	if !utils.IsVerbose(cmd) {
		t.Error("IsVerbose should return true when viper config is true and flag is not set")
	}

	viper.Set("verbose", false)
	if utils.IsVerbose(cmd) {
		t.Error("IsVerbose should return false when viper config is false and flag is not set")
	}
}

func TestIsVerbose_FlagDefault_UsesViper(t *testing.T) {
	cmd := &cobra.Command{}
	// Flag defined with a default value, but not explicitly changed via Set()
	cmd.Flags().Bool("verbose", true, "verbose output") // Default true
	viper.Set("verbose", false)                         // Viper has a different value
	defer viper.Reset()

	// Since Changed("verbose") is false, it should use Viper's value
	if utils.IsVerbose(cmd) {
		t.Error("IsVerbose should return false (viper value) when flag uses default and is not changed")
	}

	viper.Set("verbose", true)
	if !utils.IsVerbose(cmd) {
		t.Error("IsVerbose should return true (viper value) when flag uses default and is not changed")
	}
}
