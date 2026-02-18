//go:build !integration

package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractRoles_OnRoles(t *testing.T) {
	compiler := &Compiler{}

	frontmatter := map[string]any{
		"on": map[string]any{
			"issues": map[string]any{
				"types": []any{"opened"},
			},
			"roles": []any{"admin", "write"},
		},
	}

	roles := compiler.extractRoles(frontmatter)

	assert.Equal(t, []string{"admin", "write"}, roles)
}

func TestExtractRoles_OnRolesString(t *testing.T) {
	compiler := &Compiler{}

	frontmatter := map[string]any{
		"on": map[string]any{
			"issues": map[string]any{
				"types": []any{"opened"},
			},
			"roles": "all",
		},
	}

	roles := compiler.extractRoles(frontmatter)

	assert.Equal(t, []string{"all"}, roles)
}

func TestExtractRoles_TopLevelRoles_Deprecated(t *testing.T) {
	compiler := &Compiler{}

	frontmatter := map[string]any{
		"on": map[string]any{
			"issues": map[string]any{
				"types": []any{"opened"},
			},
		},
		"roles": []any{"admin", "maintainer", "write"},
	}

	roles := compiler.extractRoles(frontmatter)

	assert.Equal(t, []string{"admin", "maintainer", "write"}, roles)
}

func TestExtractRoles_TopLevelRoles_StringArray(t *testing.T) {
	compiler := &Compiler{}

	frontmatter := map[string]any{
		"on": map[string]any{
			"issues": map[string]any{
				"types": []any{"opened"},
			},
		},
		"roles": []string{"admin", "write"},
	}

	roles := compiler.extractRoles(frontmatter)

	assert.Equal(t, []string{"admin", "write"}, roles)
}

func TestExtractRoles_OnRolesPriority(t *testing.T) {
	compiler := &Compiler{}

	// When both on.roles and top-level roles exist, on.roles should take priority
	frontmatter := map[string]any{
		"on": map[string]any{
			"issues": map[string]any{
				"types": []any{"opened"},
			},
			"roles": []any{"admin"},
		},
		"roles": []any{"admin", "maintainer", "write"},
	}

	roles := compiler.extractRoles(frontmatter)

	assert.Equal(t, []string{"admin"}, roles)
}

func TestExtractRoles_Default(t *testing.T) {
	compiler := &Compiler{}

	frontmatter := map[string]any{
		"on": map[string]any{
			"issues": map[string]any{
				"types": []any{"opened"},
			},
		},
	}

	roles := compiler.extractRoles(frontmatter)

	assert.Equal(t, []string{"admin", "maintainer", "write"}, roles)
}

func TestExtractRoles_AllValue(t *testing.T) {
	compiler := &Compiler{}

	frontmatter := map[string]any{
		"on": map[string]any{
			"issues": map[string]any{
				"types": []any{"opened"},
			},
			"roles": "all",
		},
	}

	roles := compiler.extractRoles(frontmatter)

	assert.Equal(t, []string{"all"}, roles)
}
