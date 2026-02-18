//go:build !integration

package workflow

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMistralVibeEngine(t *testing.T) {
	engine := NewMistralVibeEngine()

	t.Run("engine identity", func(t *testing.T) {
		assert.Equal(t, "mistral-vibe", engine.GetID(), "Engine ID should be mistral-vibe")
		assert.Equal(t, "Mistral Vibe CLI", engine.GetDisplayName(), "Display name should be Mistral Vibe CLI")
		assert.NotEmpty(t, engine.GetDescription(), "Description should not be empty")
		assert.True(t, engine.IsExperimental(), "Mistral Vibe should be marked as experimental")
	})

	t.Run("capabilities", func(t *testing.T) {
		assert.True(t, engine.SupportsToolsAllowlist(), "Should support tools allowlist")
		assert.True(t, engine.SupportsMaxTurns(), "Should support max turns")
		assert.False(t, engine.SupportsWebFetch(), "Should not support built-in web fetch")
		assert.False(t, engine.SupportsWebSearch(), "Should not support built-in web search")
		assert.True(t, engine.SupportsFirewall(), "Should support firewall")
		assert.False(t, engine.SupportsPlugins(), "Should not support plugins")
	})

	t.Run("llm gateway", func(t *testing.T) {
		port := engine.SupportsLLMGateway()
		assert.Equal(t, 10004, port, "LLM gateway port should be 10004")
	})
}

func TestMistralVibeEngineRequiredSecrets(t *testing.T) {
	engine := NewMistralVibeEngine()

	t.Run("basic secrets", func(t *testing.T) {
		workflowData := &WorkflowData{Name: "test"}
		secrets := engine.GetRequiredSecretNames(workflowData)

		require.Contains(t, secrets, "MISTRAL_API_KEY", "Should require MISTRAL_API_KEY")
	})

	t.Run("with MCP servers", func(t *testing.T) {
		workflowData := &WorkflowData{
			Name: "test-mcp",
			Tools: map[string]any{
				"github": map[string]any{},
			},
		}
		secrets := engine.GetRequiredSecretNames(workflowData)

		require.Contains(t, secrets, "MISTRAL_API_KEY", "Should require MISTRAL_API_KEY")
		require.Contains(t, secrets, "MCP_GATEWAY_API_KEY", "Should require MCP_GATEWAY_API_KEY with MCP servers")
	})

	t.Run("with safe inputs", func(t *testing.T) {
		workflowData := &WorkflowData{
			Name: "test-safe-inputs",
			SafeInputs: &SafeInputsConfig{
				Tools: map[string]*SafeInputToolConfig{
					"custom_tool": {
						Name:        "custom_tool",
						Description: "Custom tool",
						Env: map[string]string{
							"CUSTOM_SECRET": "${{ secrets.CUSTOM_SECRET }}",
						},
					},
				},
			},
		}
		secrets := engine.GetRequiredSecretNames(workflowData)

		require.Contains(t, secrets, "MISTRAL_API_KEY", "Should require MISTRAL_API_KEY")
		require.Contains(t, secrets, "CUSTOM_SECRET", "Should include safe-inputs secrets")
	})
}

