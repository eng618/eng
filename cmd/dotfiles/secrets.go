package dotfiles

import (
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/eng618/eng/internal/utils"
	configUtils "github.com/eng618/eng/internal/utils/config"
	"github.com/eng618/eng/internal/utils/log"
)

var (
	dotfilesSecretsManifestPath string
	dotfilesSecretsRootPath     string
	dotfilesSecretsProjectID    string
)

// SecretsCmd manages manifest-driven dotfiles env secrets with Bitwarden Secrets Manager.
var SecretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Backup and restore dotfiles env secrets via Bitwarden Secrets Manager",
	Long:  `Backup and restore managed dotfiles env files using a tracked manifest and Bitwarden Secrets Manager.`,
	RunE: func(cmd *cobra.Command, _args []string) error {
		return cmd.Help()
	},
}

// SecretsBackupCmd saves managed env keys from the dotfiles worktree into bws.
var SecretsBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup managed dotfiles env values into Bitwarden Secrets Manager",
	RunE: func(cmd *cobra.Command, _args []string) error {
		return utils.BackupDotfilesSecrets(dotfilesSecretsOptions(cmd))
	},
}

// SecretsRestoreCmd restores managed env files from tracked templates and bws values.
var SecretsRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore managed dotfiles env files from Bitwarden Secrets Manager",
	RunE: func(cmd *cobra.Command, _args []string) error {
		return utils.RestoreDotfilesSecrets(dotfilesSecretsOptions(cmd))
	},
}

func init() {
	SecretsCmd.PersistentFlags().StringVar(&dotfilesSecretsManifestPath, "manifest", "", "Path to the dotfiles secrets manifest")
	SecretsCmd.PersistentFlags().StringVar(&dotfilesSecretsRootPath, "root", "", "Root path for manifest-relative env files")
	SecretsCmd.PersistentFlags().StringVar(&dotfilesSecretsProjectID, "project-id", "", "Bitwarden Secrets Manager project ID override")

	SecretsCmd.AddCommand(SecretsBackupCmd)
	SecretsCmd.AddCommand(SecretsRestoreCmd)
}

func dotfilesSecretsOptions(cmd *cobra.Command) utils.DotfilesSecretsOptions {
	manifestPath := dotfilesSecretsManifestPath
	if manifestPath == "" {
		manifestPath = filepath.Join(configUtils.WorktreePath(), "bin", "secrets", "server.manifest")
	}

	log.Verbose(utils.IsVerbose(cmd), "Using dotfiles secrets manifest: %s", manifestPath)
	if dotfilesSecretsRootPath != "" {
		log.Verbose(utils.IsVerbose(cmd), "Using dotfiles secrets root override: %s", dotfilesSecretsRootPath)
	}

	return utils.DotfilesSecretsOptions{
		ManifestPath: manifestPath,
		RootPath:     dotfilesSecretsRootPath,
		ProjectID:    dotfilesSecretsProjectID,
		Verbose:      utils.IsVerbose(cmd),
	}
}
