//go:build !integration

package workflow

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContainsSparseCheckout(t *testing.T) {
	tests := []struct {
		name         string
		customSteps  string
		shouldDetect bool
	}{
		{
			name: "detects git sparse-checkout set",
			customSteps: `steps:
  - name: Checkout Python files
    run: |
      git sparse-checkout init --cone
      git sparse-checkout set src
`,
			shouldDetect: true,
		},
		{
			name: "detects sparse-checkout init",
			customSteps: `steps:
  - name: Setup sparse checkout
    run: git sparse-checkout init
`,
			shouldDetect: true,
		},
		{
			name: "detects mixed case",
			customSteps: `steps:
  - name: Use sparse checkout
    run: |
      Git Sparse-Checkout init
`,
			shouldDetect: true,
		},
		{
			name: "no sparse-checkout commands",
			customSteps: `steps:
  - name: Regular checkout
    uses: actions/checkout@v4
`,
			shouldDetect: false,
		},
		{
			name:         "empty custom steps",
			customSteps:  "",
			shouldDetect: false,
		},
		{
			name: "sparse in comment detected (conservative)",
			customSteps: `steps:
  - name: Do something
    run: |
      # This doesn't use git sparse-checkout
      git checkout main
`,
			shouldDetect: true, // Conservative detection includes comments
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsSparseCheckout(tt.customSteps)
			assert.Equal(t, tt.shouldDetect, result,
				"ContainsSparseCheckout() detection mismatch")
		})
	}
}

func TestSparseCheckoutRecoveryStep(t *testing.T) {
	tests := []struct {
		name                 string
		customSteps          string
		importPaths          []string
		mainWorkflowMarkdown string
		shouldGenerate       bool
	}{
		{
			name: "generates step when sparse-checkout is used with runtime imports",
			customSteps: `steps:
  - name: Sparse checkout
    run: git sparse-checkout set src
`,
			importPaths:          []string{".github/workflows/shared.md"},
			mainWorkflowMarkdown: "# Test",
			shouldGenerate:       true,
		},
		{
			name: "generates step with only main workflow markdown",
			customSteps: `steps:
  - name: Sparse checkout
    run: git sparse-checkout init
`,
			importPaths:          []string{},
			mainWorkflowMarkdown: "# Main workflow",
			shouldGenerate:       true,
		},
		{
			name: "skips step when no sparse-checkout",
			customSteps: `steps:
  - name: Regular step
    run: echo "hello"
`,
			importPaths:          []string{".github/workflows/shared.md"},
			mainWorkflowMarkdown: "# Test",
			shouldGenerate:       false,
		},
		{
			name: "skips step when no runtime imports",
			customSteps: `steps:
  - name: Sparse checkout
    run: git sparse-checkout set src
`,
			importPaths:          []string{},
			mainWorkflowMarkdown: "",
			shouldGenerate:       false,
		},
		{
			name:                 "skips step when no custom steps",
			customSteps:          "",
			importPaths:          []string{".github/workflows/shared.md"},
			mainWorkflowMarkdown: "# Test",
			shouldGenerate:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiler := NewCompiler()
			data := &WorkflowData{
				CustomSteps:          tt.customSteps,
				ImportPaths:          tt.importPaths,
				MainWorkflowMarkdown: tt.mainWorkflowMarkdown,
			}

			steps := compiler.generateSparseCheckoutRecoveryStep(data)

			if tt.shouldGenerate {
				require.NotNil(t, steps, "Expected recovery step to be generated")
				assert.Greater(t, len(steps), 0, "Expected non-empty recovery step")

				// Check that the step contains expected content
				stepContent := strings.Join(steps, "")
				assert.Contains(t, stepContent, "Re-checkout .github and .agents",
					"Step should have descriptive name")
				assert.Contains(t, stepContent, "actions/checkout",
					"Step should use actions/checkout")
				assert.Contains(t, stepContent, "sparse-checkout:",
					"Step should use sparse-checkout")
				assert.Contains(t, stepContent, ".github",
					"Step should checkout .github folder")
				assert.Contains(t, stepContent, ".agents",
					"Step should checkout .agents folder")
			} else {
				assert.Nil(t, steps, "Expected no recovery step to be generated")
			}
		})
	}
}

func TestSparseCheckoutRecoveryIntegration(t *testing.T) {
	compiler := NewCompiler()

	data := &WorkflowData{
		Name:        "Test Workflow",
		Description: "Test workflow with sparse checkout",
		AI:          "copilot",
		Permissions: "contents: read",
		CustomSteps: `steps:
  - name: Sparse checkout source files
    run: |
      git sparse-checkout init --cone
      git sparse-checkout set src
`,
		MainWorkflowMarkdown: "# Test Workflow\n\nThis workflow uses sparse checkout to only get source files.",
		ImportPaths:          []string{},
		ParsedTools:          &ToolsConfig{},
	}

	// Generate YAML
	var yaml strings.Builder
	err := compiler.generateMainJobSteps(&yaml, data)
	require.NoError(t, err, "Failed to generate main job steps")

	yamlContent := yaml.String()

	// Verify the recovery step was added
	assert.Contains(t, yamlContent, "Re-checkout .github and .agents after sparse-checkout",
		"Recovery step should be present in generated YAML")
	assert.Contains(t, yamlContent, "sparse-checkout set src",
		"Custom sparse-checkout step should be present")

	// Verify the recovery step comes after custom steps and before prompt creation
	sparseCheckoutIndex := strings.Index(yamlContent, "sparse-checkout set src")
	recoveryIndex := strings.Index(yamlContent, "Re-checkout .github and .agents")
	promptIndex := strings.Index(yamlContent, "Create prompt")

	require.Greater(t, sparseCheckoutIndex, 0, "Custom sparse-checkout step not found")
	require.Greater(t, recoveryIndex, 0, "Recovery step not found")
	require.Greater(t, promptIndex, 0, "Prompt creation step not found")

	assert.Less(t, sparseCheckoutIndex, recoveryIndex,
		"Recovery step should come after custom sparse-checkout step")
	assert.Less(t, recoveryIndex, promptIndex,
		"Recovery step should come before prompt creation")
}
