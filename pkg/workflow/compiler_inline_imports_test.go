//go:build !integration

package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInlineImports_Enabled(t *testing.T) {
	// Create temporary directory for test files
	tempDir := t.TempDir()

	// Create a shared fragment file
	fragmentContent := `---
tools:
  bash: [echo]
---

# Shared Fragment

This is shared content that should be inlined.
`
	sharedDir := filepath.Join(tempDir, "shared")
	err := os.MkdirAll(sharedDir, 0755)
	require.NoError(t, err, "Failed to create shared directory")

	fragmentPath := filepath.Join(sharedDir, "test-fragment.md")
	err = os.WriteFile(fragmentPath, []byte(fragmentContent), 0644)
	require.NoError(t, err, "Failed to write fragment file")

	// Create a workflow file with inline-imports: true
	workflowContent := `---
name: Test Inline Imports
on: issues
engine: copilot
inline-imports: true
imports:
  - shared/test-fragment.md
---

# Main Workflow

This is the main workflow content.
`
	workflowPath := filepath.Join(tempDir, "test-workflow.md")
	err = os.WriteFile(workflowPath, []byte(workflowContent), 0644)
	require.NoError(t, err, "Failed to write workflow file")

	// Compile the workflow
	compiler := NewCompiler(

		WithSkipValidation(true),
	)

	err = compiler.CompileWorkflow(workflowPath)
	require.NoError(t, err, "Compilation should succeed")

	// Read the generated lock file
	lockFilePath := strings.TrimSuffix(workflowPath, ".md") + ".lock.yml"
	lockContent, err := os.ReadFile(lockFilePath)
	require.NoError(t, err, "Failed to read lock file")
	lockString := string(lockContent)

	// Verify that runtime-import macros are NOT present
	assert.NotContains(t, lockString, "{{#runtime-import", "Should not contain runtime-import macros when inline-imports is true")

	// Verify that both the fragment and main workflow content are inlined
	assert.Contains(t, lockString, "# Shared Fragment", "Should contain inlined fragment heading")
	assert.Contains(t, lockString, "This is shared content that should be inlined", "Should contain inlined fragment content")
	assert.Contains(t, lockString, "# Main Workflow", "Should contain inlined main workflow heading")
	assert.Contains(t, lockString, "This is the main workflow content", "Should contain inlined main workflow content")
}

func TestInlineImports_Disabled(t *testing.T) {
	// Create temporary directory for test files
	tempDir := t.TempDir()

	// Create a shared fragment file
	fragmentContent := `---
tools:
  bash: [echo]
---

# Shared Fragment

This is shared content loaded at runtime.
`
	sharedDir := filepath.Join(tempDir, "shared")
	err := os.MkdirAll(sharedDir, 0755)
	require.NoError(t, err, "Failed to create shared directory")

	fragmentPath := filepath.Join(sharedDir, "test-fragment.md")
	err = os.WriteFile(fragmentPath, []byte(fragmentContent), 0644)
	require.NoError(t, err, "Failed to write fragment file")

	// Create a workflow file with inline-imports: false
	workflowContent := `---
name: Test Runtime Import
on: issues
engine: copilot
inline-imports: false
imports:
  - shared/test-fragment.md
---

# Main Workflow

This is the main workflow content with runtime imports.
`
	workflowPath := filepath.Join(tempDir, "test-workflow.md")
	err = os.WriteFile(workflowPath, []byte(workflowContent), 0644)
	require.NoError(t, err, "Failed to write workflow file")

	// Compile the workflow
	compiler := NewCompiler(

		WithSkipValidation(true),
	)

	err = compiler.CompileWorkflow(workflowPath)
	require.NoError(t, err, "Compilation should succeed")

	// Read the generated lock file
	lockFilePath := strings.TrimSuffix(workflowPath, ".md") + ".lock.yml"
	lockContent, err := os.ReadFile(lockFilePath)
	require.NoError(t, err, "Failed to read lock file")
	lockString := string(lockContent)

	// Verify that runtime-import macros ARE present
	assert.Contains(t, lockString, "{{#runtime-import shared/test-fragment.md}}", "Should contain runtime-import macro for fragment when inline-imports is false")
	assert.Contains(t, lockString, "{{#runtime-import test-workflow.md}}", "Should contain runtime-import macro for main workflow when inline-imports is false")

	// Verify that the fragment content is NOT inlined
	assert.NotContains(t, lockString, "This is shared content loaded at runtime", "Should not inline fragment content when inline-imports is false")
}

