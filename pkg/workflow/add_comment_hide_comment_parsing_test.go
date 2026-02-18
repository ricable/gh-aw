//go:build !integration

package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseHideCommentConfig tests that all fields are properly parsed
func TestParseHideCommentConfig(t *testing.T) {
	tests := []struct {
		name     string
		inputMap map[string]any
		expected *HideCommentConfig
		isNil    bool
	}{
		{
			name: "all fields parsed correctly",
			inputMap: map[string]any{
				"hide-comment": map[string]any{
					"max":             10,
					"target-repo":     "owner/repo",
					"discussion":      true,
					"allowed-reasons": []any{"spam", "outdated"},
				},
			},
			expected: &HideCommentConfig{
				BaseSafeOutputConfig: BaseSafeOutputConfig{
					Max: 10,
				},
				SafeOutputTargetConfig: SafeOutputTargetConfig{
					TargetRepoSlug: "owner/repo",
				},
				Discussion:     ptrBool(true),
				AllowedReasons: []string{"spam", "outdated"},
			},
		},
		{
			name: "discussion false is preserved",
			inputMap: map[string]any{
				"hide-comment": map[string]any{
					"max":        5,
					"discussion": false,
				},
			},
			expected: &HideCommentConfig{
				BaseSafeOutputConfig: BaseSafeOutputConfig{
					Max: 5,
				},
				Discussion: ptrBool(false),
			},
		},
		{
			name: "no discussion field defaults to nil",
			inputMap: map[string]any{
				"hide-comment": map[string]any{
					"max": 3,
				},
			},
			expected: &HideCommentConfig{
				BaseSafeOutputConfig: BaseSafeOutputConfig{
					Max: 3,
				},
				Discussion: nil,
			},
		},
		{
			name: "default max of 5 when not specified",
			inputMap: map[string]any{
				"hide-comment": map[string]any{},
			},
			expected: &HideCommentConfig{
				BaseSafeOutputConfig: BaseSafeOutputConfig{
					Max: 5,
				},
			},
		},
		{
			name: "nil config sets default max",
			inputMap: map[string]any{
				"hide-comment": nil,
			},
			expected: &HideCommentConfig{
				BaseSafeOutputConfig: BaseSafeOutputConfig{
					Max: 5,
				},
			},
		},
		{
			name: "wildcard target-repo returns nil",
			inputMap: map[string]any{
				"hide-comment": map[string]any{
					"target-repo": "*",
				},
			},
			isNil: true,
		},
		{
			name: "allowed-repos is parsed",
			inputMap: map[string]any{
				"hide-comment": map[string]any{
					"allowed-repos": []any{"owner/repo1", "owner/repo2"},
				},
			},
			expected: &HideCommentConfig{
				BaseSafeOutputConfig: BaseSafeOutputConfig{
					Max: 5,
				},
				SafeOutputTargetConfig: SafeOutputTargetConfig{
					AllowedRepos: []string{"owner/repo1", "owner/repo2"},
				},
			},
		},
		{
			name:     "missing hide-comment key returns nil",
			inputMap: map[string]any{},
			isNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := &Compiler{}
			result := compiler.parseHideCommentConfig(tt.inputMap)

			if tt.isNil {
				assert.Nil(t, result, "Expected nil config")
				return
			}

			require.NotNil(t, result, "Expected non-nil config")

			assert.Equal(t, tt.expected.Max, result.Max, "Max should match")
			assert.Equal(t, tt.expected.TargetRepoSlug, result.TargetRepoSlug, "TargetRepoSlug should match")

			if tt.expected.Discussion != nil {
				require.NotNil(t, result.Discussion, "Discussion should not be nil")
				assert.Equal(t, *tt.expected.Discussion, *result.Discussion, "Discussion value should match")
			} else {
				assert.Nil(t, result.Discussion, "Discussion should be nil")
			}

			assert.Equal(t, tt.expected.AllowedReasons, result.AllowedReasons, "AllowedReasons should match")
			assert.Equal(t, tt.expected.AllowedRepos, result.AllowedRepos, "AllowedRepos should match")
		})
	}
}

