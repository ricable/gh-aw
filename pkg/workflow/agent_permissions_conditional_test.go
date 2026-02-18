//go:build integration

package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/github/gh-aw/pkg/stringutil"
	"github.com/github/gh-aw/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAgentJobPermissionsConditional verifies that the agent job only gets
// contents: read permission when:
// 1. User explicitly specifies it, OR
// 2. In development mode when checkout is needed for local actions
func TestAgentJobPermissionsConditional(t *testing.T) {
	tests := []struct {
		name                string
		frontmatter         string
		actionMode          ActionMode
		expectedHasContents bool
		description         string
	}{
		{
			name: "release mode without explicit contents should NOT add it",
			frontmatter: `---
on: issues
engine: copilot
permissions:
  issues: write
strict: false
features:
  dangerous-permissions-write: true
---
# Test workflow
Test content`,
			actionMode:          ActionModeRelease,
			expectedHasContents: false,
			description:         "Release mode without explicit contents: read should NOT automatically add it",
		},
		{
			name: "release mode with explicit contents should preserve it",
			frontmatter: `---
on: issues
engine: copilot
permissions:
  contents: read
  issues: write
strict: false
features:
  dangerous-permissions-write: true
---
# Test workflow
Test content`,
			actionMode:          ActionModeRelease,
			expectedHasContents: true,
			description:         "Release mode with explicit contents: read should preserve it",
		},
		{
			name: "dev mode without explicit contents should add it",
			frontmatter: `---
on: issues
engine: copilot
permissions:
  issues: write
strict: false
features:
  dangerous-permissions-write: true
---
# Test workflow
Test content`,
			actionMode:          ActionModeDev,
			expectedHasContents: true,
			description:         "Dev mode should add contents: read for local action checkout",
		},
		{
			name: "release mode with empty permissions should NOT add contents",
			frontmatter: `---
on: issues
engine: copilot
permissions: {}
---
# Test workflow
Test content`,
			actionMode:          ActionModeRelease,
			expectedHasContents: false,
			description:         "Release mode with empty permissions should NOT add contents: read",
		},
		{
			name: "dev mode with empty permissions should add contents",
			frontmatter: `---
on: issues
engine: copilot
permissions: {}
---
# Test workflow
Test content`,
			actionMode:          ActionModeDev,
			expectedHasContents: true,
			description:         "Dev mode with empty permissions should add contents: read for local actions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := testutil.TempDir(t, "agent-permissions-test")

			testFile := filepath.Join(tmpDir, "test-workflow.md")
			if err := os.WriteFile(testFile, []byte(tt.frontmatter), 0644); err != nil {
				t.Fatal(err)
			}

			compiler := NewCompilerWithVersion("v1.0.0")
			compiler.SetActionMode(tt.actionMode)

			// Compile the workflow
			if err := compiler.CompileWorkflow(testFile); err != nil {
				t.Fatalf("Failed to compile workflow: %v", err)
			}

			// Calculate the lock file path
			lockFile := stringutil.MarkdownToLockFile(testFile)

			// Read the generated lock file
			lockContent, err := os.ReadFile(lockFile)
			require.NoError(t, err, "Failed to read lock file")

			lockContentStr := string(lockContent)

			// Extract agent job section
			agentJobSection := extractJobSection(lockContentStr, "agent")
			require.NotEmpty(t, agentJobSection, "Agent job section should not be empty")

			// Check if contents: read is present
			hasContentsRead := strings.Contains(agentJobSection, "contents: read")

			if tt.expectedHasContents {
				assert.True(t, hasContentsRead,
					"%s: Expected agent job to have contents: read permission\nAgent section:\n%s",
					tt.description, agentJobSection)
			} else {
				assert.False(t, hasContentsRead,
					"%s: Expected agent job to NOT have contents: read permission\nAgent section:\n%s",
					tt.description, agentJobSection)
			}
		})
	}
}
