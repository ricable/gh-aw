//go:build !integration

package cli

import (
	"strings"
	"testing"
)

func TestCreateWorkflowTemplateDefault(t *testing.T) {
	content := createWorkflowTemplate("my-workflow")

	expectedMarkers := []string{
		"# my-workflow",
		"# Trigger - when should this workflow run?",
		"permissions:",
		"safe-outputs:",
		"## Instructions",
		"workflow_dispatch:",
	}

	for _, marker := range expectedMarkers {
		if !strings.Contains(content, marker) {
			t.Errorf("default template missing marker: %s", marker)
		}
	}

	nonExpectedMarkers := []string{
		"promote_to_next_wave:",
		"next_rollout_profile:",
		"reusable-central-deterministic-gate.yml",
		"centralrepoops-",
		"imports:",
	}

	for _, marker := range nonExpectedMarkers {
		if strings.Contains(content, marker) {
			t.Errorf("default template should not contain marker: %s", marker)
		}
	}
}

func TestCreateWorkflowTemplateFilesMainOnly(t *testing.T) {
	files := createWorkflowTemplateFiles("my-workflow", ".github/workflows")
	if len(files) != 1 {
		t.Fatalf("expected 1 file for workflow template, got %d", len(files))
	}
}
