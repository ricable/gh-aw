//go:build !integration

package workflow

import (
	"strings"
	"testing"
)

func TestAddLabelsJobWithStagedFlag(t *testing.T) {
	// Create a compiler instance
	c := NewCompiler()

	// Test with staged: true
	workflowData := &WorkflowData{
		Name: "test-workflow",
		SafeOutputs: &SafeOutputsConfig{
			AddLabels: &AddLabelsConfig{},
			Staged:    true,
		},
	}

	job, err := c.buildAddLabelsJob(workflowData, "main_job")
	if err != nil {
		t.Fatalf("Unexpected error building add labels job: %v", err)
	}

	// Convert steps to a single string for testing
	stepsContent := strings.Join(job.Steps, "")

	// Check that GH_AW_SAFE_OUTPUTS_STAGED is included in the env section
	if !strings.Contains(stepsContent, "          GH_AW_SAFE_OUTPUTS_STAGED: \"true\"\n") {
		t.Error("Expected GH_AW_SAFE_OUTPUTS_STAGED environment variable to be set to true in add-labels job")
	}

	// Test with staged: false
	workflowData.SafeOutputs.Staged = false

	job, err = c.buildAddLabelsJob(workflowData, "main_job")
	if err != nil {
		t.Fatalf("Unexpected error building add labels job: %v", err)
	}

	stepsContent = strings.Join(job.Steps, "")

	// Check that GH_AW_SAFE_OUTPUTS_STAGED is not included in the env section when false
	// We need to be specific to avoid matching the JavaScript code that references the variable
	if strings.Contains(stepsContent, "          GH_AW_SAFE_OUTPUTS_STAGED:") {
		t.Error("Expected GH_AW_SAFE_OUTPUTS_STAGED environment variable not to be set when staged is false")
	}

}

func TestAddLabelsJobWithPerHandlerStagedFlag(t *testing.T) {
	// Create a compiler instance
	c := NewCompiler()

	// Test with per-handler staged: true (global staged: false)
	workflowData := &WorkflowData{
		Name: "test-workflow",
		SafeOutputs: &SafeOutputsConfig{
			AddLabels: &AddLabelsConfig{
				BaseSafeOutputConfig: BaseSafeOutputConfig{
					Staged: true,
				},
			},
			Staged: false,
		},
	}

	job, err := c.buildAddLabelsJob(workflowData, "main_job")
	if err != nil {
		t.Fatalf("Unexpected error building add labels job: %v", err)
	}

	// Convert steps to a single string for testing
	stepsContent := strings.Join(job.Steps, "")

	// Check that GH_AW_SAFE_OUTPUTS_STAGED is included when per-handler staged is true
	if !strings.Contains(stepsContent, "          GH_AW_SAFE_OUTPUTS_STAGED: \"true\"\n") {
		t.Error("Expected GH_AW_SAFE_OUTPUTS_STAGED environment variable to be set to true when per-handler staged is true")
	}

	// Test with per-handler staged: false, global staged: true
	workflowData.SafeOutputs.AddLabels.Staged = false
	workflowData.SafeOutputs.Staged = true

	job, err = c.buildAddLabelsJob(workflowData, "main_job")
	if err != nil {
		t.Fatalf("Unexpected error building add labels job: %v", err)
	}

	stepsContent = strings.Join(job.Steps, "")

	// Check that GH_AW_SAFE_OUTPUTS_STAGED is included when global staged is true
	if !strings.Contains(stepsContent, "          GH_AW_SAFE_OUTPUTS_STAGED: \"true\"\n") {
		t.Error("Expected GH_AW_SAFE_OUTPUTS_STAGED environment variable to be set to true when global staged is true")
	}

	// Test with both per-handler and global staged: false
	workflowData.SafeOutputs.AddLabels.Staged = false
	workflowData.SafeOutputs.Staged = false

	job, err = c.buildAddLabelsJob(workflowData, "main_job")
	if err != nil {
		t.Fatalf("Unexpected error building add labels job: %v", err)
	}

	stepsContent = strings.Join(job.Steps, "")

	// Check that GH_AW_SAFE_OUTPUTS_STAGED is not included when both are false
	if strings.Contains(stepsContent, "          GH_AW_SAFE_OUTPUTS_STAGED:") {
		t.Error("Expected GH_AW_SAFE_OUTPUTS_STAGED environment variable not to be set when both staged flags are false")
	}
}

func TestAddLabelsJobWithNilSafeOutputs(t *testing.T) {
	// Create a compiler instance
	c := NewCompiler()

	// Test with no SafeOutputs config - this should fail
	workflowData := &WorkflowData{
		Name:        "test-workflow",
		SafeOutputs: nil,
	}

	_, err := c.buildAddLabelsJob(workflowData, "main_job")
	if err == nil {
		t.Error("Expected error when SafeOutputs is nil")
	}

	expectedError := "safe-outputs configuration is required"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error message to contain '%s', got: %v", expectedError, err)
	}
}
