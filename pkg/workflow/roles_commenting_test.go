//go:build !integration

package workflow

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/github/gh-aw/pkg/stringutil"
	"github.com/github/gh-aw/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRolesCommentedOut tests that roles field is properly commented out in the on section
func TestRolesCommentedOut(t *testing.T) {
	tmpDir := testutil.TempDir(t, "roles-commenting-test")
	compiler := NewCompiler()

	t.Run("roles_array_commented_out", func(t *testing.T) {
		workflowContent := `---
on:
  roles: [admin, maintainer, write]
  slash_command:
    name: test
engine: copilot
permissions:
  contents: read
---

# Test Workflow with Roles Array

This workflow has roles specified as an array.
`
		workflowFile := filepath.Join(tmpDir, "roles-array.md")
		err := os.WriteFile(workflowFile, []byte(workflowContent), 0644)
		require.NoError(t, err, "Failed to write workflow file")

		err = compiler.CompileWorkflow(workflowFile)
		require.NoError(t, err, "Compilation failed")

		lockFile := stringutil.MarkdownToLockFile(workflowFile)
		lockContent, err := os.ReadFile(lockFile)
		require.NoError(t, err, "Failed to read lock file")

		lockContentStr := string(lockContent)

		// Verify roles field is commented out
		assert.Contains(t, lockContentStr, "# roles:", "Expected roles field to be commented out in lock file")
		assert.Contains(t, lockContentStr, "# Roles processed as role check in pre-activation job", "Expected comment reason for roles field")

		// Verify array items are commented out (with indentation of 2 spaces)
		assert.Contains(t, lockContentStr, "  # - admin", "Expected admin role to be commented out")
		assert.Contains(t, lockContentStr, "  # - maintainer", "Expected maintainer role to be commented out")
		assert.Contains(t, lockContentStr, "  # - write", "Expected write role to be commented out")

		// Verify the slash_command is NOT commented out
		assert.Contains(t, lockContentStr, "slash_command:", "Expected slash_command to remain in lock file")
		assert.NotContains(t, lockContentStr, "# slash_command:", "Expected slash_command to NOT be commented out")
	})

	t.Run("roles_single_value_all_commented_out", func(t *testing.T) {
		workflowContent := `---
on:
  roles: all
  workflow_dispatch:
engine: copilot
permissions:
  contents: read
---

# Test Workflow with Single Role

This workflow allows all roles.
`
		workflowFile := filepath.Join(tmpDir, "roles-single.md")
		err := os.WriteFile(workflowFile, []byte(workflowContent), 0644)
		require.NoError(t, err, "Failed to write workflow file")

		err = compiler.CompileWorkflow(workflowFile)
		require.NoError(t, err, "Compilation failed")

		lockFile := stringutil.MarkdownToLockFile(workflowFile)
		lockContent, err := os.ReadFile(lockFile)
		require.NoError(t, err, "Failed to read lock file")

		lockContentStr := string(lockContent)

		// Verify roles field is commented out
		assert.Contains(t, lockContentStr, "# roles:", "Expected roles field to be commented out in lock file")
		assert.Contains(t, lockContentStr, "# Roles processed as role check in pre-activation job", "Expected comment reason for roles field")

		// Verify workflow_dispatch is NOT commented out
		assert.Contains(t, lockContentStr, "workflow_dispatch:", "Expected workflow_dispatch to remain in lock file")
		assert.NotContains(t, lockContentStr, "# workflow_dispatch:", "Expected workflow_dispatch to NOT be commented out")
	})

	t.Run("roles_with_special_value_all", func(t *testing.T) {
		workflowContent := `---
on:
  roles: all
  issue_comment:
    types: [created]
engine: copilot
permissions:
  contents: read
---

# Test Workflow with Roles All

This workflow allows all roles.
`
		workflowFile := filepath.Join(tmpDir, "roles-all.md")
		err := os.WriteFile(workflowFile, []byte(workflowContent), 0644)
		require.NoError(t, err, "Failed to write workflow file")

		err = compiler.CompileWorkflow(workflowFile)
		require.NoError(t, err, "Compilation failed")

		lockFile := stringutil.MarkdownToLockFile(workflowFile)
		lockContent, err := os.ReadFile(lockFile)
		require.NoError(t, err, "Failed to read lock file")

		lockContentStr := string(lockContent)

		// Verify roles field is commented out
		assert.Contains(t, lockContentStr, "# roles: all", "Expected roles field with 'all' value to be commented out in lock file")

		// Verify issue_comment is NOT commented out
		assert.Contains(t, lockContentStr, "issue_comment:", "Expected issue_comment to remain in lock file")
	})

	t.Run("roles_with_other_on_fields", func(t *testing.T) {
		workflowContent := `---
on:
  roles: [admin, maintainer]
  skip-roles: [read]
  skip-bots: [github-actions]
  issues:
    types: [opened]
engine: copilot
permissions:
  contents: read
---

# Test Workflow with Multiple On Fields

This workflow has multiple custom on fields.
`
		workflowFile := filepath.Join(tmpDir, "roles-with-others.md")
		err := os.WriteFile(workflowFile, []byte(workflowContent), 0644)
		require.NoError(t, err, "Failed to write workflow file")

		err = compiler.CompileWorkflow(workflowFile)
		require.NoError(t, err, "Compilation failed")

		lockFile := stringutil.MarkdownToLockFile(workflowFile)
		lockContent, err := os.ReadFile(lockFile)
		require.NoError(t, err, "Failed to read lock file")

		lockContentStr := string(lockContent)

		// Verify all custom fields are commented out
		assert.Contains(t, lockContentStr, "# roles:", "Expected roles field to be commented out")
		assert.Contains(t, lockContentStr, "# skip-roles:", "Expected skip-roles field to be commented out")
		assert.Contains(t, lockContentStr, "# skip-bots:", "Expected skip-bots field to be commented out")

		// Verify issues is NOT commented out
		assert.Contains(t, lockContentStr, "issues:", "Expected issues to remain in lock file")
		assert.NotContains(t, lockContentStr, "# issues:", "Expected issues to NOT be commented out")
	})
}