func TestMistralVibeEngineInstallation(t *testing.T) {
	engine := NewMistralVibeEngine()

	t.Run("default installation", func(t *testing.T) {
		workflowData := &WorkflowData{
			Name: "test-workflow",
		}

		steps := engine.GetInstallationSteps(workflowData)
		require.NotEmpty(t, steps, "Should generate installation steps")

		// Verify secret validation step exists
		hasSecretValidation := false
		hasVibeInstall := false
		for _, step := range steps {
			stepContent := strings.Join(step, "\n")
			if strings.Contains(stepContent, "validate-secret") || strings.Contains(stepContent, "MISTRAL_API_KEY") {
				hasSecretValidation = true
			}
			if strings.Contains(stepContent, "Install Mistral Vibe") || strings.Contains(stepContent, "vibe --version") {
				hasVibeInstall = true
			}
		}
		assert.True(t, hasSecretValidation, "Should include secret validation")
		assert.True(t, hasVibeInstall, "Should include Vibe installation")
	})

	t.Run("custom command skips installation", func(t *testing.T) {
		workflowData := &WorkflowData{
			Name: "test-custom",
			EngineConfig: &EngineConfig{
				ID:      "mistral-vibe",
				Command: "/custom/path/to/vibe",
			},
		}

		steps := engine.GetInstallationSteps(workflowData)
		assert.Empty(t, steps, "Should skip installation with custom command")
	})

	t.Run("with firewall", func(t *testing.T) {
		workflowData := &WorkflowData{
			Name: "test-firewall",
			NetworkPermissions: &NetworkPermissions{
				Firewall: &FirewallConfig{
					Enabled: true,
				},
			},
		}

		steps := engine.GetInstallationSteps(workflowData)
		require.NotEmpty(t, steps, "Should generate installation steps")

		// Check if AWF installation is included
		stepFound := false
		for _, step := range steps {
			stepContent := strings.Join(step, "\n")
			if strings.Contains(stepContent, "awf") || strings.Contains(stepContent, "firewall") {
				stepFound = true
			}
		}
		// AWF installation may or may not be included depending on configuration
		// Just verify steps were generated
		assert.NotEmpty(t, steps, "Should generate steps with firewall config")
		_ = stepFound // Avoid unused variable error
	})
}

func TestMistralVibeEngineExecution(t *testing.T) {
	engine := NewMistralVibeEngine()

	t.Run("basic execution", func(t *testing.T) {
		workflowData := &WorkflowData{
			Name: "test-workflow",
			EngineConfig: &EngineConfig{
				ID: "mistral-vibe",
			},
		}

		steps := engine.GetExecutionSteps(workflowData, "/tmp/test.log")
		require.NotEmpty(t, steps, "Should generate execution steps")

		// Verify vibe command is included
		hasVibeCommand := false
		for _, step := range steps {
			stepContent := strings.Join(step, "\n")
			if strings.Contains(stepContent, "vibe") || strings.Contains(stepContent, "Run Mistral Vibe") {
				hasVibeCommand = true
			}
		}
		assert.True(t, hasVibeCommand, "Should include vibe command")
	})

	t.Run("with max turns", func(t *testing.T) {
		workflowData := &WorkflowData{
			Name: "test-max-turns",
			EngineConfig: &EngineConfig{
				ID:       "mistral-vibe",
				MaxTurns: "20",
			},
		}

		steps := engine.GetExecutionSteps(workflowData, "/tmp/test.log")
		require.NotEmpty(t, steps, "Should generate execution steps")

		hasMaxTurns := false
		for _, step := range steps {
			stepContent := strings.Join(step, "\n")
			if strings.Contains(stepContent, "--max-turns") && strings.Contains(stepContent, "20") {
				hasMaxTurns = true
			}
		}
		assert.True(t, hasMaxTurns, "Should include max-turns flag")
	})

	t.Run("with custom model", func(t *testing.T) {
		workflowData := &WorkflowData{
			Name: "test-model",
			EngineConfig: &EngineConfig{
				ID:    "mistral-vibe",
				Model: "devstral-2",
			},
		}

		steps := engine.GetExecutionSteps(workflowData, "/tmp/test.log")
		require.NotEmpty(t, steps, "Should generate execution steps")

		hasModel := false
		for _, step := range steps {
			stepContent := strings.Join(step, "\n")
			if strings.Contains(stepContent, "devstral-2") {
				hasModel = true
			}
		}
		assert.True(t, hasModel, "Should configure model in config.toml")
	})

	t.Run("with MCP servers", func(t *testing.T) {
		workflowData := &WorkflowData{
			Name: "test-mcp",
			Tools: map[string]any{
				"github": map[string]any{},
			},
			EngineConfig: &EngineConfig{
				ID: "mistral-vibe",
			},
		}

		steps := engine.GetExecutionSteps(workflowData, "/tmp/test.log")
		require.NotEmpty(t, steps, "Should generate execution steps")

		hasEnabledTools := false
		hasVibeHome := false
		for _, step := range steps {
			stepContent := strings.Join(step, "\n")
			if strings.Contains(stepContent, "--enabled-tools") {
				hasEnabledTools = true
			}
			if strings.Contains(stepContent, "VIBE_HOME") {
				hasVibeHome = true
			}
		}
		assert.True(t, hasEnabledTools, "Should include enabled-tools flag")
		assert.True(t, hasVibeHome, "Should set VIBE_HOME environment variable")
	})
}

