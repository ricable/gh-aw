//go:build !integration

package lsp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleHover_InsideFrontmatter(t *testing.T) {
	sp, err := NewSchemaProvider()
	require.NoError(t, err, "NewSchemaProvider should succeed")

	snap := newSnapshot(DocumentURI("file:///test.md"), 1,
		"---\non:\n  issues:\n    types: [opened]\nengine: copilot\n---\n# Title")

	// Hover on "engine" (line 4 in document)
	hover := HandleHover(snap, Position{Line: 4, Character: 2}, sp)
	require.NotNil(t, hover, "should return hover for 'engine' key")
	assert.Equal(t, "markdown", hover.Contents.Kind, "hover should be markdown")
	assert.Contains(t, hover.Contents.Value, "engine", "hover should mention engine")
}

func TestHandleHover_OutsideFrontmatter(t *testing.T) {
	sp, err := NewSchemaProvider()
	require.NoError(t, err, "NewSchemaProvider should succeed")

	snap := newSnapshot(DocumentURI("file:///test.md"), 1,
		"---\non: issues\n---\n# Title")

	// Hover on markdown content (line 3)
	hover := HandleHover(snap, Position{Line: 3, Character: 0}, sp)
	assert.Nil(t, hover, "should return nil for position outside frontmatter")
}

func TestHandleHover_NoFrontmatter(t *testing.T) {
	sp, err := NewSchemaProvider()
	require.NoError(t, err, "NewSchemaProvider should succeed")

	snap := newSnapshot(DocumentURI("file:///test.md"), 1, "# Just Markdown")

	hover := HandleHover(snap, Position{Line: 0, Character: 0}, sp)
	assert.Nil(t, hover, "should return nil when no frontmatter")
}

func TestHandleHover_NilSnapshot(t *testing.T) {
	sp, err := NewSchemaProvider()
	require.NoError(t, err, "NewSchemaProvider should succeed")

	hover := HandleHover(nil, Position{Line: 0, Character: 0}, sp)
	assert.Nil(t, hover, "should return nil for nil snapshot")
}

func TestFormatHoverContent(t *testing.T) {
	info := &PropertyInfo{
		Name:        "engine",
		Description: "AI engine configuration",
		Type:        "string",
		Default:     "copilot",
		Required:    false,
		Enum:        []string{"copilot", "claude", "codex"},
	}

	result := formatHoverContent(info)
	assert.Contains(t, result, "### `engine`", "should contain header")
	assert.Contains(t, result, "AI engine configuration", "should contain description")
	assert.Contains(t, result, "`string`", "should contain type")
	assert.Contains(t, result, "`copilot`", "should contain default")
	assert.Contains(t, result, "`claude`", "should contain enum value")
}

func TestFormatHoverContent_Deprecated(t *testing.T) {
	info := &PropertyInfo{
		Name:        "infer",
		Description: "Deprecated field",
		Deprecated:  true,
	}

	result := formatHoverContent(info)
	assert.Contains(t, result, "Deprecated", "should show deprecated warning")
}

func TestFormatHoverContent_Required(t *testing.T) {
	info := &PropertyInfo{
		Name:     "on",
		Required: true,
	}

	result := formatHoverContent(info)
	assert.Contains(t, result, "Required", "should show required status")
}
