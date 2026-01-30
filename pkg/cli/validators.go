package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/githubnext/gh-aw/pkg/logger"
)

var validatorsLog = logger.New("cli:validators")

// workflowNameRegex validates workflow names contain only alphanumeric characters, hyphens, and underscores
var workflowNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// ValidateWorkflowName checks if the provided workflow name is valid.
// It ensures the name is not empty and contains only alphanumeric characters, hyphens, and underscores.
func ValidateWorkflowName(s string) error {
	validatorsLog.Printf("Validating workflow name: %s", s)
	if s == "" {
		validatorsLog.Print("Workflow name validation failed: empty name")
		return errors.New("workflow name cannot be empty")
	}
	if !workflowNameRegex.MatchString(s) {
		validatorsLog.Printf("Workflow name validation failed: invalid characters in %s", s)
		return errors.New("workflow name must contain only alphanumeric characters, hyphens, and underscores")
	}
	validatorsLog.Printf("Workflow name validated successfully: %s", s)
	return nil
}

// ValidateWorkflowIntent checks if the provided workflow intent is valid.
// It ensures the intent has meaningful content with at least 20 characters
// and is not just whitespace.
func ValidateWorkflowIntent(s string) error {
	validatorsLog.Printf("Validating workflow intent: length=%d", len(s))
	trimmed := strings.TrimSpace(s)
	if len(trimmed) == 0 {
		validatorsLog.Print("Workflow intent validation failed: empty content")
		return errors.New("workflow instructions cannot be empty")
	}
	if len(trimmed) < 20 {
		validatorsLog.Printf("Workflow intent validation failed: too short (%d chars)", len(trimmed))
		return errors.New("please provide at least 20 characters of instructions")
	}
	validatorsLog.Printf("Workflow intent validated successfully: %d chars", len(trimmed))
	return nil
}

// validateActionYml validates that an action.yml file exists and contains required fields.
//
// This validates GitHub Actions custom action structure:
//   - action.yml file exists in the action directory
//   - Required fields are present (name, description, runs)
//   - Runtime is either 'node20' or 'composite'
func validateActionYml(actionPath string) error {
	ymlPath := filepath.Join(actionPath, "action.yml")

	if _, err := os.Stat(ymlPath); os.IsNotExist(err) {
		return fmt.Errorf("action.yml not found")
	}

	content, err := os.ReadFile(ymlPath)
	if err != nil {
		return fmt.Errorf("failed to read action.yml: %w", err)
	}

	contentStr := string(content)

	// Check required fields
	requiredFields := []string{"name:", "description:", "runs:"}
	for _, field := range requiredFields {
		if !strings.Contains(contentStr, field) {
			return fmt.Errorf("missing required field '%s'", strings.TrimSuffix(field, ":"))
		}
	}

	// Check that it's either a node20 or composite action
	isNode20 := strings.Contains(contentStr, "using: 'node20'") || strings.Contains(contentStr, "using: \"node20\"")
	isComposite := strings.Contains(contentStr, "using: 'composite'") || strings.Contains(contentStr, "using: \"composite\"")

	if !isNode20 && !isComposite {
		return fmt.Errorf("action must use either 'node20' or 'composite' runtime")
	}

	return nil
}
