//go:build !integration

package lsp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSchemaProvider(t *testing.T) {
	sp, err := NewSchemaProvider()
	require.NoError(t, err, "NewSchemaProvider should succeed")
	require.NotNil(t, sp, "schema provider should not be nil")
}

func TestSchemaProvider_TopLevelProperties(t *testing.T) {
	sp, err := NewSchemaProvider()
	require.NoError(t, err, "NewSchemaProvider should succeed")

	props := sp.TopLevelProperties()
	assert.NotEmpty(t, props, "should return top-level properties")

	// Check for required well-known properties
	names := make(map[string]bool)
	for _, p := range props {
		names[p.Name] = true
	}

	assert.True(t, names["on"], "should contain 'on' property")
	assert.True(t, names["engine"], "should contain 'engine' property")
	assert.True(t, names["imports"], "should contain 'imports' property")
	assert.True(t, names["tools"], "should contain 'tools' property")
	assert.True(t, names["safe-outputs"], "should contain 'safe-outputs' property")
}

func TestSchemaProvider_NestedProperties_On(t *testing.T) {
	sp, err := NewSchemaProvider()
	require.NoError(t, err, "NewSchemaProvider should succeed")

	props := sp.NestedProperties([]string{"on"})
	assert.NotEmpty(t, props, "should return nested properties for 'on'")

	names := make(map[string]bool)
	for _, p := range props {
		names[p.Name] = true
	}

	assert.True(t, names["issues"], "should contain 'issues' trigger")
	assert.True(t, names["pull_request"], "should contain 'pull_request' trigger")
	assert.True(t, names["slash_command"], "should contain 'slash_command' trigger")
}

func TestSchemaProvider_PropertyDescription(t *testing.T) {
	sp, err := NewSchemaProvider()
	require.NoError(t, err, "NewSchemaProvider should succeed")

	info := sp.PropertyDescription([]string{"on"})
	require.NotNil(t, info, "should return property info for 'on'")
	assert.Equal(t, "on", info.Name, "name should be 'on'")
	assert.NotEmpty(t, info.Description, "description should not be empty")

	info = sp.PropertyDescription([]string{"engine"})
	require.NotNil(t, info, "should return property info for 'engine'")
	assert.NotEmpty(t, info.Description, "engine description should not be empty")
}

func TestSchemaProvider_PropertyDescription_NotFound(t *testing.T) {
	sp, err := NewSchemaProvider()
	require.NoError(t, err, "NewSchemaProvider should succeed")

	info := sp.PropertyDescription([]string{"nonexistent-property"})
	assert.Nil(t, info, "should return nil for unknown property")
}

func TestSchemaProvider_EnumValues(t *testing.T) {
	sp, err := NewSchemaProvider()
	require.NoError(t, err, "NewSchemaProvider should succeed")

	enums := sp.EnumValues([]string{"engine"})
	assert.NotEmpty(t, enums, "should return enum values for 'engine'")
	assert.Contains(t, enums, "copilot", "should contain 'copilot'")
	assert.Contains(t, enums, "claude", "should contain 'claude'")
}

func TestYAMLPathAtPosition(t *testing.T) {
	tests := []struct {
		name        string
		yaml        string
		line        int
		wantPath    []string
		wantCurrent string
	}{
		{
			name:        "top-level key",
			yaml:        "on:\n  issues:\n    types: [opened]\nengine: copilot",
			line:        0,
			wantPath:    nil,
			wantCurrent: "on",
		},
		{
			name:        "nested key under on",
			yaml:        "on:\n  issues:\n    types: [opened]\nengine: copilot",
			line:        1,
			wantPath:    []string{"on"},
			wantCurrent: "issues",
		},
		{
			name:        "deeply nested key",
			yaml:        "on:\n  issues:\n    types: [opened]\nengine: copilot",
			line:        2,
			wantPath:    []string{"on", "issues"},
			wantCurrent: "types",
		},
		{
			name:        "second top-level key",
			yaml:        "on:\n  issues:\n    types: [opened]\nengine: copilot",
			line:        3,
			wantPath:    nil,
			wantCurrent: "engine",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, current := YAMLPathAtPosition(tt.yaml, tt.line)
			if tt.wantPath == nil {
				assert.Empty(t, path, "path should be empty for %s", tt.name)
			} else {
				assert.Equal(t, tt.wantPath, path, "path should match for %s", tt.name)
			}
			assert.Equal(t, tt.wantCurrent, current, "current key should match for %s", tt.name)
		})
	}
}

func TestYAMLPathAtPosition_EmptyContent(t *testing.T) {
	path, key := YAMLPathAtPosition("", 0)
	assert.Nil(t, path, "path should be nil for empty content")
	assert.Empty(t, key, "key should be empty for empty content")
}

func TestYAMLPathAtPosition_InvalidYAML(t *testing.T) {
	// Invalid YAML should fallback to indentation heuristic
	yaml := "on:\n  issues\n    types: [opened"
	path, key := YAMLPathAtPosition(yaml, 0)
	// Should still work via fallback
	assert.Equal(t, "on", key, "should detect key via fallback")
	assert.Empty(t, path, "path should be empty at top level")
}
