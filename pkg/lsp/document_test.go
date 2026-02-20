//go:build !integration

package lsp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDocumentStore(t *testing.T) {
	store := NewDocumentStore()
	require.NotNil(t, store, "document store should be created")
}

func TestDocumentStore_OpenAndGet(t *testing.T) {
	store := NewDocumentStore()
	uri := DocumentURI("file:///test.md")

	snap := store.Open(uri, 1, "---\non: issues\n---\n# Test")
	require.NotNil(t, snap, "snapshot should be created on open")
	assert.Equal(t, uri, snap.URI, "URI should match")
	assert.Equal(t, 1, snap.Version, "version should match")

	got := store.Get(uri)
	require.NotNil(t, got, "should retrieve opened document")
	assert.Equal(t, snap, got, "retrieved snapshot should match opened snapshot")
}

func TestDocumentStore_GetNonExistent(t *testing.T) {
	store := NewDocumentStore()
	got := store.Get(DocumentURI("file:///nonexistent.md"))
	assert.Nil(t, got, "should return nil for non-existent document")
}

func TestDocumentStore_Close(t *testing.T) {
	store := NewDocumentStore()
	uri := DocumentURI("file:///test.md")

	store.Open(uri, 1, "---\non: issues\n---")
	store.Close(uri)

	got := store.Get(uri)
	assert.Nil(t, got, "should return nil after close")
}

func TestDocumentStore_Update(t *testing.T) {
	store := NewDocumentStore()
	uri := DocumentURI("file:///test.md")

	store.Open(uri, 1, "---\non: issues\n---")
	snap := store.Update(uri, 2, "---\non: pull_request\n---")

	assert.Equal(t, 2, snap.Version, "version should be updated")
	assert.Contains(t, snap.Text, "pull_request", "text should be updated")
}

func TestParseFrontmatterRegion(t *testing.T) {
	tests := []struct {
		name             string
		text             string
		hasFrontmatter   bool
		frontmatterStart int
		frontmatterEnd   int
		frontmatterYAML  string
	}{
		{
			name:             "valid frontmatter",
			text:             "---\non: issues\nengine: copilot\n---\n# Title",
			hasFrontmatter:   true,
			frontmatterStart: 0,
			frontmatterEnd:   3,
			frontmatterYAML:  "on: issues\nengine: copilot",
		},
		{
			name:           "no frontmatter",
			text:           "# Just Markdown\nSome content",
			hasFrontmatter: false,
		},
		{
			name:           "unclosed frontmatter",
			text:           "---\non: issues\nengine: copilot",
			hasFrontmatter: false,
		},
		{
			name:             "empty frontmatter",
			text:             "---\n---\n# Title",
			hasFrontmatter:   true,
			frontmatterStart: 0,
			frontmatterEnd:   1,
			frontmatterYAML:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snap := newSnapshot(DocumentURI("file:///test.md"), 1, tt.text)
			assert.Equal(t, tt.hasFrontmatter, snap.HasFrontmatter, "HasFrontmatter should match")
			if tt.hasFrontmatter {
				assert.Equal(t, tt.frontmatterStart, snap.FrontmatterStartLine, "FrontmatterStartLine should match")
				assert.Equal(t, tt.frontmatterEnd, snap.FrontmatterEndLine, "FrontmatterEndLine should match")
				assert.Equal(t, tt.frontmatterYAML, snap.FrontmatterYAML, "FrontmatterYAML should match")
			}
		})
	}
}

func TestPositionInFrontmatter(t *testing.T) {
	snap := newSnapshot(DocumentURI("file:///test.md"), 1, "---\non: issues\nengine: copilot\n---\n# Title")

	tests := []struct {
		name     string
		position Position
		inside   bool
	}{
		{"on opening delimiter", Position{Line: 0, Character: 0}, false},
		{"first content line", Position{Line: 1, Character: 0}, true},
		{"second content line", Position{Line: 2, Character: 5}, true},
		{"on closing delimiter", Position{Line: 3, Character: 0}, false},
		{"after frontmatter", Position{Line: 4, Character: 0}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := snap.PositionInFrontmatter(tt.position)
			assert.Equal(t, tt.inside, result, "PositionInFrontmatter should match for %s", tt.name)
		})
	}
}
