package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/eng618/eng/internal/config"
	"github.com/eng618/eng/internal/log"
	gitrepo "github.com/eng618/eng/internal/repo"
	"github.com/eng618/eng/internal/ui"
	"github.com/eng618/eng/internal/ui/theme"
)

// doctorCmd validates glab availability, token validity, and project access.
var (
	docHostOpt    string
	docProjectOpt string
	docQuiet      bool
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Validate GitLab token and project access",
	RunE: func(cmd *cobra.Command, args []string) error {
		headerStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Primary).
			MarginBottom(1)
		if !ui.DisableProgress {
			fmt.Fprintln(log.Out, headerStyle.Render("🩺 GitLab Auth Diagnostic Doctor"))
		}

		// Check glab is installed
		if _, err := lookPath("glab"); err != nil {
			return theme.NewActionableError(
				errors.New("glab CLI not found in PATH"),
				"Install the glab CLI from https://gitlab.com/gitlab-org/cli to enable this command.",
			)
		}

		// Resolve host and project similar to mr-rules apply
		host := docHostOpt
		project := docProjectOpt
		if host == "" {
			host = viper.GetString("gitlab.host")
		}
		if project == "" {
			project = viper.GetString("gitlab.project")
		}
		if host == "" || project == "" {
			if h, p, err := gitrepo.GetGitLabHostAndProjectPath(cmd.Context(), "."); err == nil {
				if host == "" {
					host = h
				}
				if project == "" {
					project = p
				}
			}
		}

		// Prepare environment with token via the central secure store
		// (env -> Bitwarden -> keychain -> config).
		env := os.Environ()
		// If token not already in process env, resolve via the secure store.
		if os.Getenv("GITLAB_TOKEN") == "" {
			if token, _, err := config.ResolveGitLabToken(); err == nil {
				env = append(env, "GITLAB_TOKEN="+token)
			} else {
				log.Warn("GitLab token not available: %v", err)
			}
		}
		if host != "" {
			env = append(env, "GITLAB_HOST="+host)
		}

		// 1) Validate token by calling /user
		{
			cmdUser := execCommand("glab", "api", "user")
			cmdUser.Env = env
			out, err := cmdUser.Output()
			if err != nil {
				return fmt.Errorf(
					"failed to call glab api user (ensure GITLAB_TOKEN is set or configured via Bitwarden/config): %w",
					err,
				)
			}
			var user struct {
				Username string `json:"username"`
				Name     string `json:"name"`
			}
			if err := json.Unmarshal(out, &user); err != nil {
				return fmt.Errorf("failed to parse user response: %w", err)
			}
			if user.Username == "" {
				return errors.New("token validation returned empty username")
			}
			if docQuiet {
				log.Message("OK token %s", user.Username)
			} else {
				log.Success("Token valid for user: %s (%s)", user.Username, user.Name)
			}
		}

		// 2) If project resolvable, ensure access by GET /projects/:id
		if project != "" {
			cmdProj := execCommand("glab", "api", fmt.Sprintf("projects/%s", project))
			cmdProj.Env = env
			out, err := cmdProj.Output()
			if err != nil {
				if !docQuiet {
					log.Warn("Unable to access project %s: %v", project, err)
				}
			} else {
				var p struct {
					PathWithNamespace string `json:"path_with_namespace"`
				}
				_ = json.Unmarshal(out, &p)
				if p.PathWithNamespace != "" {
					if docQuiet {
						log.Message("OK project %s", p.PathWithNamespace)
					} else {
						log.Success("Project access OK: %s", p.PathWithNamespace)
					}
				} else {
					if !docQuiet {
						log.Info("Project access check returned without path; access may be limited")
					}
				}
			}
		} else {
			if !docQuiet {
				log.Warn("No project detected; set gitlab.project in config or run inside a repo")
			}
		}

		if !docQuiet {
			log.Message("Doctor checks completed")
		}
		return nil
	},
}

func init() {
	AuthCmd.AddCommand(doctorCmd)
	doctorCmd.Flags().StringVar(&docHostOpt, "host", "", "GitLab host (e.g., gitlab.com)")
	doctorCmd.Flags().StringVar(&docProjectOpt, "project", "", "GitLab project path (e.g., group/subgroup/repo)")
	doctorCmd.Flags().BoolVar(&docQuiet, "quiet", false, "Only print essential OK/error messages")
}
