//go:build !integration

package workflow

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePluginInstallationSteps(t *testing.T) {
	tests := []struct {
		name         string
		plugins      []string
		engineID     string
		githubToken  string
		expectSteps  int
		expectCmds   []string
		expectTokens []string
	}{
		{
			name:         "No plugins",
			plugins:      []string{},
			engineID:     "copilot",
			githubToken:  "",
			expectSteps:  0,
			expectCmds:   []string{},
			expectTokens: []string{},
		},
		{
			name:         "Single plugin for Copilot with custom token",
			plugins:      []string{"github/test-plugin"},
			engineID:     "copilot",
			githubToken:  "${{ secrets.CUSTOM_TOKEN }}",
			expectSteps:  1,
			expectCmds:   []string{"copilot plugin install github/test-plugin"},
			expectTokens: []string{"${{ secrets.CUSTOM_TOKEN }}"},
		},
		{
			name:        "Multiple plugins for Claude with custom token",
			plugins:     []string{"github/plugin1", "acme/plugin2"},
			engineID:    "claude",
			githubToken: "${{ secrets.CUSTOM_TOKEN }}",
			expectSteps: 2,
			expectCmds: []string{
				"claude plugin install github/plugin1",
				"claude plugin install acme/plugin2",
			},
			expectTokens: []string{
				"${{ secrets.CUSTOM_TOKEN }}",
				"${{ secrets.CUSTOM_TOKEN }}",
			},
		},
		{
			name:         "Plugin for Codex with cascading token fallback",
			plugins:      []string{"org/codex-plugin"},
			engineID:     "codex",
			githubToken:  "",
			expectSteps:  0, // Codex doesn't support plugins, should be skipped
			expectCmds:   []string{},
			expectTokens: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			steps := GeneratePluginInstallationSteps(tt.plugins, tt.engineID, tt.githubToken)

			// Verify number of steps
			assert.Len(t, steps, tt.expectSteps, "Number of steps should match")

			// Verify each step
			for i, step := range steps {
				stepText := strings.Join(step, "\n")

				// Verify plugin name in step name (with quotes)
				assert.Contains(t, stepText, fmt.Sprintf("'Install plugin: %s'", tt.plugins[i]),
					"Step should contain quoted plugin name")

				// Verify command
				assert.Contains(t, stepText, tt.expectCmds[i],
					"Step should contain correct install command")

				// Verify GitHub token
				assert.Contains(t, stepText, tt.expectTokens[i],
					"Step should contain correct GitHub token")

				// Verify env section
				assert.Contains(t, stepText, "env:",
					"Step should have env section")
				assert.Contains(t, stepText, "GITHUB_TOKEN:",
					"Step should set GITHUB_TOKEN environment variable")
			}
		})
	}
}

func TestExtractPluginsFromFrontmatter(t *testing.T) {
	tests := []struct {
		name          string
		frontmatter   map[string]any
		expectedRepos []string
		expectedToken string
	}{
		{
			name:          "No plugins field",
			frontmatter:   map[string]any{},
			expectedRepos: nil,
			expectedToken: "",
		},
		{
			name: "Empty plugins array",
			frontmatter: map[string]any{
				"plugins": []any{},
			},
			expectedRepos: nil,
			expectedToken: "",
		},
		{
			name: "Single plugin in array format",
			frontmatter: map[string]any{
				"plugins": []any{"github/test-plugin"},
			},
			expectedRepos: []string{"github/test-plugin"},
			expectedToken: "",
		},
		{
			name: "Multiple plugins in array format",
			frontmatter: map[string]any{
				"plugins": []any{"github/plugin1", "acme/plugin2", "org/plugin3"},
			},
			expectedRepos: []string{"github/plugin1", "acme/plugin2", "org/plugin3"},
			expectedToken: "",
		},
		{
			name: "Mixed types in array (only strings extracted)",
			frontmatter: map[string]any{
				"plugins": []any{"github/plugin1", 123, "acme/plugin2"},
			},
			expectedRepos: []string{"github/plugin1", "acme/plugin2"},
			expectedToken: "",
		},
		{
			name: "Object format with repos only",
			frontmatter: map[string]any{
				"plugins": map[string]any{
					"repos": []any{"github/plugin1", "acme/plugin2"},
				},
			},
			expectedRepos: []string{"github/plugin1", "acme/plugin2"},
			expectedToken: "",
		},
		{
			name: "Object format with repos and custom token",
			frontmatter: map[string]any{
				"plugins": map[string]any{
					"repos":        []any{"github/plugin1", "acme/plugin2"},
					"github-token": "${{ secrets.CUSTOM_PLUGIN_TOKEN }}",
				},
			},
			expectedRepos: []string{"github/plugin1", "acme/plugin2"},
			expectedToken: "${{ secrets.CUSTOM_PLUGIN_TOKEN }}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pluginInfo := extractPluginsFromFrontmatter(tt.frontmatter)
			var repos []string
			var token string
			if pluginInfo != nil {
				repos = pluginInfo.Plugins
				token = pluginInfo.CustomToken
			}
			assert.Equal(t, tt.expectedRepos, repos, "Extracted plugin repos should match expected")
			assert.Equal(t, tt.expectedToken, token, "Extracted plugin token should match expected")
		})
	}
}

