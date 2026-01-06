package gitlab

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/utils"
	gitlabcfg "github.com/eng618/eng/utils/config/gitlab"
	"github.com/eng618/eng/utils/log"
	gitrepo "github.com/eng618/eng/utils/repo"
)

var (
	rulesPath    string
	projectOpt   string
	hostOpt      string
	dryRun       bool
	tokenItemOpt string
)

// mrRulesApplyCmd applies MR rules to a GitLab project using glab api.
var mrRulesApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply merge request rules from a JSON file",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load and validate rules file
		if rulesPath == "" {
			return fmt.Errorf("--rules is required")
		}
		abs, err := filepath.Abs(rulesPath)
		if err != nil {
			return err
		}
		b, err := os.ReadFile(abs)
		if err != nil {
			return err
		}
		var spec gitlabcfg.MRRules
		if err := json.Unmarshal(b, &spec); err != nil {
			return fmt.Errorf("failed to parse rules json: %w", err)
		}
		if err := spec.Validate(); err != nil {
			return err
		}

		// Resolve host and project
		host := hostOpt
		project := projectOpt
		if host == "" || project == "" {
			if h, p, err := gitrepo.GetGitLabHostAndProjectPath("."); err == nil {
				if host == "" {
					host = h
				}
				if project == "" {
					project = p
				}
			}
		}
		// Optional: read from config if still unset
		if host == "" {
			host = viper.GetString("gitlab.host")
		}
		if project == "" {
			project = viper.GetString("gitlab.project")
		}
		if project == "" {
			return errors.New("could not determine GitLab project; use --project or run inside a Git repo")
		}

		// Prepare token via env or Bitwarden
		env := os.Environ()
		if os.Getenv("GITLAB_TOKEN") == "" {
			// Prefer Bitwarden item reference if configured
			itemName := tokenItemOpt
			if itemName == "" {
				itemName = viper.GetString("gitlab.tokenItem")
			}
			if itemName != "" {
				// Ensure BW session and fetch item password as token
				sess, err := utils.EnsureBitwardenSession()
				if err != nil {
					return err
				}
				if sess != "" {
					env = append(env, "BW_SESSION="+sess)
				}
				item, err := utils.GetBitwardenItem(itemName)
				if err != nil {
					return fmt.Errorf("failed to read Bitwarden item '%s': %w", itemName, err)
				}
				var token string
				if item.Login != nil && item.Login.Password != "" {
					token = item.Login.Password
				}
				for _, f := range item.Fields {
					if f.Name == "token" && f.Value != "" {
						token = f.Value
						break
					}
				}
				if token != "" {
					env = append(env, "GITLAB_TOKEN="+token)
				}
			}
			// Lastly, fall back to config literal (not recommended)
			if os.Getenv("GITLAB_TOKEN") == "" && viper.GetString("gitlab.token") != "" {
				env = append(env, "GITLAB_TOKEN="+viper.GetString("gitlab.token"))
			}
		}

		if host != "" {
			env = append(env, "GITLAB_HOST="+host)
		}

		// Build glab api call
		apiPath := fmt.Sprintf("projects/%s", url.PathEscape(project))
		apiArgs := []string{"api", apiPath, "-X", "PUT"}
		for k, v := range spec.ToAPIFields() {
			apiArgs = append(apiArgs, "-F", fmt.Sprintf("%s=%v", k, v))
		}

		log.Info("Applying MR rules to %s on host %s", project, host)
		if dryRun {
			log.Message("dry-run: glab %v", apiArgs)
			return nil
		}

		cmdExec := exec.Command("glab", apiArgs...)
		cmdExec.Env = env
		return utils.StartChildProcess(cmdExec)
	},
}

func init() {
	mrRulesCmd.AddCommand(mrRulesApplyCmd)
	mrRulesApplyCmd.Flags().StringVar(&rulesPath, "rules", "", "Path to MR rules JSON file")
	mrRulesApplyCmd.Flags().StringVar(&projectOpt, "project", "", "GitLab project path (e.g., group/subgroup/repo)")
	mrRulesApplyCmd.Flags().StringVar(&hostOpt, "host", "", "GitLab host (e.g., gitlab.com)")
	// Support the user's typo as an alias for convenience
	mrRulesApplyCmd.Flags().StringVar(&hostOpt, "hose", "", "Alias for --host")
	mrRulesApplyCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the glab command without executing")
	mrRulesApplyCmd.Flags().StringVar(&tokenItemOpt, "token-item", "", "Bitwarden item name containing a GitLab token")
	_ = mrRulesApplyCmd.MarkFlagRequired("rules")
}
