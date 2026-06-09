package gitlab

import (
	"errors"
)

type MRRules struct {
	SchemaVersion string `json:"schemaVersion,omitempty"`
	// MergeMethod accepts: ff | merge_commit | rebase_merge
	MergeMethod string `json:"mergeMethod,omitempty"`
	// DeleteSourceBranch enables default remove source branch
	DeleteSourceBranch bool `json:"deleteSourceBranch,omitempty"`
	// RequireSquash sets squash_option=always
	RequireSquash bool `json:"requireSquash,omitempty"`
	// Checks
	PipelinesMustSucceed  bool `json:"pipelinesMustSucceed,omitempty"`
	AllowSkippedAsSuccess bool `json:"allowSkippedAsSuccess,omitempty"`
	AllThreadsMustResolve bool `json:"allThreadsMustResolve,omitempty"`
}

func (m *MRRules) Validate() error {
	// Normalize MergeMethod
	if m.MergeMethod == "" {
		m.MergeMethod = "ff"
	}
	switch m.MergeMethod {
	case "ff", "merge_commit", "rebase_merge":
		// ok
	default:
		return errors.New("mergeMethod must be one of: ff, merge_commit, rebase_merge")
	}
	return nil
}

// ToAPIFields maps rules to GitLab project API fields.
func (m *MRRules) ToAPIFields() map[string]any {
	fields := map[string]any{
		"merge_method": m.MergeMethod,
	}
	if m.DeleteSourceBranch {
		fields["remove_source_branch_after_merge"] = true
	}
	if m.RequireSquash {
		fields["squash_option"] = "always"
	}
	if m.PipelinesMustSucceed {
		fields["only_allow_merge_if_pipeline_succeeds"] = true
	}
	// allow skipped considered success
	fields["allow_merge_on_skipped_pipeline"] = m.AllowSkippedAsSuccess
	if m.AllThreadsMustResolve {
		fields["only_allow_merge_if_all_discussions_are_resolved"] = true
	}
	return fields
}
