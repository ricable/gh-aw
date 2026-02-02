package workflow

import (
	"strings"

	"github.com/githubnext/gh-aw/pkg/logger"
)

var copilotMCPLog = logger.New("workflow:copilot_mcp")

// RenderMCPConfig generates MCP server configuration for Copilot CLI
func (e *CopilotEngine) RenderMCPConfig(yaml *strings.Builder, tools map[string]any, mcpTools []string, workflowData *WorkflowData) {
	copilotMCPLog.Printf("Rendering MCP config for Copilot engine: mcpTools=%d", len(mcpTools))

	// Create the directory first
	yaml.WriteString("          mkdir -p /home/runner/.copilot\n")

	// Create unified renderer with Copilot-specific options
	// Copilot uses JSON format with type and tools fields, and inline args
	createRenderer := func(isLast bool) *MCPConfigRendererUnified {
		return NewMCPConfigRenderer(MCPRendererOptions{
			IncludeCopilotFields: true, // Copilot uses "type" and "tools" fields
			InlineArgs:           true, // Copilot uses inline args format
			Format:               "json",
			IsLast:               isLast,
		})
	}

	// Build gateway configuration for MCP config
	// Per MCP Gateway Specification v1.0.0 section 4.1.3, the gateway section is required
	gatewayConfig := buildMCPGatewayConfig(workflowData)

	// Use shared JSON MCP config renderer with unified renderer methods
	options := e.buildCopilotMCPConfigOptions(createRenderer, gatewayConfig, workflowData, false)

	RenderJSONMCPConfig(yaml, tools, mcpTools, workflowData, options)
}

// RenderMCPConfigWithoutGateway generates MCP server configuration for Copilot CLI
// without the MCP gateway proxy. This is used when sandbox is disabled and
// MCP servers run in their configured mode (stdio, Docker, or HTTP) and communicate directly with the agent.
// Note: Container-based MCP servers (playwright, serena, agentic-workflows) are filtered out
// because they require Docker/container runtime which is not available without the sandbox.
func (e *CopilotEngine) RenderMCPConfigWithoutGateway(yaml *strings.Builder, tools map[string]any, mcpTools []string, workflowData *WorkflowData) {
	copilotMCPLog.Printf("Rendering MCP config without gateway for Copilot engine: mcpTools=%d", len(mcpTools))

	// Create the directory first
	yaml.WriteString("          mkdir -p /home/runner/.copilot\n")

	// Create unified renderer with Copilot-specific options
	createRenderer := func(isLast bool) *MCPConfigRendererUnified {
		return NewMCPConfigRenderer(MCPRendererOptions{
			IncludeCopilotFields: true,
			InlineArgs:           true,
			Format:               "json",
			IsLast:               isLast,
		})
	}

	// Build base options without gateway
	options := e.buildCopilotMCPConfigOptions(createRenderer, nil, workflowData, true)

	// Override the FilterTool to also filter out container-based MCP servers
	// These require Docker/container runtime which is not available when sandbox is disabled
	baseFilter := options.FilterTool
	options.FilterTool = func(toolName string) bool {
		// First apply base filter (e.g., cache-memory)
		if baseFilter != nil && !baseFilter(toolName) {
			return false
		}
		// Filter out container-based MCP servers that won't work without Docker
		// playwright, serena, and agentic-workflows all require container runtime
		containerBasedTools := map[string]bool{
			"playwright":        true,
			"serena":            true,
			"agentic-workflows": true,
		}
		if containerBasedTools[toolName] {
			copilotMCPLog.Printf("Filtering out container-based MCP tool '%s' (sandbox disabled, no Docker)", toolName)
			return false
		}
		return true
	}

	RenderJSONMCPConfig(yaml, tools, mcpTools, workflowData, options)
}

