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

// TestFrontmatterCheckout_ImportedSingleCheckout verifies that a checkout field in an imported
// agentic workflow is merged into the main workflow as an additional checkout.
func TestFrontmatterCheckout_ImportedSingleCheckout(t *testing.T) {
	tmpDir := testutil.TempDir(t, "frontmatter-checkout-import-single-test")

	// Shared/imported workflow that declares a checkout for an extra repo
	importContent := `---
checkout:
  repository: org/shared-tools
  ref: v1.0.0
  path: shared-tools
---

# Shared Tools

Use shared tools from org/shared-tools.
`
	importPath := filepath.Join(tmpDir, "shared.md")
	require.NoError(t, os.WriteFile(importPath, []byte(importContent), 0644))

	// Main workflow that imports the shared workflow
	mainContent := `---
on:
  issues:
    types: [opened]
permissions:
  contents: read
  issues: read
engine: copilot
strict: false
imports:
  - shared.md
---

# Main Workflow

Complete the task.
`
	workflowPath := filepath.Join(tmpDir, "main.md")
	require.NoError(t, os.WriteFile(workflowPath, []byte(mainContent), 0644))

	compiler := NewCompiler()
	require.NoError(t, compiler.CompileWorkflow(workflowPath))

	lockFile := strings.TrimSuffix(workflowPath, ".md") + ".lock.yml"
	lockContent, err := os.ReadFile(lockFile)
	require.NoError(t, err)

	lockStr := string(lockContent)

	// The imported checkout should appear as an additional checkout step
	assert.Contains(t, lockStr, "repository: org/shared-tools", "should include imported repository")
	assert.Contains(t, lockStr, "ref: v1.0.0", "should include imported ref")
	assert.Contains(t, lockStr, "path: shared-tools", "should include imported path")
	assert.Contains(t, lockStr, "persist-credentials: false", "should default persist-credentials to false")

	// Main checkout should still be present
	assert.Contains(t, lockStr, "name: Checkout repository", "should still have main checkout")

	// Main checkout should come before the imported additional checkout
	mainIdx := strings.Index(lockStr, "name: Checkout repository")
	importedIdx := strings.Index(lockStr, "repository: org/shared-tools")
	assert.Less(t, mainIdx, importedIdx, "main checkout should precede imported additional checkout")
}

// TestFrontmatterCheckout_ImportedArrayCheckout verifies that multiple checkout entries in an
// imported agentic workflow are all merged as additional checkouts.
func TestFrontmatterCheckout_ImportedArrayCheckout(t *testing.T) {
	tmpDir := testutil.TempDir(t, "frontmatter-checkout-import-array-test")

	// Shared/imported workflow that declares multiple checkouts
	importContent := `---
checkout:
  - repository: org/lib-a
    path: lib-a
  - repository: org/lib-b
    ref: develop
    path: lib-b
---

# Shared Libraries
`
	importPath := filepath.Join(tmpDir, "libs.md")
	require.NoError(t, os.WriteFile(importPath, []byte(importContent), 0644))

	mainContent := `---
on:
  issues:
    types: [opened]
permissions:
  contents: read
  issues: read
engine: copilot
strict: false
imports:
  - libs.md
---

# Main Workflow
`
	workflowPath := filepath.Join(tmpDir, "main.md")
	require.NoError(t, os.WriteFile(workflowPath, []byte(mainContent), 0644))

	compiler := NewCompiler()
	require.NoError(t, compiler.CompileWorkflow(workflowPath))

	lockFile := strings.TrimSuffix(workflowPath, ".md") + ".lock.yml"
	lockContent, err := os.ReadFile(lockFile)
	require.NoError(t, err)

	lockStr := string(lockContent)

	// Both imported checkouts should be present
	assert.Contains(t, lockStr, "repository: org/lib-a", "should include first imported checkout")
	assert.Contains(t, lockStr, "path: lib-a", "should include first imported path")
	assert.Contains(t, lockStr, "repository: org/lib-b", "should include second imported checkout")
	assert.Contains(t, lockStr, "ref: develop", "should include second imported ref")
	assert.Contains(t, lockStr, "path: lib-b", "should include second imported path")
}

