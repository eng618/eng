package gitlab

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	gitlabcfg "github.com/eng618/eng/internal/utils/config/gitlab"
	"github.com/eng618/eng/internal/utils/log"
)

var (
	initOutputPath string
	initForce      bool
	initYes        bool
)

// mrRulesInitCmd interactively creates a MR rules JSON file.
var mrRulesInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactively generate a MR rules JSON file",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default output
		if initOutputPath == "" {
			initOutputPath = "gitlab-rules.json"
		}
		abs, err := filepath.Abs(initOutputPath)
		if err != nil {
			return err
		}

		// Check existing file
		if _, err := os.Stat(abs); err == nil && !initForce && !initYes {
			if !confirm("File exists. Overwrite?", false) {
				log.Warn("Aborted: will not overwrite %s", abs)
				return nil
			}
		}

		// Build spec
		spec := gitlabcfg.MRRules{}
		if initYes {
			// Use sensible defaults
			spec.SchemaVersion = "1"
			spec.MergeMethod = "ff"
			spec.DeleteSourceBranch = true
			spec.RequireSquash = true
			spec.PipelinesMustSucceed = true
			spec.AllowSkippedAsSuccess = true
			spec.AllThreadsMustResolve = true
		} else {
			log.Start("Answer prompts to generate MR rules")
			spec.SchemaVersion = askString("Schema version", "1")
			spec.MergeMethod = askSelectRadio("Merge method", []string{"ff", "merge_commit", "rebase_merge"}, "ff")
			spec.DeleteSourceBranch = askBoolRadio("Delete source branch by default?", true)
			spec.RequireSquash = askBoolRadio("Require squash when merging?", true)
			spec.PipelinesMustSucceed = askBoolRadio("Pipelines must succeed?", true)
			spec.AllowSkippedAsSuccess = askBoolRadio("Treat skipped pipelines as success?", true)
			spec.AllThreadsMustResolve = askBoolRadio("All threads must be resolved?", true)
		}

		if err := spec.Validate(); err != nil {
			return err
		}

		b, err := json.MarshalIndent(spec, "", "  ")
		if err != nil {
			return err
		}

		if err := os.WriteFile(abs, b, 0o644); err != nil {
			return err
		}
		log.Success("Wrote MR rules to %s", abs)
		return nil
	},
}

func init() {
	mrRulesCmd.AddCommand(mrRulesInitCmd)
	mrRulesInitCmd.Flags().StringVar(&initOutputPath, "output", "", "Output path for rules JSON (default: gitlab-rules.json)")
	mrRulesInitCmd.Flags().BoolVar(&initForce, "force", false, "Overwrite existing file without prompt")
	mrRulesInitCmd.Flags().BoolVar(&initYes, "yes", false, "Use defaults without prompting")
}

var reader = bufio.NewReader(os.Stdin)

func askString(prompt, def string) string {
	log.Message("%s [%s]:", prompt, def)
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	if text == "" {
		return def
	}
	return text
}

func askBoolRadio(prompt string, def bool) bool {
	opts := []string{"Yes", "No"}
	defOpt := opts[0]
	if !def {
		defOpt = opts[1]
	}
	var sel string
	_ = survey.AskOne(&survey.Select{Message: prompt, Options: opts, Default: defOpt}, &sel)
	return sel == "Yes"
}

func askSelectRadio(prompt string, options []string, def string) string {
	var sel string
	_ = survey.AskOne(&survey.Select{Message: prompt, Options: options, Default: def}, &sel)
	if sel == "" {
		return def
	}
	return sel
}

func confirm(prompt string, def bool) bool {
	return askBoolRadio(prompt, def)
}