func TestInlineImports_Default(t *testing.T) {
	// Create temporary directory for test files
	tempDir := t.TempDir()

	// Create a shared fragment file
	fragmentContent := `---
tools:
  bash: [echo]
---

# Shared Fragment

This is shared content loaded at runtime by default.
`
	sharedDir := filepath.Join(tempDir, "shared")
	err := os.MkdirAll(sharedDir, 0755)
	require.NoError(t, err, "Failed to create shared directory")

	fragmentPath := filepath.Join(sharedDir, "test-fragment.md")
	err = os.WriteFile(fragmentPath, []byte(fragmentContent), 0644)
	require.NoError(t, err, "Failed to write fragment file")

	// Create a workflow file WITHOUT inline-imports field (should default to false)
	workflowContent := `---
name: Test Default Behavior
on: issues
engine: copilot
imports:
  - shared/test-fragment.md
---

# Main Workflow

This is the main workflow content with default behavior.
`
	workflowPath := filepath.Join(tempDir, "test-workflow.md")
	err = os.WriteFile(workflowPath, []byte(workflowContent), 0644)
	require.NoError(t, err, "Failed to write workflow file")

	// Compile the workflow
	compiler := NewCompiler(

		WithSkipValidation(true),
	)

	err = compiler.CompileWorkflow(workflowPath)
	require.NoError(t, err, "Compilation should succeed")

	// Read the generated lock file
	lockFilePath := strings.TrimSuffix(workflowPath, ".md") + ".lock.yml"
	lockContent, err := os.ReadFile(lockFilePath)
	require.NoError(t, err, "Failed to read lock file")
	lockString := string(lockContent)

	// Verify that runtime-import macros ARE present (default behavior)
	assert.Contains(t, lockString, "{{#runtime-import shared/test-fragment.md}}", "Should contain runtime-import macro for fragment by default")
	assert.Contains(t, lockString, "{{#runtime-import test-workflow.md}}", "Should contain runtime-import macro for main workflow by default")

	// Verify that the fragment content is NOT inlined
	assert.NotContains(t, lockString, "This is shared content loaded at runtime by default", "Should not inline fragment content by default")
}

func TestInlineImports_WithImportInputs(t *testing.T) {
	// Create temporary directory for test files
	tempDir := t.TempDir()

	// Create a shared fragment file with inputs
	fragmentContent := `---
inputs:
  message:
    description: A message to display
    required: true
    default: "Hello"
tools:
  bash: [echo]
---

# Shared Fragment

Message: {{ inputs.message }}
`
	sharedDir := filepath.Join(tempDir, "shared")
	err := os.MkdirAll(sharedDir, 0755)
	require.NoError(t, err, "Failed to create shared directory")

	fragmentPath := filepath.Join(sharedDir, "test-fragment-inputs.md")
	err = os.WriteFile(fragmentPath, []byte(fragmentContent), 0644)
	require.NoError(t, err, "Failed to write fragment file")

	// Create a workflow file with inline-imports: true and import with inputs
	workflowContent := `---
name: Test Inline Imports With Inputs
on: issues
engine: copilot
inline-imports: true
imports:
  - path: shared/test-fragment-inputs.md
    inputs:
      message: "Custom message"
---

# Main Workflow

This is the main workflow content.
`
	workflowPath := filepath.Join(tempDir, "test-workflow.md")
	err = os.WriteFile(workflowPath, []byte(workflowContent), 0644)
	require.NoError(t, err, "Failed to write workflow file")

	// Compile the workflow
	compiler := NewCompiler(

		WithSkipValidation(true),
	)

	err = compiler.CompileWorkflow(workflowPath)
	require.NoError(t, err, "Compilation should succeed")

	// Read the generated lock file
	lockFilePath := strings.TrimSuffix(workflowPath, ".md") + ".lock.yml"
	lockContent, err := os.ReadFile(lockFilePath)
	require.NoError(t, err, "Failed to read lock file")
	lockString := string(lockContent)

	// Verify that runtime-import macros are NOT present
	assert.NotContains(t, lockString, "{{#runtime-import", "Should not contain runtime-import macros when inline-imports is true")

	// Verify that the fragment content is inlined (input template is preserved)
	assert.Contains(t, lockString, "# Shared Fragment", "Should contain inlined fragment heading")
	assert.Contains(t, lockString, "Message: {{ inputs.message }}", "Should contain inlined fragment with input template")
}

func TestExtractInlineImports(t *testing.T) {
	compiler := NewCompiler()

	tests := []struct {
		name        string
		frontmatter map[string]any
		expected    bool
	}{
		{
			name: "inline-imports true",
			frontmatter: map[string]any{
				"name":           "Test",
				"on":             "issues",
				"inline-imports": true,
			},
			expected: true,
		},
		{
			name: "inline-imports false",
			frontmatter: map[string]any{
				"name":           "Test",
				"on":             "issues",
				"inline-imports": false,
			},
			expected: false,
		},
		{
			name: "inline-imports not present",
			frontmatter: map[string]any{
				"name": "Test",
				"on":   "issues",
			},
			expected: false,
		},
		{
			name: "inline-imports wrong type",
			frontmatter: map[string]any{
				"name":           "Test",
				"on":             "issues",
				"inline-imports": "true", // string instead of bool
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compiler.extractInlineImports(tt.frontmatter)
			assert.Equal(t, tt.expected, result, "extractInlineImports should return %v for %s", tt.expected, tt.name)
		})
	}
}
