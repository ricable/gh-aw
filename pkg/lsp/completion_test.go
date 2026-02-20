//go:build !integration

package lsp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleCompletion_TopLevel(t *testing.T) {
	sp, err := NewSchemaProvider()
	require.NoError(t, err, "NewSchemaProvider should succeed")

	snap := newSnapshot(DocumentURI("file:///test.md"), 1,
		"---\non:\n  issues:\n    types: [opened]\nengine: copilot\n---\n# Title")

	// Completion at top-level (line 4 = "engine: copilot")
	list := HandleCompletion(snap, Position{Line: 4, Character: 0}, sp)
	require.NotNil(t, list, "should return completion list")
	assert.NotEmpty(t, list.Items, "should have completion items")

	// Should include top-level property or enum suggestions + snippets
	hasEngine := false
	for _, item := range list.Items {
		if item.Label == "engine" {
			hasEngine = true
			break
		}
	}
	assert.True(t, hasEngine, "should include 'engine' in completions")
}

func TestHandleCompletion_NestedUnderOn(t *testing.T) {
	sp, err := NewSchemaProvider()
	require.NoError(t, err, "NewSchemaProvider should succeed")

	snap := newSnapshot(DocumentURI("file:///test.md"), 1,
		"---\non:\n  issues:\n    types: [opened]\nengine: copilot\n---\n# Title")

	// Completion inside "on:" block (line 2 = "  issues:")
	list := HandleCompletion(snap, Position{Line: 2, Character: 2}, sp)
	require.NotNil(t, list, "should return completion list")
	assert.NotEmpty(t, list.Items, "should have nested completion items")

	// Should include event trigger types
	labels := make(map[string]bool)
	for _, item := range list.Items {
		labels[item.Label] = true
	}
	assert.True(t, labels["issues"], "should suggest 'issues'")
	assert.True(t, labels["pull_request"], "should suggest 'pull_request'")
}

func TestHandleCompletion_NoFrontmatter(t *testing.T) {
	sp, err := NewSchemaProvider()
	require.NoError(t, err, "NewSchemaProvider should succeed")

	snap := newSnapshot(DocumentURI("file:///test.md"), 1, "# Just Markdown")

	list := HandleCompletion(snap, Position{Line: 0, Character: 0}, sp)
	require.NotNil(t, list, "should return completion list")

	// Should suggest snippets
	hasSnippet := false
	for _, item := range list.Items {
		if item.Kind == CompletionItemKindSnippet {
			hasSnippet = true
			break
		}
	}
	assert.True(t, hasSnippet, "should include snippet completions when no frontmatter")
}

func TestHandleCompletion_OutsideFrontmatter(t *testing.T) {
	sp, err := NewSchemaProvider()
	require.NoError(t, err, "NewSchemaProvider should succeed")

	snap := newSnapshot(DocumentURI("file:///test.md"), 1,
		"---\non: issues\n---\n# Title")

	// Completion outside frontmatter (line 3 = "# Title")
	list := HandleCompletion(snap, Position{Line: 3, Character: 0}, sp)
	require.NotNil(t, list, "should return completion list")
	assert.Empty(t, list.Items, "should have no items outside frontmatter")
}

func TestHandleCompletion_NilSnapshot(t *testing.T) {
	sp, err := NewSchemaProvider()
	require.NoError(t, err, "NewSchemaProvider should succeed")

	list := HandleCompletion(nil, Position{Line: 0, Character: 0}, sp)
	require.NotNil(t, list, "should return empty completion list for nil snapshot")
	assert.Empty(t, list.Items, "should have no items for nil snapshot")
}

func TestSnippetCompletions(t *testing.T) {
	snippets := snippetCompletions()
	assert.NotEmpty(t, snippets, "should have snippet completions")

	for _, s := range snippets {
		assert.Equal(t, CompletionItemKindSnippet, s.Kind, "all snippets should have snippet kind")
		assert.Equal(t, InsertTextFormatSnippet, s.InsertTextFormat, "all snippets should use snippet format")
		assert.NotEmpty(t, s.InsertText, "all snippets should have insert text")
	}
}

func TestFilterCompletions(t *testing.T) {
	items := []CompletionItem{
		{Label: "engine"},
		{Label: "env"},
		{Label: "on"},
		{Label: "imports"},
	}

	filtered := filterCompletions(items, "en")
	assert.Len(t, filtered, 2, "should filter to items starting with 'en'")

	labels := make([]string, len(filtered))
	for i, item := range filtered {
		labels[i] = item.Label
	}
	assert.Contains(t, labels, "engine", "should include 'engine'")
	assert.Contains(t, labels, "env", "should include 'env'")
}

func TestFilterCompletions_EmptyPrefix(t *testing.T) {
	items := []CompletionItem{
		{Label: "engine"},
		{Label: "on"},
	}

	filtered := filterCompletions(items, "")
	assert.Len(t, filtered, len(items), "empty prefix should return all items")
}
