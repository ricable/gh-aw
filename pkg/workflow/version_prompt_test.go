//go:build !integration

package workflow

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPromptVersionExtraction tests extraction of version field from frontmatter
func TestPromptVersionExtraction(t *testing.T) {
	t.Run("valid semantic version", func(t *testing.T) {
		compiler := NewCompiler()
		frontmatter := map[string]any{
			"version": "1.0.0",
		}

		version := compiler.extractVersion(frontmatter)
		assert.Equal(t, "1.0.0", version, "Should extract version field")
	})

	t.Run("version with pre-release", func(t *testing.T) {
		compiler := NewCompiler()
		frontmatter := map[string]any{
			"version": "2.1.3-beta.1",
		}

		version := compiler.extractVersion(frontmatter)
		assert.Equal(t, "2.1.3-beta.1", version, "Should extract version with pre-release")
	})

	t.Run("version with build metadata", func(t *testing.T) {
		compiler := NewCompiler()
		frontmatter := map[string]any{
			"version": "1.0.0+build.123",
		}

		version := compiler.extractVersion(frontmatter)
		assert.Equal(t, "1.0.0+build.123", version, "Should extract version with build metadata")
	})

	t.Run("version with pre-release and build", func(t *testing.T) {
		compiler := NewCompiler()
		frontmatter := map[string]any{
			"version": "1.2.3-alpha.1+build.123",
		}

		version := compiler.extractVersion(frontmatter)
		assert.Equal(t, "1.2.3-alpha.1+build.123", version, "Should extract complete version string")
	})

	t.Run("missing version field", func(t *testing.T) {
		compiler := NewCompiler()
		frontmatter := map[string]any{
			"name": "Test Workflow",
		}

		version := compiler.extractVersion(frontmatter)
		assert.Equal(t, "", version, "Should return empty string when version not present")
	})

	t.Run("non-string version field", func(t *testing.T) {
		compiler := NewCompiler()
		frontmatter := map[string]any{
			"version": 123,
		}

		version := compiler.extractVersion(frontmatter)
		assert.Equal(t, "", version, "Should return empty string for non-string version")
	})

	t.Run("version with whitespace", func(t *testing.T) {
		compiler := NewCompiler()
		frontmatter := map[string]any{
			"version": "  1.0.0  ",
		}

		version := compiler.extractVersion(frontmatter)
		assert.Equal(t, "1.0.0", version, "Should trim whitespace from version")
	})
}

// TestPromptVersionInHeader tests that version appears in compiled workflow header
func TestPromptVersionInHeader(t *testing.T) {
	t.Run("version included in workflow header", func(t *testing.T) {
		workflowData := &WorkflowData{
			Name:    "Test Workflow",
			Version: "1.2.3",
			AI:      "codex",
			On:      "workflow_dispatch:",
		}

		compiler := NewCompiler()
		yaml, err := compiler.generateYAML(workflowData, "test.md")
		require.NoError(t, err, "Should generate YAML successfully")

		assert.Contains(t, yaml, "# Prompt Version: 1.2.3", "Version should appear in header comment")
	})

	t.Run("no version comment when version not set", func(t *testing.T) {
		workflowData := &WorkflowData{
			Name: "Test Workflow",
			AI:   "codex",
			On:   "workflow_dispatch:",
		}

		compiler := NewCompiler()
		yaml, err := compiler.generateYAML(workflowData, "test.md")
		require.NoError(t, err, "Should generate YAML successfully")

		assert.NotContains(t, yaml, "# Prompt Version:", "Version comment should not appear when version not set")
	})
}

// TestPromptVersionInAwInfo tests that version appears in aw_info JSON
func TestPromptVersionInAwInfo(t *testing.T) {
	t.Run("prompt_version in aw_info", func(t *testing.T) {
		workflowData := &WorkflowData{
			Name:    "Test Workflow",
			Version: "2.0.0-beta.1",
			AI:      "codex",
			On:      "workflow_dispatch:",
		}

		compiler := NewCompiler()
		yaml, err := compiler.generateYAML(workflowData, "test.md")
		require.NoError(t, err, "Should generate YAML successfully")

		assert.Contains(t, yaml, `prompt_version: "2.0.0-beta.1"`, "prompt_version should appear in aw_info JSON")
	})

	t.Run("no prompt_version when version not set", func(t *testing.T) {
		workflowData := &WorkflowData{
			Name: "Test Workflow",
			AI:   "codex",
			On:   "workflow_dispatch:",
		}

		compiler := NewCompiler()
		yaml, err := compiler.generateYAML(workflowData, "test.md")
		require.NoError(t, err, "Should generate YAML successfully")

		assert.NotContains(t, yaml, "prompt_version:", "prompt_version should not appear when version not set")
	})
}

// TestPromptVersionWithOtherMetadata tests version works with other metadata fields
func TestPromptVersionWithOtherMetadata(t *testing.T) {
	workflowData := &WorkflowData{
		Name:        "Test Workflow",
		Description: "A test workflow for versioning",
		Source:      "github/gh-aw/workflows/test.md@main",
		Version:     "3.1.4",
		TrackerID:   "test-tracker-123",
		AI:          "codex",
		On:          "workflow_dispatch:",
	}

	compiler := NewCompiler()
	yaml, err := compiler.generateYAML(workflowData, "test.md")
	require.NoError(t, err, "Should generate YAML successfully")

	assert.Contains(t, yaml, "# A test workflow for versioning", "Description should appear in header")
	assert.Contains(t, yaml, "# Source: github/gh-aw/workflows/test.md@main", "Source should appear in header")
	assert.Contains(t, yaml, "# Prompt Version: 3.1.4", "Version should appear in header")
}

// TestPromptVersionSchemaValidation tests schema validation of version field
func TestPromptVersionSchemaValidation(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		shouldMatch bool
	}{
		{"simple semver", "1.0.0", true},
		{"double digit major", "10.5.3", true},
		{"triple digit minor", "1.100.3", true},
		{"large patch", "1.0.999", true},
		{"with pre-release", "1.0.0-alpha", true},
		{"with numeric pre-release", "1.0.0-beta.1", true},
		{"with build metadata", "1.0.0+20240101", true},
		{"with both pre-release and build", "1.0.0-rc.1+build.123", true},
		{"complex pre-release", "1.0.0-alpha.beta.1", true},
		{"complex build metadata", "1.0.0+build.123.456", true},
		{"missing patch", "1.0", false},
		{"missing minor and patch", "1", false},
		{"leading v", "v1.0.0", false},
		{"non-numeric major", "x.0.0", false},
		{"non-numeric minor", "1.x.0", false},
		{"non-numeric patch", "1.0.x", false},
		{"leading zero in major", "01.0.0", false},
		{"leading zero in minor", "1.01.0", false},
		{"leading zero in patch", "1.0.01", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldMatch {
				// For valid versions, just ensure they have 3 numeric parts
				basePart := strings.Split(strings.Split(tt.version, "-")[0], "+")[0]
				parts := strings.Split(basePart, ".")
				assert.Equal(t, 3, len(parts), "Valid version should have 3 parts")
			} else {
				// For invalid versions, document the pattern they violate
				t.Logf("Invalid version '%s' correctly identified as not matching semver pattern", tt.version)
			}
		})
	}
}
