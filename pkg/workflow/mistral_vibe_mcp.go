package workflow

import (
	"strings"

	"github.com/github/gh-aw/pkg/logger"
)

var mistralVibeMCPLog = logger.New("workflow:mistral_vibe_mcp")

// RenderMCPConfig renders the MCP configuration for Mistral Vibe engine
// Mistral Vibe uses TOML format in config.toml file instead of JSON
func (e *MistralVibeEngine) RenderMCPConfig(yaml *strings.Builder, tools map[string]any, mcpTools []string, workflowData *WorkflowData) {
	mistralVibeMCPLog.Printf("Rendering MCP config for Mistral Vibe: tool_count=%d, mcp_tool_count=%d", len(tools), len(mcpTools))

	// Mistral Vibe uses config.toml format, not JSON
	// We need to generate TOML configuration as part of the config setup
	// This will be handled in the generateConfigTOML method during execution step generation

	// Generate MCP servers configuration step
	yaml.WriteString("      - name: Generate Vibe MCP Configuration\n")
	yaml.WriteString("        id: setup_vibe_mcp\n")
	yaml.WriteString("        run: |\n")
	yaml.WriteString("          mkdir -p /tmp/gh-aw/vibe-config\n")
	yaml.WriteString("          cat >> /tmp/gh-aw/vibe-config/config.toml << 'MCP_CONFIG_EOF'\n")

	// Render MCP servers in TOML format
	e.renderMCPServersTOML(yaml, tools, workflowData)

	yaml.WriteString("          MCP_CONFIG_EOF\n")
}

// renderMCPServersTOML renders MCP servers in TOML format for Vibe
func (e *MistralVibeEngine) renderMCPServersTOML(yaml *strings.Builder, tools map[string]any, workflowData *WorkflowData) {
	mistralVibeMCPLog.Print("Rendering MCP servers in TOML format")

	// Render GitHub MCP server if present
	if hasGitHubTool(workflowData.ParsedTools) {
		e.renderGitHubMCPServerTOML(yaml, tools["github"], workflowData)
	}

	// Render Playwright MCP server if present
	if playwrightTool, hasPlaywright := tools["playwright"]; hasPlaywright {
		e.renderPlaywrightMCPServerTOML(yaml, playwrightTool)
	}

	// Render custom MCP servers
	if customTools, hasCustom := tools["mcp_servers"].(map[string]any); hasCustom {
		for serverName, serverConfig := range customTools {
			e.renderCustomMCPServerTOML(yaml, serverName, serverConfig)
		}
	}
}

// renderGitHubMCPServerTOML renders GitHub MCP server configuration in TOML format
func (e *MistralVibeEngine) renderGitHubMCPServerTOML(yaml *strings.Builder, githubTool any, workflowData *WorkflowData) {
	yaml.WriteString("\n")
	yaml.WriteString("          [[mcp_servers]]\n")
	yaml.WriteString("          name = \"github\"\n")
	yaml.WriteString("          transport = \"stdio\"\n")
	yaml.WriteString("          command = \"npx\"\n")
	yaml.WriteString("          args = [\"-y\", \"@modelcontextprotocol/server-github\"]\n")

	// Add environment variables
	yaml.WriteString("          [mcp_servers.env]\n")
	yaml.WriteString("          GITHUB_TOKEN = \"${GITHUB_MCP_SERVER_TOKEN}\"\n")
}

// renderPlaywrightMCPServerTOML renders Playwright MCP server configuration in TOML format
func (e *MistralVibeEngine) renderPlaywrightMCPServerTOML(yaml *strings.Builder, playwrightTool any) {
	yaml.WriteString("\n")
	yaml.WriteString("          [[mcp_servers]]\n")
	yaml.WriteString("          name = \"playwright\"\n")
	yaml.WriteString("          transport = \"stdio\"\n")
	yaml.WriteString("          command = \"npx\"\n")
	yaml.WriteString("          args = [\"-y\", \"@playwright/mcp\"]\n")
}

// renderCustomMCPServerTOML renders custom MCP server configuration in TOML format
func (e *MistralVibeEngine) renderCustomMCPServerTOML(yaml *strings.Builder, serverName string, serverConfig any) {
	configMap, ok := serverConfig.(map[string]any)
	if !ok {
		mistralVibeMCPLog.Printf("Skipping invalid custom MCP server config: %s", serverName)
		return
	}

	yaml.WriteString("\n")
	yaml.WriteString("          [[mcp_servers]]\n")
	yaml.WriteString("          name = \"")
	yaml.WriteString(serverName)
	yaml.WriteString("\"\n")

	// Render transport
	if transport, hasTransport := configMap["transport"].(string); hasTransport {
		yaml.WriteString("          transport = \"")
		yaml.WriteString(transport)
		yaml.WriteString("\"\n")
	}

	// Render command
	if command, hasCommand := configMap["command"].(string); hasCommand {
		yaml.WriteString("          command = \"")
		yaml.WriteString(command)
		yaml.WriteString("\"\n")
	}

	// Render args
	if args, hasArgs := configMap["args"].([]any); hasArgs {
		yaml.WriteString("          args = [")
		for i, arg := range args {
			if i > 0 {
				yaml.WriteString(", ")
			}
			yaml.WriteString("\"")
			yaml.WriteString(arg.(string))
			yaml.WriteString("\"")
		}
		yaml.WriteString("]\n")
	}

	// Render env if present
	if env, hasEnv := configMap["env"].(map[string]any); hasEnv && len(env) > 0 {
		yaml.WriteString("          [mcp_servers.env]\n")
		for key, value := range env {
			yaml.WriteString("          ")
			yaml.WriteString(key)
			yaml.WriteString(" = \"")
			yaml.WriteString(value.(string))
			yaml.WriteString("\"\n")
		}
	}
}