func TestPluginInstallationIntegration(t *testing.T) {
	// Test that plugins are properly integrated into engine installation steps
	engines := []struct {
		engineID string
		engine   CodingAgentEngine
	}{
		{"copilot", NewCopilotEngine()},
		{"claude", NewClaudeEngine()},
		{"codex", NewCodexEngine()},
	}

	for _, e := range engines {
		t.Run(e.engineID, func(t *testing.T) {
			// Create workflow data with plugins
			workflowData := &WorkflowData{
				Name: "test-workflow",
				PluginInfo: &PluginInfo{
					Plugins: []string{"github/test-plugin"},
				},
			}

			// Get installation steps
			steps := e.engine.GetInstallationSteps(workflowData)

			// Convert steps to string for searching
			var allStepsText string
			for _, step := range steps {
				allStepsText += strings.Join(step, "\n") + "\n"
			}

			// For Codex engine, plugins should be skipped (not supported)
			if e.engineID == "codex" {
				assert.NotContains(t, allStepsText, fmt.Sprintf("%s plugin install github/test-plugin", e.engineID),
					"Codex engine should not generate plugin installation commands")
				return
			}

			// Verify plugin installation step is present for other engines
			assert.Contains(t, allStepsText, fmt.Sprintf("%s plugin install github/test-plugin", e.engineID),
				"Installation steps should include plugin installation command")

			// Verify GITHUB_TOKEN is set
			assert.Contains(t, allStepsText, "GITHUB_TOKEN:",
				"Plugin installation should have GITHUB_TOKEN environment variable")

			// Verify cascading token is used when no custom token provided
			assert.Contains(t, allStepsText, "secrets.GH_AW_PLUGINS_TOKEN || secrets.GH_AW_GITHUB_TOKEN || secrets.GITHUB_TOKEN",
				"Plugin installation should use cascading token when no custom token provided")
		})
	}
}

func TestPluginTokenCascading(t *testing.T) {
	tests := []struct {
		name          string
		customToken   string
		expectedToken string
	}{
		{
			name:          "Custom token provided",
			customToken:   "${{ secrets.CUSTOM_PLUGIN_TOKEN }}",
			expectedToken: "${{ secrets.CUSTOM_PLUGIN_TOKEN }}",
		},
		{
			name:          "No custom token - uses cascading fallback",
			customToken:   "",
			expectedToken: "${{ secrets.GH_AW_PLUGINS_TOKEN || secrets.GH_AW_GITHUB_TOKEN || secrets.GITHUB_TOKEN }}",
		},
		{
			name:          "Frontmatter github-token provided",
			customToken:   "${{ secrets.MY_GITHUB_TOKEN }}",
			expectedToken: "${{ secrets.MY_GITHUB_TOKEN }}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getEffectivePluginGitHubToken(tt.customToken)
			assert.Equal(t, tt.expectedToken, result, "Token resolution should match expected")
		})
	}
}

func TestPluginObjectFormatWithCustomToken(t *testing.T) {
	// Test that object format with custom token overrides cascading resolution
	engines := []struct {
		engineID string
		engine   CodingAgentEngine
	}{
		{"copilot", NewCopilotEngine()},
		{"claude", NewClaudeEngine()},
		{"codex", NewCodexEngine()},
	}

	for _, e := range engines {
		t.Run(e.engineID, func(t *testing.T) {
			// Create workflow data with plugins and custom token
			workflowData := &WorkflowData{
				Name: "test-workflow",
				PluginInfo: &PluginInfo{
					Plugins:     []string{"github/test-plugin"},
					CustomToken: "${{ secrets.CUSTOM_PLUGIN_TOKEN }}",
				},
			}

			// Get installation steps
			steps := e.engine.GetInstallationSteps(workflowData)

			// Convert steps to string for searching
			var allStepsText string
			for _, step := range steps {
				allStepsText += strings.Join(step, "\n") + "\n"
			}

			// For Codex engine, plugins should be skipped (not supported)
			if e.engineID == "codex" {
				assert.NotContains(t, allStepsText, fmt.Sprintf("%s plugin install github/test-plugin", e.engineID),
					"Codex engine should not generate plugin installation commands")
				return
			}

			// Verify plugin installation step is present for other engines
			assert.Contains(t, allStepsText, fmt.Sprintf("%s plugin install github/test-plugin", e.engineID),
				"Installation steps should include plugin installation command")

			// Verify custom token is used (not the cascading fallback)
			assert.Contains(t, allStepsText, "GITHUB_TOKEN: ${{ secrets.CUSTOM_PLUGIN_TOKEN }}",
				"Plugin installation should use custom token when provided")

			// Verify cascading token is NOT used
			assert.NotContains(t, allStepsText, "secrets.GH_AW_PLUGINS_TOKEN",
				"Plugin installation should not use cascading token when custom token is provided")
		})
	}
}

