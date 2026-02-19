package cli

import (
	_ "embed"
	"path/filepath"
	"strings"

	"github.com/github/gh-aw/pkg/constants"
)

const (
	workflowNamePlaceholder = "__WORKFLOW_NAME__"
	cliPrefixPlaceholder    = "__CLI_PREFIX__"
)

//go:embed workflowtemplates/default.md.tmpl
var defaultWorkflowTemplate string

func renderWorkflowTemplate(raw string, workflowName string) string {
	replacer := strings.NewReplacer(
		workflowNamePlaceholder, workflowName,
		cliPrefixPlaceholder, string(constants.CLIExtensionPrefix),
	)
	return replacer.Replace(raw)
}

// createWorkflowTemplate generates a concise workflow template with essential options
func createWorkflowTemplate(workflowName string) string {
	return renderWorkflowTemplate(defaultWorkflowTemplate, workflowName)
}

func createWorkflowTemplateFiles(workflowName string, githubWorkflowsDir string) []workflowTemplateFile {
	files := []workflowTemplateFile{
		{Path: filepath.Join(githubWorkflowsDir, workflowName+".md"), Content: createWorkflowTemplate(workflowName), Mode: 0600},
	}

	return files
}