// TestFrontmatterCheckout_MainAndImportedMerged verifies that checkout configs from both the
// main workflow and an imported workflow are merged: the main workflow's config controls the main
// checkout step, and the imported workflow's checkout(s) are appended as additional checkouts.
func TestFrontmatterCheckout_MainAndImportedMerged(t *testing.T) {
	tmpDir := testutil.TempDir(t, "frontmatter-checkout-main-and-import-test")

	// Imported workflow declares an additional checkout
	importContent := `---
checkout:
  repository: org/data
  ref: main
  path: data
---

# Data
`
	importPath := filepath.Join(tmpDir, "data.md")
	require.NoError(t, os.WriteFile(importPath, []byte(importContent), 0644))

	// Main workflow overrides the main checkout ref AND imports the shared workflow
	mainContent := `---
on:
  issues:
    types: [opened]
permissions:
  contents: read
  issues: read
engine: copilot
strict: false
checkout:
  ref: my-branch
imports:
  - data.md
---

# Main Workflow
`
	workflowPath := filepath.Join(tmpDir, "main.md")
	require.NoError(t, os.WriteFile(workflowPath, []byte(mainContent), 0644))

	compiler := NewCompiler()
	require.NoError(t, compiler.CompileWorkflow(workflowPath))

	lockFile := strings.TrimSuffix(workflowPath, ".md") + ".lock.yml"
	lockContent, err := os.ReadFile(lockFile)
	require.NoError(t, err)

	lockStr := string(lockContent)

	// Main checkout should use the main workflow's ref override
	assert.Contains(t, lockStr, "name: Checkout repository", "should have main checkout step")
	assert.Contains(t, lockStr, "ref: my-branch", "main checkout should use main workflow's ref")

	// The imported additional checkout should also be present
	assert.Contains(t, lockStr, "repository: org/data", "should include imported repo")
	assert.Contains(t, lockStr, "path: data", "should include imported path")

	// Main checkout precedes imported additional checkout
	mainCheckoutIdx := strings.Index(lockStr, "name: Checkout repository")
	importedCheckoutIdx := strings.Index(lockStr, "repository: org/data")
	assert.Less(t, mainCheckoutIdx, importedCheckoutIdx, "main checkout should come before imported checkout")
}

