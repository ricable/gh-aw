//go:build integration

package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/github/gh-aw/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFrontmatterCheckout_SingleObject verifies that a single checkout object in frontmatter
// is used to override the main repository checkout step.
func TestFrontmatterCheckout_SingleObject(t *testing.T) {
	frontmatter := `---
on:
  issues:
    types: [opened]
permissions:
  contents: read
  issues: read
engine: copilot
strict: false
checkout:
  ref: my-feature-branch
  fetch-depth: 0
---`
	markdown := "# Agent\n\nComplete the task."

	tmpDir := testutil.TempDir(t, "frontmatter-checkout-single-test")
	workflowPath := filepath.Join(tmpDir, "test.md")
	require.NoError(t, os.WriteFile(workflowPath, []byte(frontmatter+"\n\n"+markdown), 0644))

	compiler := NewCompiler()
	require.NoError(t, compiler.CompileWorkflow(workflowPath))

	lockFile := strings.TrimSuffix(workflowPath, ".md") + ".lock.yml"
	lockContent, err := os.ReadFile(lockFile)
	require.NoError(t, err)

	lockStr := string(lockContent)

	// The main checkout step should include the user-specified fields.
	assert.Contains(t, lockStr, "name: Checkout repository", "should have main checkout step")
	assert.Contains(t, lockStr, "ref: my-feature-branch", "should include user-specified ref")
	assert.Contains(t, lockStr, "fetch-depth: 0", "should include fetch-depth")
	assert.Contains(t, lockStr, "persist-credentials: false", "should keep persist-credentials false")
}

// TestFrontmatterCheckout_ArrayMultiple verifies that an array of checkout objects generates
// the main checkout plus additional checkout steps, each in its own subfolder.
func TestFrontmatterCheckout_ArrayMultiple(t *testing.T) {
	frontmatter := `---
on:
  issues:
    types: [opened]
permissions:
  contents: read
  issues: read
engine: copilot
strict: false
checkout:
  - ref: main
  - repository: org/tools
    ref: v2.0.0
    path: tools
---`
	markdown := "# Agent\n\nComplete the task."

	tmpDir := testutil.TempDir(t, "frontmatter-checkout-array-test")
	workflowPath := filepath.Join(tmpDir, "test.md")
	require.NoError(t, os.WriteFile(workflowPath, []byte(frontmatter+"\n\n"+markdown), 0644))

	compiler := NewCompiler()
	require.NoError(t, compiler.CompileWorkflow(workflowPath))

	lockFile := strings.TrimSuffix(workflowPath, ".md") + ".lock.yml"
	lockContent, err := os.ReadFile(lockFile)
	require.NoError(t, err)

	lockStr := string(lockContent)

	// Main checkout step should reflect first array entry (no path → main checkout override).
	assert.Contains(t, lockStr, "name: Checkout repository", "should have main checkout step")
	assert.Contains(t, lockStr, "ref: main", "first array entry without path should override main checkout ref")

	// Additional checkout for org/tools.
	assert.Contains(t, lockStr, "repository: org/tools", "should include additional repository")
	assert.Contains(t, lockStr, "ref: v2.0.0", "should include additional ref")
	assert.Contains(t, lockStr, "path: tools", "should include explicit path for additional checkout")
}

// TestFrontmatterCheckout_ArrayAllWithPaths verifies that when all array entries have explicit paths,
// all of them are emitted as additional checkouts and the main checkout uses defaults.
func TestFrontmatterCheckout_ArrayAllWithPaths(t *testing.T) {
	frontmatter := `---
on:
  issues:
    types: [opened]
permissions:
  contents: read
  issues: read
engine: copilot
strict: false
checkout:
  - repository: org/repo1
    path: repo1
  - repository: org/repo2
    ref: develop
    path: repo2
---`
	markdown := "# Agent\n\nComplete the task."

	tmpDir := testutil.TempDir(t, "frontmatter-checkout-all-paths-test")
	workflowPath := filepath.Join(tmpDir, "test.md")
	require.NoError(t, os.WriteFile(workflowPath, []byte(frontmatter+"\n\n"+markdown), 0644))

	compiler := NewCompiler()
	require.NoError(t, compiler.CompileWorkflow(workflowPath))

	lockFile := strings.TrimSuffix(workflowPath, ".md") + ".lock.yml"
	lockContent, err := os.ReadFile(lockFile)
	require.NoError(t, err)

	lockStr := string(lockContent)

	// Main checkout should be the default.
	assert.Contains(t, lockStr, "name: Checkout repository", "should have default main checkout")

	// Both additional checkouts.
	assert.Contains(t, lockStr, "repository: org/repo1", "should include repo1")
	assert.Contains(t, lockStr, "path: repo1", "should include path for repo1")
	assert.Contains(t, lockStr, "repository: org/repo2", "should include repo2")
	assert.Contains(t, lockStr, "ref: develop", "should include ref for repo2")
	assert.Contains(t, lockStr, "path: repo2", "should include path for repo2")

	// Check ordering: main checkout before additional checkouts.
	mainIdx := strings.Index(lockStr, "name: Checkout repository")
	repo1Idx := strings.Index(lockStr, "repository: org/repo1")
	assert.Less(t, mainIdx, repo1Idx, "main checkout should come before additional checkouts")
}

// TestFrontmatterCheckout_AutoPath verifies that when an additional checkout has no path,
// the path is automatically derived from the repository slug.
func TestFrontmatterCheckout_AutoPath(t *testing.T) {
	frontmatter := `---
on:
  issues:
    types: [opened]
permissions:
  contents: read
  issues: read
engine: copilot
strict: false
checkout:
  - path: main
  - repository: org/mytools
---`
	markdown := "# Agent\n\nComplete the task."

	tmpDir := testutil.TempDir(t, "frontmatter-checkout-autopath-test")
	workflowPath := filepath.Join(tmpDir, "test.md")
	require.NoError(t, os.WriteFile(workflowPath, []byte(frontmatter+"\n\n"+markdown), 0644))

	compiler := NewCompiler()
	require.NoError(t, compiler.CompileWorkflow(workflowPath))

	lockFile := strings.TrimSuffix(workflowPath, ".md") + ".lock.yml"
	lockContent, err := os.ReadFile(lockFile)
	require.NoError(t, err)

	lockStr := string(lockContent)

	// Second additional checkout: path auto-derived from "org/mytools" → "mytools"
	assert.Contains(t, lockStr, "repository: org/mytools", "should include the repo")
	assert.Contains(t, lockStr, "path: mytools", "should auto-derive path from repo slug")
}