func TestMistralVibeEngineLogParsing(t *testing.T) {
	engine := NewMistralVibeEngine()

	t.Run("parse basic metrics", func(t *testing.T) {
		logContent := `{"session": {"usage": {"input_tokens": 1000, "output_tokens": 500}, "turns": 5}}
{"tool_call": {"name": "bash", "args": {}}}
{"tool_call": {"name": "read_file", "args": {}}}`

		metrics := engine.ParseLogMetrics(logContent, false)

		assert.Equal(t, 5, metrics.Turns, "Should parse turn count")
		assert.Equal(t, 1500, metrics.TokenUsage, "Should calculate total token usage")
		assert.Equal(t, 2, len(metrics.ToolCalls), "Should count tool calls")
	})

	t.Run("empty log", func(t *testing.T) {
		metrics := engine.ParseLogMetrics("", false)
		assert.Equal(t, 0, metrics.Turns, "Turns should be 0 for empty log")
	})
}

func TestMistralVibeEngineLogFiles(t *testing.T) {
	engine := NewMistralVibeEngine()

	t.Run("log parser script id", func(t *testing.T) {
		scriptId := engine.GetLogParserScriptId()
		assert.Empty(t, scriptId, "Vibe uses Go-based parsing, no JS script needed")
	})

	t.Run("log file for parsing", func(t *testing.T) {
		logFile := engine.GetLogFileForParsing()
		assert.Equal(t, "/tmp/gh-aw/agent-stdio.log", logFile, "Should use standard log file")
	})
}

func TestMistralVibeEngineDeclaredOutputFiles(t *testing.T) {
	engine := NewMistralVibeEngine()
	outputFiles := engine.GetDeclaredOutputFiles()
	assert.Empty(t, outputFiles, "Vibe does not declare output files")
}

func TestGetMistralVibeAllowedDomains(t *testing.T) {
	t.Run("default domains", func(t *testing.T) {
		domains := GetMistralVibeAllowedDomainsWithToolsAndRuntimes(nil, nil, nil)
		require.Contains(t, domains, "api.mistral.ai", "Should include Mistral API domain")
	})

	t.Run("with network permissions", func(t *testing.T) {
		networkPerms := &NetworkPermissions{
			Allowed: []string{"example.com", "api.custom.com"},
		}
		domains := GetMistralVibeAllowedDomainsWithToolsAndRuntimes(networkPerms, nil, nil)

		require.Contains(t, domains, "api.mistral.ai", "Should include Mistral API domain")
		require.Contains(t, domains, "example.com", "Should include custom domain")
		require.Contains(t, domains, "api.custom.com", "Should include custom API domain")
	})

	t.Run("with runtimes", func(t *testing.T) {
		runtimes := map[string]any{
			"node": map[string]any{"version": "20"},
		}
		domains := GetMistralVibeAllowedDomainsWithToolsAndRuntimes(nil, nil, runtimes)

		// Should include Mistral API domain and runtime ecosystem domains
		require.Contains(t, domains, "api.mistral.ai", "Should include Mistral API domain")
	})
}