// TestFrontmatterCheckout_MultipleCheckouts_DifferentTokens verifies that an array of checkout
// entries with distinct tokens compiles correctly and that every step has persist-credentials: false.
func TestFrontmatterCheckout_MultipleCheckouts_DifferentTokens(t *testing.T) {
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
  - repository: org/private-a
    ref: v1.0.0
    path: private-a
    token: ${{ secrets.TOKEN_A }}
  - repository: org/private-b
    ref: develop
    path: private-b
    token: ${{ secrets.TOKEN_B }}
---`
	markdown := "# Agent\n\nComplete the task."

	tmpDir := testutil.TempDir(t, "frontmatter-checkout-tokens-test")
	workflowPath := filepath.Join(tmpDir, "test.md")
	require.NoError(t, os.WriteFile(workflowPath, []byte(frontmatter+"\n\n"+markdown), 0644))

	compiler := NewCompiler()
	require.NoError(t, compiler.CompileWorkflow(workflowPath))

	lockFile := strings.TrimSuffix(workflowPath, ".md") + ".lock.yml"
	lockContent, err := os.ReadFile(lockFile)
	require.NoError(t, err)
	lockStr := string(lockContent)

	// Main checkout (ref: main override) must be present and have persist-credentials: false
	assert.Contains(t, lockStr, "name: Checkout repository", "should have main checkout step")
	assert.Contains(t, lockStr, "ref: main", "main checkout should use overridden ref")

	// Each additional checkout should carry its token
	assert.Contains(t, lockStr, "repository: org/private-a", "should include private-a")
	assert.Contains(t, lockStr, "token: ${{ secrets.TOKEN_A }}", "should include TOKEN_A")
	assert.Contains(t, lockStr, "repository: org/private-b", "should include private-b")
	assert.Contains(t, lockStr, "token: ${{ secrets.TOKEN_B }}", "should include TOKEN_B")

	// Every checkout step must have persist-credentials: false
	persistFalseCount := strings.Count(lockStr, "persist-credentials: false")
	assert.GreaterOrEqual(t, persistFalseCount, 3, "all checkout steps must have persist-credentials: false")
}

// TestFrontmatterCheckout_MultipleCheckouts_DifferentFetchDepths verifies that multiple checkouts
// can specify different fetch depths and that the compiled output reflects each independently.
func TestFrontmatterCheckout_MultipleCheckouts_DifferentFetchDepths(t *testing.T) {
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
  - fetch-depth: 0
  - repository: org/large-repo
    path: large-repo
    fetch-depth: 1
  - repository: org/small-repo
    path: small-repo
---`
	markdown := "# Agent\n\nComplete the task."

	tmpDir := testutil.TempDir(t, "frontmatter-checkout-depths-test")
	workflowPath := filepath.Join(tmpDir, "test.md")
	require.NoError(t, os.WriteFile(workflowPath, []byte(frontmatter+"\n\n"+markdown), 0644))

	compiler := NewCompiler()
	require.NoError(t, compiler.CompileWorkflow(workflowPath))

	lockFile := strings.TrimSuffix(workflowPath, ".md") + ".lock.yml"
	lockContent, err := os.ReadFile(lockFile)
	require.NoError(t, err)
	lockStr := string(lockContent)

	// Main checkout overrides: fetch-depth 0 (full history)
	assert.Contains(t, lockStr, "name: Checkout repository", "should have main checkout step")
	assert.Contains(t, lockStr, "fetch-depth: 0", "main checkout should have full history")

	// large-repo: shallow clone depth 1
	assert.Contains(t, lockStr, "repository: org/large-repo", "should include large-repo")
	assert.Contains(t, lockStr, "fetch-depth: 1", "large-repo should have shallow clone")

	// small-repo: no fetch-depth (actions/checkout default)
	assert.Contains(t, lockStr, "repository: org/small-repo", "should include small-repo")

	// persist-credentials: false must appear for every checkout step
	persistFalseCount := strings.Count(lockStr, "persist-credentials: false")
	assert.GreaterOrEqual(t, persistFalseCount, 3, "all checkout steps must have persist-credentials: false")
}

// TestFrontmatterCheckout_PersistCredentialsFalse_Enforced verifies that persist-credentials is
// always false by default even when other fields such as token and fetch-depth are specified.
func TestFrontmatterCheckout_PersistCredentialsFalse_Enforced(t *testing.T) {
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
    token: ${{ secrets.GITHUB_TOKEN }}
    fetch-depth: 0
  - repository: org/extra
    path: extra
    token: ${{ secrets.EXTRA_TOKEN }}
    fetch-depth: 10
---`
	markdown := "# Agent\n\nComplete the task."

	tmpDir := testutil.TempDir(t, "frontmatter-checkout-persist-creds-test")
	workflowPath := filepath.Join(tmpDir, "test.md")
	require.NoError(t, os.WriteFile(workflowPath, []byte(frontmatter+"\n\n"+markdown), 0644))

	compiler := NewCompiler()
	require.NoError(t, compiler.CompileWorkflow(workflowPath))

	lockFile := strings.TrimSuffix(workflowPath, ".md") + ".lock.yml"
	lockContent, err := os.ReadFile(lockFile)
	require.NoError(t, err)
	lockStr := string(lockContent)

	// persist-credentials: true must never appear
	assert.NotContains(t, lockStr, "persist-credentials: true", "persist-credentials must never be true by default")

	// persist-credentials: false must appear for every checkout step (main + additional)
	persistFalseCount := strings.Count(lockStr, "persist-credentials: false")
	assert.GreaterOrEqual(t, persistFalseCount, 2, "all checkout steps must have persist-credentials: false")
}