// TestRolesFieldSchemaValidation tests that workflows with roles field pass GitHub Actions schema validation
func TestRolesFieldSchemaValidation(t *testing.T) {
	tmpDir := testutil.TempDir(t, "roles-schema-test")
	compiler := NewCompiler()
	// Schema validation is enabled by default (skipValidation = false)

	workflowContent := `---
on:
  roles: [admin, maintainer, write]
  slash_command:
    name: test
engine: copilot
permissions:
  contents: read
  issues: read
  pull-requests: read
---

# Schema Validation Test

This workflow should pass schema validation with roles field properly commented out.
`
	workflowFile := filepath.Join(tmpDir, "schema-validation.md")
	err := os.WriteFile(workflowFile, []byte(workflowContent), 0644)
	require.NoError(t, err, "Failed to write workflow file")

	err = compiler.CompileWorkflow(workflowFile)
	require.NoError(t, err, "Compilation should succeed - roles field should be commented out and not cause validation errors")

	lockFile := stringutil.MarkdownToLockFile(workflowFile)
	lockContent, err := os.ReadFile(lockFile)
	require.NoError(t, err, "Failed to read lock file")

	lockContentStr := string(lockContent)

	// Verify roles is commented out
	assert.Contains(t, lockContentStr, "# roles:", "Expected roles field to be commented out")

	// Verify the lock file doesn't have uncommented roles field in the on section
	// This regex ensures roles: is only found as a comment
	assert.NotRegexp(t, `(?m)^\s{2}roles:`, lockContentStr, "Expected no uncommented 'roles:' field at indent level 2 in on section")
}
