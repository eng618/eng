package repo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepoNameFromURL(t *testing.T) {
	testCases := []struct {
		name        string
		url         string
		expected    string
		expectError bool
	}{
		{
			name:     "SSH format with .git",
			url:      "git@github.com:org/my-repo.git",
			expected: "my-repo",
		},
		{
			name:     "SSH format without .git",
			url:      "git@github.com:org/my-repo",
			expected: "my-repo",
		},
		{
			name:     "SSH format with nested path",
			url:      "git@gitlab.com:group/subgroup/my-repo.git",
			expected: "my-repo",
		},
		{
			name:     "SSH format with protocol",
			url:      "ssh://git@github.com/org/my-repo.git",
			expected: "my-repo",
		},
		{
			name:     "HTTPS format with .git",
			url:      "https://github.com/org/my-repo.git",
			expected: "my-repo",
		},
		{
			name:     "HTTPS format without .git",
			url:      "https://github.com/org/my-repo",
			expected: "my-repo",
		},
		{
			name:     "HTTPS format with nested path",
			url:      "https://gitlab.com/group/subgroup/my-repo.git",
			expected: "my-repo",
		},
		{
			name:     "HTTP format",
			url:      "http://github.com/org/my-repo.git",
			expected: "my-repo",
		},
		{
			name:        "Invalid format",
			url:         "not-a-valid-url",
			expectError: true,
		},
		{
			name:        "Empty string",
			url:         "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := RepoNameFromURL(tc.url)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestEffectivePath(t *testing.T) {
	got, err := EffectivePath("git@github.com:org/my-repo.git", "")
	assert.NoError(t, err)
	assert.Equal(t, "my-repo", got)

	got, err = EffectivePath("git@github.com:org/my-repo.git", "custom-name")
	assert.NoError(t, err)
	assert.Equal(t, "custom-name", got)

	_, err = EffectivePath("not-a-valid-url", "")
	assert.Error(t, err)
}
