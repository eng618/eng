package gitlab

import (
	"fmt"
	"testing"
)

// ============================================================================
// T - Tests (Unit and Table-driven Tests)
// ============================================================================

func TestMRRules_Validate(t *testing.T) {
	tests := []struct {
		name           string
		rules          MRRules
		expectedMethod string
		expectErr      bool
	}{
		{
			name: "Valid default ff",
			rules: MRRules{
				MergeMethod: "",
			},
			expectedMethod: "ff",
			expectErr:      false,
		},
		{
			name: "Valid ff",
			rules: MRRules{
				MergeMethod: "ff",
			},
			expectedMethod: "ff",
			expectErr:      false,
		},
		{
			name: "Valid merge_commit",
			rules: MRRules{
				MergeMethod: "merge_commit",
			},
			expectedMethod: "merge_commit",
			expectErr:      false,
		},
		{
			name: "Valid rebase_merge",
			rules: MRRules{
				MergeMethod: "rebase_merge",
			},
			expectedMethod: "rebase_merge",
			expectErr:      false,
		},
		{
			name: "Invalid method",
			rules: MRRules{
				MergeMethod: "invalid_method",
			},
			expectedMethod: "invalid_method",
			expectErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.rules
			err := r.Validate()

			if tt.expectErr {
				if err == nil {
					t.Errorf("expected validation error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}

			if r.MergeMethod != tt.expectedMethod {
				t.Errorf("expected MergeMethod to be %q, got %q", tt.expectedMethod, r.MergeMethod)
			}
		})
	}
}

func TestMRRules_ToAPIFields(t *testing.T) {
	tests := []struct {
		name         string
		rules        MRRules
		expectFields map[string]any
	}{
		{
			name: "Only default fields",
			rules: MRRules{
				MergeMethod: "ff",
			},
			expectFields: map[string]any{
				"merge_method":                    "ff",
				"allow_merge_on_skipped_pipeline": false,
			},
		},
		{
			name: "All fields enabled",
			rules: MRRules{
				MergeMethod:           "rebase_merge",
				DeleteSourceBranch:    true,
				RequireSquash:         true,
				PipelinesMustSucceed:  true,
				AllowSkippedAsSuccess: true,
				AllThreadsMustResolve: true,
			},
			expectFields: map[string]any{
				"merge_method":                                     "rebase_merge",
				"remove_source_branch_after_merge":                 true,
				"squash_option":                                    "always",
				"only_allow_merge_if_pipeline_succeeds":            true,
				"allow_merge_on_skipped_pipeline":                  true,
				"only_allow_merge_if_all_discussions_are_resolved": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := tt.rules.ToAPIFields()

			// Check sizes and matches
			for key, val := range tt.expectFields {
				got, exists := fields[key]
				if !exists {
					t.Errorf("expected API field %s to exist, but it was missing", key)
					continue
				}
				if got != val {
					t.Errorf("expected key %s to have value %v, got %v", key, val, got)
				}
			}

			// Ensure extra keys are not present
			for key := range fields {
				if _, exists := tt.expectFields[key]; !exists {
					t.Errorf("unexpected API field key %s returned", key)
				}
			}
		})
	}
}

// ============================================================================
// E - Examples (Executable Documentation)
// ============================================================================

func ExampleMRRules() {
	rules := MRRules{
		MergeMethod:          "merge_commit",
		PipelinesMustSucceed: true,
	}

	// Validate rule structure
	if err := rules.Validate(); err == nil {
		fields := rules.ToAPIFields()
		// Print mapped properties for API
		fmt.Println("Merge Method:", fields["merge_method"].(string))
		fmt.Println("Pipeline Succeeds Required:", fields["only_allow_merge_if_pipeline_succeeds"].(bool))
	}
	// Output:
	// Merge Method: merge_commit
	// Pipeline Succeeds Required: true
}

// ============================================================================
// B - Benchmarks (Performance Profiling)
// ============================================================================

func BenchmarkMRRules_Validate(b *testing.B) {
	rules := MRRules{
		MergeMethod: "ff",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rules.Validate()
	}
}

func BenchmarkMRRules_ToAPIFields(b *testing.B) {
	rules := MRRules{
		MergeMethod:           "rebase_merge",
		DeleteSourceBranch:    true,
		RequireSquash:         true,
		PipelinesMustSucceed:  true,
		AllowSkippedAsSuccess: true,
		AllThreadsMustResolve: true,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rules.ToAPIFields()
	}
}
