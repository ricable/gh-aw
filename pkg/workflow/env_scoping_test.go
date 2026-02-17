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

// TestEnvNonStringValues verifies that non-string env values are converted to strings
func TestEnvNonStringValues(t *testing.T) {
	frontmatter := map[string]any{
		"name":   "Test Non-String Env",
		"on":     "workflow_dispatch",
		"engine": "copilot",
		"env": map[string]any{
			"DEBUG_MODE":  true,   // boolean
			"PORT":        3000,   // number
			"MAX_RETRIES": 5,      // number
			"STRING_VAR":  "test", // string
		},
	}

	compiler := NewCompiler()

	workflowData := &WorkflowData{
		Name: "Test Non-String Env",
		On:   "on:\n  workflow_dispatch:",
		AI:   "copilot",
		EngineConfig: &EngineConfig{
			ID: "copilot",
		},
		MarkdownContent: "# Test content",
	}

	// Extract env map from frontmatter
	compiler.extractYAMLSections(frontmatter, workflowData)

	// Verify all types were converted to strings
	assert.NotNil(t, workflowData.EnvMap, "EnvMap should be populated")
	assert.Equal(t, "true", workflowData.EnvMap["DEBUG_MODE"], "Boolean should be converted to string")
	assert.Equal(t, "3000", workflowData.EnvMap["PORT"], "Number should be converted to string")
	assert.Equal(t, "5", workflowData.EnvMap["MAX_RETRIES"], "Number should be converted to string")
	assert.Equal(t, "test", workflowData.EnvMap["STRING_VAR"], "String should remain unchanged")

	// Build the main job
	job, err := compiler.buildMainJob(workflowData, false)
	require.NoError(t, err, "buildMainJob should succeed")

	// Verify all converted values are in the job
	assert.Equal(t, "true", job.Env["DEBUG_MODE"])
	assert.Equal(t, "3000", job.Env["PORT"])
	assert.Equal(t, "5", job.Env["MAX_RETRIES"])
	assert.Equal(t, "test", job.Env["STRING_VAR"])
}

// TestEnvReservedNamesProtection verifies that reserved system variable names
// cannot be overridden by user-defined env variables
func TestEnvReservedNamesProtection(t *testing.T) {
	frontmatter := map[string]any{
		"name":   "Test Reserved Names",
		"on":     "workflow_dispatch",
		"engine": "copilot",
		"env": map[string]any{
			"CUSTOM_VAR":         "allowed",
			"GH_AW_SAFE_OUTPUTS": "should_be_ignored", // Reserved
			"GH_AW_WORKFLOW_ID":  "should_be_ignored", // Reserved
			"DEFAULT_BRANCH":     "should_be_ignored", // Reserved
			"GH_AW_CUSTOM":       "should_be_ignored", // Reserved (GH_AW_ prefix)
		},
	}

	compiler := NewCompiler()

	workflowData := &WorkflowData{
		Name: "Test Reserved Names",
		On:   "on:\n  workflow_dispatch:",
		AI:   "copilot",
		EngineConfig: &EngineConfig{
			ID: "copilot",
		},
		MarkdownContent: "# Test content",
		WorkflowID:      "test-workflow", // This will generate GH_AW_WORKFLOW_ID_SANITIZED
	}

	// Extract env map from frontmatter
	compiler.extractYAMLSections(frontmatter, workflowData)

	// Build the main job
	job, err := compiler.buildMainJob(workflowData, false)
	require.NoError(t, err, "buildMainJob should succeed")

	// Verify allowed variable is present
	assert.Contains(t, job.Env, "CUSTOM_VAR", "Non-reserved var should be present")
	assert.Equal(t, "allowed", job.Env["CUSTOM_VAR"])

	// Verify reserved variables are NOT overridden by user values
	// Instead, they should either not exist or have system values
	if val, exists := job.Env["GH_AW_SAFE_OUTPUTS"]; exists {
		assert.NotEqual(t, "should_be_ignored", val, "GH_AW_SAFE_OUTPUTS should not be overridden")
	}
	if val, exists := job.Env["DEFAULT_BRANCH"]; exists {
		assert.NotEqual(t, "should_be_ignored", val, "DEFAULT_BRANCH should not be overridden")
	}

	// Verify system-generated variables are present with correct values
	assert.Contains(t, job.Env, "GH_AW_WORKFLOW_ID_SANITIZED", "System var should be present")
	assert.Equal(t, "testworkflow", job.Env["GH_AW_WORKFLOW_ID_SANITIZED"], "System var should have correct value")
}

// TestEnvVariableOrdering verifies that env variables are rendered in stable alphabetical order
func TestEnvVariableOrdering(t *testing.T) {
	frontmatter := map[string]any{
		"name":   "Test Env Ordering",
		"on":     "workflow_dispatch",
		"engine": "copilot",
		"env": map[string]any{
			"ZEBRA":   "last",
			"ALPHA":   "first",
			"MIDDLE":  "middle",
			"BETA":    "second",
		},
	}

	compiler := NewCompiler()

	workflowData := &WorkflowData{
		Name: "Test Env Ordering",
		On:   "on:\n  workflow_dispatch:",
		AI:   "copilot",
		EngineConfig: &EngineConfig{
			ID: "copilot",
		},
		MarkdownContent: "# Test content",
		WorkflowID:      "test-workflow",
	}

	// Extract env map from frontmatter
	compiler.extractYAMLSections(frontmatter, workflowData)

	// Build the main job
	job, err := compiler.buildMainJob(workflowData, false)
	require.NoError(t, err, "buildMainJob should succeed")

	// Render to YAML
	jobManager := NewJobManager()
	err = jobManager.AddJob(job)
	require.NoError(t, err, "AddJob should succeed")

	yamlOutput := jobManager.RenderToYAML()

	// Extract the env section to verify ordering
	lines := strings.Split(yamlOutput, "\n")
	var envLines []string
	inEnvSection := false
	for _, line := range lines {
		if strings.Contains(line, "    env:") {
			inEnvSection = true
			continue
		}
		if inEnvSection {
			if strings.HasPrefix(line, "      ") && strings.Contains(line, ":") {
				envLines = append(envLines, line)
			} else if !strings.HasPrefix(line, "      ") {
				break
			}
		}
	}

	// Verify we have env lines
	require.Greater(t, len(envLines), 0, "Should have env variables in YAML output")

	// Verify alphabetical ordering
	// Expected order: ALPHA, BETA, GH_AW_WORKFLOW_ID_SANITIZED, MIDDLE, ZEBRA
	assert.Contains(t, envLines[0], "ALPHA:", "First env var should be ALPHA (alphabetically first user var)")
	assert.Contains(t, envLines[1], "BETA:", "Second env var should be BETA")
	assert.Contains(t, envLines[2], "GH_AW_WORKFLOW_ID_SANITIZED:", "Third should be GH_AW_WORKFLOW_ID_SANITIZED")
	assert.Contains(t, envLines[3], "MIDDLE:", "Fourth env var should be MIDDLE")
	assert.Contains(t, envLines[4], "ZEBRA:", "Fifth env var should be ZEBRA")

	// Verify stable ordering by compiling multiple times
	for i := 0; i < 5; i++ {
		jobManager2 := NewJobManager()
		err = jobManager2.AddJob(job)
		require.NoError(t, err)
		yamlOutput2 := jobManager2.RenderToYAML()
		assert.Equal(t, yamlOutput, yamlOutput2, "YAML output should be identical across multiple renderings (stable ordering)")
	}
}