// TestParseAddCommentsConfig tests that all fields are properly parsed
func TestParseAddCommentsConfig(t *testing.T) {
	tests := []struct {
		name     string
		inputMap map[string]any
		expected *AddCommentsConfig
		isNil    bool
	}{
		{
			name: "all fields parsed correctly",
			inputMap: map[string]any{
				"add-comment": map[string]any{
					"max":                  3,
					"target":               "*",
					"target-repo":          "owner/repo",
					"allowed-repos":        []any{"owner/repo1", "owner/repo2"},
					"discussion":           true,
					"hide-older-comments":  true,
					"allowed-reasons":      []any{"spam", "resolved"},
				},
			},
			expected: &AddCommentsConfig{
				BaseSafeOutputConfig: BaseSafeOutputConfig{
					Max: 3,
				},
				Target:            "*",
				TargetRepoSlug:    "owner/repo",
				AllowedRepos:      []string{"owner/repo1", "owner/repo2"},
				Discussion:        ptrBool(true),
				HideOlderComments: true,
				AllowedReasons:    []string{"spam", "resolved"},
			},
		},
		{
			name: "discussion false is preserved",
			inputMap: map[string]any{
				"add-comment": map[string]any{
					"max":        1,
					"discussion": false,
				},
			},
			expected: &AddCommentsConfig{
				BaseSafeOutputConfig: BaseSafeOutputConfig{
					Max: 1,
				},
				Discussion: ptrBool(false),
			},
		},
		{
			name: "no discussion field defaults to nil",
			inputMap: map[string]any{
				"add-comment": map[string]any{
					"max": 2,
				},
			},
			expected: &AddCommentsConfig{
				BaseSafeOutputConfig: BaseSafeOutputConfig{
					Max: 2,
				},
				Discussion: nil,
			},
		},
		{
			name: "default max of 1 when not specified",
			inputMap: map[string]any{
				"add-comment": map[string]any{},
			},
			expected: &AddCommentsConfig{
				BaseSafeOutputConfig: BaseSafeOutputConfig{
					Max: 1,
				},
			},
		},
		{
			name: "nil config sets default max",
			inputMap: map[string]any{
				"add-comment": nil,
			},
			expected: &AddCommentsConfig{
				BaseSafeOutputConfig: BaseSafeOutputConfig{
					Max: 1,
				},
			},
		},
		{
			name: "wildcard target-repo returns nil",
			inputMap: map[string]any{
				"add-comment": map[string]any{
					"target-repo": "*",
				},
			},
			isNil: true,
		},
		{
			name:     "missing add-comment key returns nil",
			inputMap: map[string]any{},
			isNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := &Compiler{}
			result := compiler.parseCommentsConfig(tt.inputMap)

			if tt.isNil {
				assert.Nil(t, result, "Expected nil config")
				return
			}

			require.NotNil(t, result, "Expected non-nil config")

			assert.Equal(t, tt.expected.Max, result.Max, "Max should match")
			assert.Equal(t, tt.expected.Target, result.Target, "Target should match")
			assert.Equal(t, tt.expected.TargetRepoSlug, result.TargetRepoSlug, "TargetRepoSlug should match")
			assert.Equal(t, tt.expected.AllowedRepos, result.AllowedRepos, "AllowedRepos should match")
			assert.Equal(t, tt.expected.HideOlderComments, result.HideOlderComments, "HideOlderComments should match")
			assert.Equal(t, tt.expected.AllowedReasons, result.AllowedReasons, "AllowedReasons should match")

			if tt.expected.Discussion != nil {
				require.NotNil(t, result.Discussion, "Discussion should not be nil")
				assert.Equal(t, *tt.expected.Discussion, *result.Discussion, "Discussion value should match")
			} else {
				assert.Nil(t, result.Discussion, "Discussion should be nil")
			}
		})
	}
}