func TestValidatePluginForEngine(t *testing.T) {
	tests := []struct {
		name        string
		plugin      string
		engineID    string
		expectErr   bool
		errContains string
	}{
		{
			name:      "GitHub plugin works with any engine",
			plugin:    "github/test-plugin",
			engineID:  "copilot",
			expectErr: false,
		},
		{
			name:        "Regular plugin with Codex engine fails",
			plugin:      "github/test-plugin",
			engineID:    "codex",
			expectErr:   true,
			errContains: "codex engine does not support plugin install",
		},
		{
			name:        "Codex plugin with Codex engine fails",
			plugin:      "org/codex-plugin",
			engineID:    "codex",
			expectErr:   true,
			errContains: "codex engine does not support plugin install",
		},
		{
			name:      "Org/repo format works with any engine",
			plugin:    "acme/my-plugin",
			engineID:  "claude",
			expectErr: false,
		},
		{
			name:        "Claude marketplace plugin with Codex engine fails",
			plugin:      "explanatory-output-style@claude-plugins-official",
			engineID:    "codex",
			expectErr:   true,
			errContains: "codex engine does not support plugin install",
		},
		{
			name:        "Claude marketplace plugin with Copilot engine fails",
			plugin:      "explanatory-output-style@claude-plugins-official",
			engineID:    "copilot",
			expectErr:   true,
			errContains: "Claude marketplace",
		},
		{
			name:      "Claude marketplace plugin with Claude engine works",
			plugin:    "explanatory-output-style@claude-plugins-official",
			engineID:  "claude",
			expectErr: false,
		},
		{
			name:        "Copilot marketplace plugin with Claude engine fails",
			plugin:      "some-plugin@copilot-plugins-official",
			engineID:    "claude",
			expectErr:   true,
			errContains: "Copilot marketplace",
		},
		{
			name:      "Copilot marketplace plugin with Copilot engine works",
			plugin:    "some-plugin@copilot-plugins-official",
			engineID:  "copilot",
			expectErr: false,
		},
		{
			name:      "Unknown marketplace shows warning but allows",
			plugin:    "some-plugin@unknown-marketplace",
			engineID:  "copilot",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePluginForEngine(tt.plugin, tt.engineID)

			if tt.expectErr {
				require.Error(t, err, "Expected validation to fail")
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains, "Error should contain expected text")
				}
			} else {
				assert.NoError(t, err, "Expected validation to pass")
			}
		})
	}
}

func TestGeneratePluginInstallationStepsSkipsIncompatible(t *testing.T) {
	// Test that incompatible plugins are skipped
	tests := []struct {
		name        string
		plugins     []string
		engineID    string
		expectSteps int
		expectCmds  []string
	}{
		{
			name:        "Claude marketplace plugin skipped for Codex",
			plugins:     []string{"explanatory-output-style@claude-plugins-official"},
			engineID:    "codex",
			expectSteps: 0,
			expectCmds:  []string{},
		},
		{
			name:        "Claude marketplace plugin skipped for Copilot",
			plugins:     []string{"explanatory-output-style@claude-plugins-official"},
			engineID:    "copilot",
			expectSteps: 0,
			expectCmds:  []string{},
		},
		{
			name:        "Claude marketplace plugin works for Claude",
			plugins:     []string{"explanatory-output-style@claude-plugins-official"},
			engineID:    "claude",
			expectSteps: 1,
			expectCmds:  []string{"claude plugin install explanatory-output-style@claude-plugins-official"},
		},
		{
			name:        "Mixed plugins - compatible ones processed, incompatible skipped",
			plugins:     []string{"github/valid-plugin", "invalid@claude-plugins-official", "acme/another"},
			engineID:    "copilot",
			expectSteps: 2,
			expectCmds:  []string{"copilot plugin install github/valid-plugin", "copilot plugin install acme/another"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			steps := GeneratePluginInstallationSteps(tt.plugins, tt.engineID, "")

			assert.Len(t, steps, tt.expectSteps, "Number of steps should match")

			for i, step := range steps {
				stepText := strings.Join(step, "\n")
				assert.Contains(t, stepText, tt.expectCmds[i], "Step should contain expected command")
			}
		})
	}
}
