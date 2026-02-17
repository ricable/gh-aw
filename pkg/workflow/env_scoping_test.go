//go:build !integration

package workflow

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnvScopingToAgentJob verifies that env variables from frontmatter
// are scoped to the agent job, not applied globally
func TestEnvScopingToAgentJob(t *testing.T) {
	frontmatter := map[string]any{
		"name":   "Test Env Scoping",
		"on":     "workflow_dispatch",
		"engine": "copilot",
		"env": map[string]any{
			"TEST_VAR":    "test_value",
			"ANOTHER_VAR": "another_value",
		},
	}

	compiler := NewCompiler()

	// Initialize workflow data
	workflowData := &WorkflowData{
		Name: "Test Env Scoping",
		On:   "on:\n  workflow_dispatch:",
		AI:   "copilot",
		EngineConfig: &EngineConfig{
			ID: "copilot",
		},
		MarkdownContent: "# Test content",
	}

	// Extract env map from frontmatter
	compiler.extractYAMLSections(frontmatter, workflowData)

	// Verify EnvMap was populated
	assert.NotNil(t, workflowData.EnvMap, "EnvMap should be populated from frontmatter")
	assert.Len(t, workflowData.EnvMap, 2, "EnvMap should have 2 entries")
	assert.Equal(t, "test_value", workflowData.EnvMap["TEST_VAR"])
	assert.Equal(t, "another_value", workflowData.EnvMap["ANOTHER_VAR"])

	// Build the main job
	job, err := compiler.buildMainJob(workflowData, false)
	require.NoError(t, err, "buildMainJob should succeed")

	// Verify env variables are in the job
	assert.NotNil(t, job.Env, "Job should have env variables")
	assert.Contains(t, job.Env, "TEST_VAR", "Job env should contain TEST_VAR")
	assert.Contains(t, job.Env, "ANOTHER_VAR", "Job env should contain ANOTHER_VAR")
	assert.Equal(t, "test_value", job.Env["TEST_VAR"])
	assert.Equal(t, "another_value", job.Env["ANOTHER_VAR"])

	// Render the job to YAML
	jobManager := NewJobManager()
	err = jobManager.AddJob(job)
	require.NoError(t, err, "AddJob should succeed")

	yamlOutput := jobManager.RenderToYAML()

	// Verify env is at job level, not workflow level
	assert.Contains(t, yamlOutput, "    env:\n", "Job should have env section")
	assert.Contains(t, yamlOutput, "      TEST_VAR: test_value", "Job env should contain TEST_VAR")
	assert.Contains(t, yamlOutput, "      ANOTHER_VAR: another_value", "Job env should contain ANOTHER_VAR")
}

// TestGlobalEnvNotRendered verifies that the global env section is not rendered
// in the workflow YAML output
func TestGlobalEnvNotRendered(t *testing.T) {
	workflowData := &WorkflowData{
		Name: "Test Workflow",
		On:   "on:\n  push:",
		Env:  "env:\n  FOO: bar", // Legacy field, should not be rendered globally
		EnvMap: map[string]string{
			"FOO": "bar",
		},
	}

	compiler := NewCompiler()
	var yamlBuilder strings.Builder

	compiler.generateWorkflowBody(&yamlBuilder, workflowData)
	yamlOutput := yamlBuilder.String()

	// Verify global env is NOT in the output
	// The output should have permissions, concurrency, run-name, but NOT env at the top level
	assert.NotContains(t, yamlOutput, "env:\n  FOO:", "Global env section should not be rendered")
}

// TestEnvMergedWithSafeOutputsEnv verifies that frontmatter env variables
// are merged with safe-outputs env variables at the job level
func TestEnvMergedWithSafeOutputsEnv(t *testing.T) {
	frontmatter := map[string]any{
		"name":   "Test Env Merging",
		"on":     "workflow_dispatch",
		"engine": "copilot",
		"env": map[string]any{
			"CUSTOM_VAR": "custom_value",
		},
		"safe-outputs": map[string]any{
			"create-issue": nil,
		},
	}

	compiler := NewCompiler()

	// Initialize workflow data
	workflowData := &WorkflowData{
		Name: "Test Env Merging",
		On:   "on:\n  workflow_dispatch:",
		AI:   "copilot",
		EngineConfig: &EngineConfig{
			ID: "copilot",
		},
		MarkdownContent: "# Test content",
		SafeOutputs:     compiler.extractSafeOutputsConfig(frontmatter),
	}

	// Extract env map from frontmatter
	compiler.extractYAMLSections(frontmatter, workflowData)

	// Build the main job
	job, err := compiler.buildMainJob(workflowData, false)
	require.NoError(t, err, "buildMainJob should succeed")

	// Verify both frontmatter env and safe-outputs env are present
	assert.NotNil(t, job.Env, "Job should have env variables")

	// Frontmatter env
	assert.Contains(t, job.Env, "CUSTOM_VAR", "Job env should contain custom env from frontmatter")
	assert.Equal(t, "custom_value", job.Env["CUSTOM_VAR"])

	// Safe-outputs env (GH_AW_SAFE_OUTPUTS, etc.)
	assert.Contains(t, job.Env, "GH_AW_SAFE_OUTPUTS", "Job env should contain GH_AW_SAFE_OUTPUTS")
	assert.Contains(t, job.Env, "GH_AW_SAFE_OUTPUTS_CONFIG_PATH", "Job env should contain config path")
}
