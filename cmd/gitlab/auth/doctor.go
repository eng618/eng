package auth

import (
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "os/exec"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"

    "github.com/eng618/eng/utils"
    "github.com/eng618/eng/utils/log"
    gitrepo "github.com/eng618/eng/utils/repo"
)

// doctorCmd validates glab availability, token validity, and project access
var (
    docHostOpt    string
    docProjectOpt string
    docQuiet      bool
)

var doctorCmd = &cobra.Command{
    Use:   "doctor",
    Short: "Validate GitLab token and project access",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Check glab is installed
        if _, err := exec.LookPath("glab"); err != nil {
            return fmt.Errorf("glab CLI not found in PATH; install from https://gitlab.com/gitlab-org/cli")
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
            if h, p, err := gitrepo.GetGitLabHostAndProjectPath("."); err == nil {
                if host == "" {
                    host = h
                }
                if project == "" {
                    project = p
                }
            }
        }

        // Prepare environment with token (env -> Bitwarden -> config)
        env := os.Environ()
        // If token not already in process env, try Bitwarden then config
        if os.Getenv("GITLAB_TOKEN") == "" {
            itemName := viper.GetString("gitlab.tokenItem")
            if itemName != "" {
                sess, err := utils.EnsureBitwardenSession()
                if err != nil {
                    log.Warn("Bitwarden session not available: %v", err)
                } else if sess != "" {
                    env = append(env, "BW_SESSION="+sess)
                    if item, err := utils.GetBitwardenItem(itemName); err == nil {
                        // prefer login.password; fallback to field named token
                        if item.Login != nil && item.Login.Password != "" {
                            env = append(env, "GITLAB_TOKEN="+item.Login.Password)
                        } else {
                            for _, f := range item.Fields {
                                if f.Name == "token" && f.Value != "" {
                                    env = append(env, "GITLAB_TOKEN="+f.Value)
                                    break
                                }
                            }
                        }
                    }
                }
            }
            if os.Getenv("GITLAB_TOKEN") == "" && viper.GetString("gitlab.token") != "" {
                env = append(env, "GITLAB_TOKEN="+viper.GetString("gitlab.token"))
            }
        }
        if host != "" {
            env = append(env, "GITLAB_HOST="+host)
        }

        // 1) Validate token by calling /user
        {
            cmdUser := exec.Command("glab", "api", "user")
            cmdUser.Env = env
            out, err := cmdUser.Output()
            if err != nil {
                return fmt.Errorf("failed to call glab api user: %w\nEnsure GITLAB_TOKEN is set or configured via Bitwarden/config.", err)
            }
            var user struct{ Username string `json:"username"`; Name string `json:"name"` }
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
            cmdProj := exec.Command("glab", "api", fmt.Sprintf("projects/%s", project))
            cmdProj.Env = env
            out, err := cmdProj.Output()
            if err != nil {
                if !docQuiet {
                    log.Warn("Unable to access project %s: %v", project, err)
                }
            } else {
                var p struct{ PathWithNamespace string `json:"path_with_namespace"` }
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
    doctorCmd.Flags().StringVar(&docHostOpt, "hose", "", "Alias for --host")
    doctorCmd.Flags().StringVar(&docProjectOpt, "project", "", "GitLab project path (e.g., group/subgroup/repo)")
    doctorCmd.Flags().BoolVar(&docQuiet, "quiet", false, "Only print essential OK/error messages")
}
