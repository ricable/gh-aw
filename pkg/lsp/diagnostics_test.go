//go:build !integration

package lsp

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeDiagnostics_ValidDocument(t *testing.T) {
	snap := newSnapshot(DocumentURI("file:///test.md"), 1,
		"---\non:\n  issues:\n    types: [opened]\nengine: copilot\n---\n# Title")

	diags := ComputeDiagnostics(snap)
	assert.Empty(t, diags, "valid document should produce no diagnostics")
}

func TestComputeDiagnostics_MissingFrontmatter(t *testing.T) {
	snap := newSnapshot(DocumentURI("file:///test.md"), 1, "# Just Markdown")

	diags := ComputeDiagnostics(snap)
	require.Len(t, diags, 1, "should produce one diagnostic")
	assert.Equal(t, SeverityWarning, diags[0].Severity, "should be a warning")
	assert.Contains(t, diags[0].Message, "missing frontmatter", "message should mention missing frontmatter")
}

func TestComputeDiagnostics_YAMLSyntaxError(t *testing.T) {
	// Invalid YAML indentation
	snap := newSnapshot(DocumentURI("file:///test.md"), 1,
		"---\non:\n  issues:\n    types: [opened\n---\n# Title")

	diags := ComputeDiagnostics(snap)
	require.NotEmpty(t, diags, "should produce diagnostics for YAML error")
	assert.Equal(t, SeverityError, diags[0].Severity, "should be an error")
	assert.Contains(t, diags[0].Message, "YAML syntax error", "message should mention YAML syntax error")
}

func TestComputeDiagnostics_MissingRequiredField(t *testing.T) {
	// Missing required "on" field
	snap := newSnapshot(DocumentURI("file:///test.md"), 1,
		"---\nengine: copilot\n---\n# Title")

	diags := ComputeDiagnostics(snap)
	require.NotEmpty(t, diags, "should produce diagnostics for missing 'on'")
	assert.Equal(t, SeverityError, diags[0].Severity, "should be an error")
	assert.Contains(t, diags[0].Message, "on", "message should mention 'on'")
}

func TestComputeDiagnostics_MultipleFrontmatterBlocks(t *testing.T) {
	snap := newSnapshot(DocumentURI("file:///test.md"), 1,
		"---\non:\n  issues:\n    types: [opened]\n---\n# Title\n---\nmore stuff\n---")

	diags := ComputeDiagnostics(snap)
	// Should have a warning about multiple frontmatter blocks
	hasMultipleWarning := false
	for _, d := range diags {
		if d.Severity == SeverityWarning && strings.Contains(d.Message, "Multiple frontmatter") {
			hasMultipleWarning = true
			break
		}
	}
	assert.True(t, hasMultipleWarning, "should warn about multiple frontmatter blocks")
}

func TestComputeDiagnostics_EmptyFrontmatter(t *testing.T) {
	snap := newSnapshot(DocumentURI("file:///test.md"), 1, "---\n---\n# Title")

	diags := ComputeDiagnostics(snap)
	// Should produce a diagnostic for missing "on" field
	require.NotEmpty(t, diags, "empty frontmatter should produce diagnostics for missing 'on'")
}

func TestCleanSchemaErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "missing property",
			input:    "'http://contoso.com/main-workflow-schema.json#'\n- at '': missing property 'on'",
			expected: "missing property 'on'",
		},
		{
			name:     "simple error passthrough",
			input:    "invalid value for engine",
			expected: "invalid value for engine",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanSchemaErrorMessage(tt.input)
			assert.Equal(t, tt.expected, result, "cleaned message should match")
		})
	}
}

func TestExtractYAMLErrorLine(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected int
	}{
		{
			name:     "line number present",
			errMsg:   "yaml: line 3: could not find expected ':'",
			expected: 3,
		},
		{
			name:     "no line number",
			errMsg:   "yaml: unexpected end of stream",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractYAMLErrorLine(tt.errMsg)
			assert.Equal(t, tt.expected, result, "extracted line should match")
		})
	}
}