// buildCopilotMCPConfigOptions creates the JSONMCPConfigOptions for Copilot engine
// This shared helper avoids code duplication between RenderMCPConfig and RenderMCPConfigWithoutGateway
func (e *CopilotEngine) buildCopilotMCPConfigOptions(
	createRenderer func(isLast bool) *MCPConfigRendererUnified,
	gatewayConfig *MCPGatewayRuntimeConfig,
	workflowData *WorkflowData,
	skipGatewayStartup bool,
) JSONMCPConfigOptions {
	return JSONMCPConfigOptions{
		ConfigPath:         "/home/runner/.copilot/mcp-config.json",
		GatewayConfig:      gatewayConfig,
		SkipGatewayStartup: skipGatewayStartup,
		Renderers: MCPToolRenderers{
			RenderGitHub: func(yaml *strings.Builder, githubTool any, isLast bool, workflowData *WorkflowData) {
				renderer := createRenderer(isLast)
				renderer.RenderGitHubMCP(yaml, githubTool, workflowData)
			},
			RenderPlaywright: func(yaml *strings.Builder, playwrightTool any, isLast bool) {
				renderer := createRenderer(isLast)
				renderer.RenderPlaywrightMCP(yaml, playwrightTool)
			},
			RenderSerena: func(yaml *strings.Builder, serenaTool any, isLast bool) {
				renderer := createRenderer(isLast)
				renderer.RenderSerenaMCP(yaml, serenaTool)
			},
			RenderCacheMemory: func(yaml *strings.Builder, isLast bool, workflowData *WorkflowData) {
				// Cache-memory is not used for Copilot (filtered out)
			},
			RenderAgenticWorkflows: func(yaml *strings.Builder, isLast bool) {
				renderer := createRenderer(isLast)
				renderer.RenderAgenticWorkflowsMCP(yaml)
			},
			RenderSafeOutputs: func(yaml *strings.Builder, isLast bool, workflowData *WorkflowData) {
				renderer := createRenderer(isLast)
				renderer.RenderSafeOutputsMCP(yaml, workflowData)
			},
			RenderSafeInputs: func(yaml *strings.Builder, safeInputs *SafeInputsConfig, isLast bool) {
				renderer := createRenderer(isLast)
				renderer.RenderSafeInputsMCP(yaml, safeInputs, workflowData)
			},
			RenderWebFetch: func(yaml *strings.Builder, isLast bool) {
				renderMCPFetchServerConfig(yaml, "json", "              ", isLast, true)
			},
			RenderCustomMCPConfig: func(yaml *strings.Builder, toolName string, toolConfig map[string]any, isLast bool) error {
				return e.renderCopilotMCPConfigWithContext(yaml, toolName, toolConfig, isLast, workflowData)
			},
		},
		FilterTool: func(toolName string) bool {
			// Filter out cache-memory for Copilot
			// Cache-memory is handled as a simple file share, not an MCP server
			return toolName != "cache-memory"
		},
	}
}

// renderCopilotMCPConfigWithContext generates custom MCP server configuration for Copilot CLI
// This version includes workflowData to determine if localhost URLs should be rewritten
func (e *CopilotEngine) renderCopilotMCPConfigWithContext(yaml *strings.Builder, toolName string, toolConfig map[string]any, isLast bool, workflowData *WorkflowData) error {
	copilotMCPLog.Printf("Rendering custom MCP config for tool: %s", toolName)

	// Determine if localhost URLs should be rewritten to host.docker.internal
	// This is needed when firewall is enabled (agent is not disabled)
	rewriteLocalhost := workflowData != nil && (workflowData.SandboxConfig == nil ||
		workflowData.SandboxConfig.Agent == nil ||
		!workflowData.SandboxConfig.Agent.Disabled)

	// Use the shared renderer with copilot-specific requirements
	renderer := MCPConfigRenderer{
		Format:                   "json",
		IndentLevel:              "                ",
		RequiresCopilotFields:    true,
		RewriteLocalhostToDocker: rewriteLocalhost,
	}

	yaml.WriteString("              \"" + toolName + "\": {\n")

	// Use shared renderer for the server configuration
	if err := renderSharedMCPConfig(yaml, toolName, toolConfig, renderer); err != nil {
		return err
	}

	if isLast {
		yaml.WriteString("              }\n")
	} else {
		yaml.WriteString("              },\n")
	}

	return nil
}
