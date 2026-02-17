//go:build !integration

package workflow

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromptVersionString(t *testing.T) {
	version := PromptVersion("2026-02-17")
	assert.Equal(t, "2026-02-17", version.String(), "String() should return the version string")
}

func TestPromptVersionIsValid(t *testing.T) {
	tests := []struct {
		name     string
		version  PromptVersion
		expected bool
	}{
		{
			name:     "valid date version",
			version:  PromptVersion("2026-02-17"),
			expected: true,
		},
		{
			name:     "empty version",
			version:  PromptVersion(""),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.version.IsValid()
			assert.Equal(t, tt.expected, result, "IsValid() should return expected value")
		})
	}
}

func TestNewPromptVersionManifest(t *testing.T) {
	manifest := NewPromptVersionManifest()
	
	require.NotNil(t, manifest, "Manifest should not be nil")
	assert.NotZero(t, manifest.GeneratedAt, "GeneratedAt should be set")
	assert.NotEmpty(t, manifest.SystemPrompts, "SystemPrompts map should not be empty")
	
	// Check that all expected prompt files have versions
	expectedPrompts := []string{
		"xpia.md",
		"temp_folder_prompt.md",
		"markdown.md",
		"playwright_prompt.md",
		"pr_context_prompt.md",
		"cache_memory_prompt.md",
		"cache_memory_prompt_multi.md",
		"github_context_prompt.md",
		"threat_detection.md",
	}
	
	for _, prompt := range expectedPrompts {
		version, ok := manifest.GetVersion(prompt)
		assert.True(t, ok, "Should have version for %s", prompt)
		assert.True(t, version.IsValid(), "Version for %s should be valid", prompt)
	}
}

func TestPromptVersionManifestGetVersion(t *testing.T) {
	manifest := NewPromptVersionManifest()
	
	t.Run("existing prompt", func(t *testing.T) {
		version, ok := manifest.GetVersion("xpia.md")
		assert.True(t, ok, "Should find xpia.md")
		assert.Equal(t, XPIAPromptVersion, version, "Should return correct version")
	})
	
	t.Run("non-existent prompt", func(t *testing.T) {
		version, ok := manifest.GetVersion("nonexistent.md")
		assert.False(t, ok, "Should not find nonexistent.md")
		assert.Equal(t, PromptVersion(""), version, "Should return empty version")
	})
}

func TestPromptVersionManifestToYAMLComment(t *testing.T) {
	manifest := NewPromptVersionManifest()
	manifest.GeneratedAt = time.Date(2026, 2, 17, 12, 0, 0, 0, time.UTC)
	manifest.CreatorPromptHash = "abc123def456"
	
	comment := manifest.ToYAMLComment()
	
	// Check that comment contains expected content
	assert.Contains(t, comment, "# System Prompt Versions:", "Should have header")
	assert.Contains(t, comment, "# Generated: 2026-02-17T12:00:00Z", "Should have timestamp")
	assert.Contains(t, comment, "xpia.md:", "Should include xpia.md version")
	assert.Contains(t, comment, "temp_folder_prompt.md:", "Should include temp_folder_prompt.md version")
	assert.Contains(t, comment, "# Creator Prompt Hash: abc123def456", "Should include creator prompt hash")
	
	// Check that it starts with # (comment character)
	lines := strings.Split(comment, "\n")
	for _, line := range lines {
		if line != "" {
			assert.True(t, strings.HasPrefix(line, "#"), "Each non-empty line should start with #: %s", line)
		}
	}
}

func TestPromptVersionConstants(t *testing.T) {
	// Verify all version constants are valid
	versions := map[string]PromptVersion{
		"XPIAPromptVersion":             XPIAPromptVersion,
		"TempFolderPromptVersion":       TempFolderPromptVersion,
		"MarkdownPromptVersion":         MarkdownPromptVersion,
		"PlaywrightPromptVersion":       PlaywrightPromptVersion,
		"PRContextPromptVersion":        PRContextPromptVersion,
		"CacheMemoryPromptVersion":      CacheMemoryPromptVersion,
		"CacheMemoryPromptMultiVersion": CacheMemoryPromptMultiVersion,
		"GitHubContextPromptVersion":    GitHubContextPromptVersion,
		"ThreatDetectionPromptVersion":  ThreatDetectionPromptVersion,
	}
	
	for name, version := range versions {
		assert.True(t, version.IsValid(), "%s should be valid", name)
		assert.NotEmpty(t, version.String(), "%s should have a non-empty string value", name)
	}
}
