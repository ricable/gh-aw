//go:build integration

package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSandboxDisabled tests that sandbox: false is now rejected
func TestSandboxDisabled(t *testing.T) {
	t.Run("sandbox: false is rejected", func(t *testing.T) {
		workflowsDir := t.TempDir()

		markdown := `---
engine: copilot
sandbox: false
strict: false
on: workflow_dispatch
---

Test workflow with sandbox disabled.
`

		workflowPath := filepath.Join(workflowsDir, "test-sandbox-disabled.md")
		err := os.WriteFile(workflowPath, []byte(markdown), 0644)
		require.NoError(t, err)

		compiler := NewCompiler()
		compiler.SetStrictMode(false)

		err = compiler.CompileWorkflow(workflowPath)
		require.Error(t, err, "Compilation should fail with sandbox: false")
		assert.Contains(t, err.Error(), "disabling the sandbox is no longer supported")
	})

	t.Run("sandbox: false is rejected in strict mode", func(t *testing.T) {
		workflowsDir := t.TempDir()

		markdown := `---
engine: copilot
sandbox: false
strict: true
on: workflow_dispatch
---

Test workflow with sandbox disabled in strict mode.
`

		workflowPath := filepath.Join(workflowsDir, "test-sandbox-disabled-strict.md")
		err := os.WriteFile(workflowPath, []byte(markdown), 0644)
		require.NoError(t, err)

		compiler := NewCompiler()
		compiler.SetStrictMode(true)

		err = compiler.CompileWorkflow(workflowPath)
		require.Error(t, err, "Expected error when sandbox: false in strict mode")
		assert.Contains(t, err.Error(), "disabling the sandbox is no longer supported")
	})

	t.Run("sandbox: true is treated as unconfigured", func(t *testing.T) {
		workflowsDir := t.TempDir()

		markdown := `---
engine: copilot
sandbox: true
network:
  allowed:
    - defaults
on: workflow_dispatch
---

Test workflow with sandbox: true.
`

		workflowPath := filepath.Join(workflowsDir, "test-sandbox-true.md")
		err := os.WriteFile(workflowPath, []byte(markdown), 0644)
		require.NoError(t, err)

		compiler := NewCompiler()
		compiler.SetStrictMode(false)
		compiler.SetSkipValidation(true)

		err = compiler.CompileWorkflow(workflowPath)
		require.NoError(t, err)

		// Read the compiled workflow
		lockPath := filepath.Join(workflowsDir, "test-sandbox-true.lock.yml")
		lockContent, err := os.ReadFile(lockPath)
		require.NoError(t, err)
		result := string(lockContent)

		// sandbox: true should be treated as if no sandbox config was specified
		// This means AWF should be enabled by default
		assert.Contains(t, result, "sudo -E awf", "Workflow should contain AWF command by default when sandbox: true")
	})
}

// TestSandboxDisabledWithToolsConfiguration tests that sandbox: false is rejected even with tools configured
func TestSandboxDisabledWithToolsConfiguration(t *testing.T) {
	workflowsDir := t.TempDir()

	markdown := `---
engine: copilot
sandbox: false
strict: false
tools:
  github:
    mode: local
    toolsets: [repos, issues]
on: workflow_dispatch
---

Test workflow with tools and sandbox disabled.
`

	workflowPath := filepath.Join(workflowsDir, "test-sandbox-disabled-tools.md")
	err := os.WriteFile(workflowPath, []byte(markdown), 0644)
	require.NoError(t, err)

	compiler := NewCompiler()
	compiler.SetStrictMode(false)

	err = compiler.CompileWorkflow(workflowPath)
	require.Error(t, err, "Compilation should fail with sandbox: false")
	assert.Contains(t, err.Error(), "disabling the sandbox is no longer supported")
}

// TestSandboxDisabledCopilotExecution tests that sandbox: false is rejected
func TestSandboxDisabledCopilotExecution(t *testing.T) {
	workflowsDir := t.TempDir()

	markdown := `---
engine: copilot
sandbox: false
strict: false
network:
  allowed:
    - api.github.com
on: workflow_dispatch
---

Test workflow with direct copilot execution.
`

	workflowPath := filepath.Join(workflowsDir, "test-sandbox-disabled-execution.md")
	err := os.WriteFile(workflowPath, []byte(markdown), 0644)
	require.NoError(t, err)

	compiler := NewCompiler()
	compiler.SetStrictMode(false)

	err = compiler.CompileWorkflow(workflowPath)
	require.Error(t, err, "Compilation should fail with sandbox: false")
	assert.Contains(t, err.Error(), "disabling the sandbox is no longer supported")
}
