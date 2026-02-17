package workflow

import (
	"fmt"
	"strings"
	"time"

	"github.com/github/gh-aw/pkg/constants"
	"github.com/github/gh-aw/pkg/logger"
)

var openclawLog = logger.New("workflow:openclaw_engine")

// OpenClawEngine represents the OpenClaw agentic engine
type OpenClawEngine struct {
	BaseEngine
}

func NewOpenClawEngine() *OpenClawEngine {
	return &OpenClawEngine{
		BaseEngine: BaseEngine{
			id:                     "openclaw",
			displayName:            "OpenClaw",
			description:            "Uses OpenClaw agent platform with ACP tool support",
			experimental:           true,
			supportsToolsAllowlist: false, // OpenClaw uses its own skill/tool system
			supportsMaxTurns:       false, // Timeout-based, not turn-based
			supportsWebFetch:       false, // Not built-in
			supportsWebSearch:      false, // Not built-in
			supportsFirewall:       true,  // Supports AWF sandboxing
			supportsLLMGateway:     false, // Manages its own API connections
		},
	}
}

// SupportsLLMGateway returns -1 as OpenClaw does not support LLM gateway
func (e *OpenClawEngine) SupportsLLMGateway() int {
	return -1
}

// GetRequiredSecretNames returns the list of secrets required by the OpenClaw engine
func (e *OpenClawEngine) GetRequiredSecretNames(workflowData *WorkflowData) []string {
	secrets := []string{"OPENCLAW_API_KEY", "ANTHROPIC_API_KEY"}

	// Add MCP gateway API key if MCP servers are present
	if HasMCPServers(workflowData) {
		secrets = append(secrets, "MCP_GATEWAY_API_KEY")
	}

	// Add safe-inputs secret names
	if IsSafeInputsEnabled(workflowData.SafeInputs, workflowData) {
		safeInputsSecrets := collectSafeInputsSecrets(workflowData.SafeInputs)
		for varName := range safeInputsSecrets {
			secrets = append(secrets, varName)
		}
	}

	return secrets
}

func (e *OpenClawEngine) GetInstallationSteps(workflowData *WorkflowData) []GitHubActionStep {
	openclawLog.Printf("Generating installation steps for OpenClaw engine: workflow=%s", workflowData.Name)

	// Skip installation if custom command is specified
	if workflowData.EngineConfig != nil && workflowData.EngineConfig.Command != "" {
		openclawLog.Printf("Skipping installation steps: custom command specified (%s)", workflowData.EngineConfig.Command)
		return []GitHubActionStep{}
	}

	// Use base installation steps (secret validation + npm install)
	steps := GetBaseInstallationSteps(EngineInstallConfig{
		Secrets:    []string{"OPENCLAW_API_KEY", "ANTHROPIC_API_KEY"},
		DocsURL:    "https://github.github.com/gh-aw/reference/engines/#openclaw",
		NpmPackage: "openclaw",
		Version:    string(constants.DefaultOpenClawVersion),
		Name:       "OpenClaw",
		CliName:    "openclaw",
	}, workflowData)

	// Add AWF installation step if firewall is enabled
	if isFirewallEnabled(workflowData) {
		firewallConfig := getFirewallConfig(workflowData)
		agentConfig := getAgentConfig(workflowData)
		var awfVersion string
		if firewallConfig != nil {
			awfVersion = firewallConfig.Version
		}

		awfInstall := generateAWFInstallationStep(awfVersion, agentConfig)
		if len(awfInstall) > 0 {
			steps = append(steps, awfInstall)
		}
	}

	return steps
}

// GetDeclaredOutputFiles returns the output files that OpenClaw may produce
func (e *OpenClawEngine) GetDeclaredOutputFiles() []string {
	return []string{}
}

// GetExecutionSteps returns the GitHub Actions steps for executing OpenClaw
func (e *OpenClawEngine) GetExecutionSteps(workflowData *WorkflowData, logFile string) []GitHubActionStep {
	modelConfigured := workflowData.EngineConfig != nil && workflowData.EngineConfig.Model != ""
	firewallEnabled := isFirewallEnabled(workflowData)
	openclawLog.Printf("Building OpenClaw execution steps: workflow=%s, model_configured=%v, has_agent_file=%v, firewall=%v",
		workflowData.Name, modelConfigured, workflowData.AgentFile != "", firewallEnabled)

	// Handle custom steps if they exist in engine config
	steps := InjectCustomEngineSteps(workflowData, e.convertStepToYAML)

	// Build OpenClaw CLI arguments
	var openclawArgs []string
	openclawArgs = append(openclawArgs, "agent")
	openclawArgs = append(openclawArgs, "--local")
	openclawArgs = append(openclawArgs, "--json")
	openclawArgs = append(openclawArgs, "--no-color")
	openclawArgs = append(openclawArgs, "--timeout", "1200")

	// Add model/agent if specified
	if modelConfigured {
		openclawLog.Printf("Using custom agent: %s", workflowData.EngineConfig.Model)
		openclawArgs = append(openclawArgs, "--agent", workflowData.EngineConfig.Model)
	}

	// Add MCP configuration if MCP servers are present
	if HasMCPServers(workflowData) {
		openclawLog.Print("Adding MCP configuration")
		openclawArgs = append(openclawArgs, "--mcp-config", "/tmp/gh-aw/mcp-config/mcp-servers.json")
	}

	// Add custom args from engine configuration
	if workflowData.EngineConfig != nil && len(workflowData.EngineConfig.Args) > 0 {
		openclawArgs = append(openclawArgs, workflowData.EngineConfig.Args...)
	}

	// Add the message argument last
	openclawArgs = append(openclawArgs, "--message")

	// Build the agent command - prepend custom agent file content if specified
	var promptSetup string
	var promptCommand string
	if workflowData.AgentFile != "" {
		agentPath := ResolveAgentFilePath(workflowData.AgentFile)
		openclawLog.Printf("Using custom agent file: %s", workflowData.AgentFile)
		promptSetup = fmt.Sprintf(`# Extract markdown body from custom agent file (skip frontmatter)
          AGENT_CONTENT="$(awk 'BEGIN{skip=1} /^---$/{if(skip){skip=0;next}else{skip=1;next}} !skip' %s)"
          # Combine agent content with prompt
          INSTRUCTION="$(printf '%%s\n\n%%s' "$AGENT_CONTENT" "$(cat /tmp/gh-aw/aw-prompts/prompt.txt)")"`, agentPath)
		promptCommand = "\"$INSTRUCTION\""
	} else {
		promptCommand = "\"$(cat /tmp/gh-aw/aw-prompts/prompt.txt)\""
	}

	// Determine which command to use
	var commandName string
	if workflowData.EngineConfig != nil && workflowData.EngineConfig.Command != "" {
		commandName = workflowData.EngineConfig.Command
		openclawLog.Printf("Using custom command: %s", commandName)
	} else {
		commandName = "openclaw"
	}

	commandParts := []string{commandName}
	commandParts = append(commandParts, openclawArgs...)
	commandParts = append(commandParts, promptCommand)

	openclawCommand := shellJoinArgs(commandParts)

	// Add conditional model flag if not explicitly configured
	isDetectionJob := workflowData.SafeOutputs == nil
	var modelEnvVar string
	if isDetectionJob {
		modelEnvVar = constants.EnvVarModelDetectionOpenClaw
	} else {
		modelEnvVar = constants.EnvVarModelAgentOpenClaw
	}
	if !modelConfigured {
		openclawCommand = fmt.Sprintf(`%s${%s:+ --agent "$%s"}`, openclawCommand, modelEnvVar, modelEnvVar)
	}

	// Build the full command based on whether firewall is enabled
	var command string
	if firewallEnabled {
		// Get allowed domains
		allowedDomains := GetOpenClawAllowedDomainsWithToolsAndRuntimes(workflowData.NetworkPermissions, workflowData.Tools, workflowData.Runtimes)

		// OpenClaw does not use LLM gateway
		usesAPIProxy := false

		// Build the command with npm PATH setup for AWF container
		npmPathSetup := GetNpmBinPathSetup()

		var openclawCommandWithSetup string
		if workflowData.AgentFile != "" {
			agentPath := ResolveAgentFilePath(workflowData.AgentFile)
			openclawCommandWithSetup = fmt.Sprintf(`%s && AGENT_CONTENT="$(awk 'BEGIN{skip=1} /^---$/{if(skip){skip=0;next}else{skip=1;next}} !skip' %s)" && INSTRUCTION="$(printf "%%s\n\n%%s" "$AGENT_CONTENT" "$(cat /tmp/gh-aw/aw-prompts/prompt.txt)")" && %s`,
				npmPathSetup, agentPath, openclawCommand)
		} else {
			openclawCommandWithSetup = fmt.Sprintf(`%s && INSTRUCTION="$(cat /tmp/gh-aw/aw-prompts/prompt.txt)" && %s`,
				npmPathSetup, openclawCommand)
		}

		command = BuildAWFCommand(AWFCommandConfig{
			EngineName:     "openclaw",
			EngineCommand:  openclawCommandWithSetup,
			LogFile:        logFile,
			WorkflowData:   workflowData,
			UsesTTY:        false, // OpenClaw uses --json output, not a TUI
			UsesAPIProxy:   usesAPIProxy,
			AllowedDomains: allowedDomains,
			PathSetup:      "mkdir -p \"$OPENCLAW_STATE_DIR\"",
		})
	} else {
		// Run OpenClaw command without AWF wrapper
		if promptSetup != "" {
			command = fmt.Sprintf(`set -o pipefail
          %s
          mkdir -p "$OPENCLAW_STATE_DIR"
          %s 2>&1 | tee %s`, promptSetup, openclawCommand, logFile)
		} else {
			command = fmt.Sprintf(`set -o pipefail
          INSTRUCTION="$(cat "$GH_AW_PROMPT")"
          mkdir -p "$OPENCLAW_STATE_DIR"
          %s 2>&1 | tee %s`, openclawCommand, logFile)
		}
	}

	// Build environment variables map
	env := map[string]string{
		"OPENCLAW_API_KEY":   "${{ secrets.OPENCLAW_API_KEY }}",
		"ANTHROPIC_API_KEY":  "${{ secrets.ANTHROPIC_API_KEY }}",
		"OPENCLAW_STATE_DIR": "/tmp/gh-aw/openclaw-state",
		"GH_AW_PROMPT":       "/tmp/gh-aw/aw-prompts/prompt.txt",
		"DISABLE_TELEMETRY":  "1",
		"GITHUB_WORKSPACE":   "${{ github.workspace }}",
	}

	// Add GH_AW_MCP_CONFIG for MCP server configuration only if there are MCP servers
	if HasMCPServers(workflowData) {
		env["GH_AW_MCP_CONFIG"] = "/tmp/gh-aw/mcp-config/mcp-servers.json"
	}

	// Add GH_AW_SAFE_OUTPUTS if output is needed
	applySafeOutputEnvToMap(env, workflowData)

	// Add GH_AW_STARTUP_TIMEOUT environment variable (in seconds) if startup-timeout is specified
	if workflowData.ToolsStartupTimeout > 0 {
		env["GH_AW_STARTUP_TIMEOUT"] = fmt.Sprintf("%d", workflowData.ToolsStartupTimeout)
	}

	// Add GH_AW_TOOL_TIMEOUT environment variable (in seconds) if timeout is specified
	if workflowData.ToolsTimeout > 0 {
		env["GH_AW_TOOL_TIMEOUT"] = fmt.Sprintf("%d", workflowData.ToolsTimeout)
	}

	// Add model environment variable if model is not explicitly configured
	if !modelConfigured {
		if isDetectionJob {
			env[constants.EnvVarModelDetectionOpenClaw] = fmt.Sprintf("${{ vars.%s || '' }}", constants.EnvVarModelDetectionOpenClaw)
		} else {
			env[constants.EnvVarModelAgentOpenClaw] = fmt.Sprintf("${{ vars.%s || '' }}", constants.EnvVarModelAgentOpenClaw)
		}
	}

	// Add custom environment variables from engine config
	if workflowData.EngineConfig != nil && len(workflowData.EngineConfig.Env) > 0 {
		for key, value := range workflowData.EngineConfig.Env {
			env[key] = value
		}
	}

	// Add custom environment variables from agent config
	agentConfig := getAgentConfig(workflowData)
	if agentConfig != nil && len(agentConfig.Env) > 0 {
		for key, value := range agentConfig.Env {
			env[key] = value
		}
		openclawLog.Printf("Added %d custom env vars from agent config", len(agentConfig.Env))
	}

	// Add safe-inputs secrets to env for passthrough to MCP servers
	if IsSafeInputsEnabled(workflowData.SafeInputs, workflowData) {
		safeInputsSecrets := collectSafeInputsSecrets(workflowData.SafeInputs)
		for varName, secretExpr := range safeInputsSecrets {
			if _, exists := env[varName]; !exists {
				env[varName] = secretExpr
			}
		}
	}

	// Generate the step for OpenClaw execution
	stepName := "Run OpenClaw"
	var stepLines []string

	stepLines = append(stepLines, fmt.Sprintf("      - name: %s", stepName))

	// Add timeout at step level
	if workflowData.TimeoutMinutes != "" {
		timeoutValue := strings.TrimPrefix(workflowData.TimeoutMinutes, "timeout-minutes: ")
		stepLines = append(stepLines, fmt.Sprintf("        timeout-minutes: %s", timeoutValue))
	} else {
		stepLines = append(stepLines, fmt.Sprintf("        timeout-minutes: %d", int(constants.DefaultAgenticWorkflowTimeout/time.Minute)))
	}

	// Filter environment variables to only include allowed secrets
	allowedSecrets := e.GetRequiredSecretNames(workflowData)
	filteredEnv := FilterEnvForSecrets(env, allowedSecrets)

	// Format step with command and filtered environment variables
	stepLines = FormatStepWithCommandAndEnv(stepLines, command, filteredEnv)

	steps = append(steps, GitHubActionStep(stepLines))

	return steps
}

// GetLogParserScriptId returns the JavaScript script name for parsing OpenClaw logs
func (e *OpenClawEngine) GetLogParserScriptId() string {
	return "parse_openclaw_log"
}

// GetFirewallLogsCollectionStep returns the step for collecting firewall logs
func (e *OpenClawEngine) GetFirewallLogsCollectionStep(workflowData *WorkflowData) []GitHubActionStep {
	return []GitHubActionStep{}
}

// GetSquidLogsSteps returns the steps for uploading and parsing Squid logs
func (e *OpenClawEngine) GetSquidLogsSteps(workflowData *WorkflowData) []GitHubActionStep {
	var steps []GitHubActionStep

	if isFirewallEnabled(workflowData) {
		openclawLog.Printf("Adding Squid logs upload and parsing steps for workflow: %s", workflowData.Name)

		squidLogsUpload := generateSquidLogsUploadStep(workflowData.Name)
		steps = append(steps, squidLogsUpload)

		firewallLogParsing := generateFirewallLogParsingStep(workflowData.Name)
		steps = append(steps, firewallLogParsing)
	} else {
		openclawLog.Print("Firewall disabled, skipping Squid logs upload")
	}

	return steps
}

// ParseLogMetrics implements engine-specific log parsing for OpenClaw
func (e *OpenClawEngine) ParseLogMetrics(logContent string, verbose bool) LogMetrics {
	openclawLog.Printf("Parsing OpenClaw log metrics: log_size=%d bytes", len(logContent))

	var metrics LogMetrics

	lines := strings.Split(logContent, "\n")
	toolCallMap := make(map[string]*ToolCallInfo)
	var currentSequence []string

	for _, line := range lines {
		// Count tool calls from JSON output
		if strings.Contains(line, "\"type\":\"tool_call\"") || strings.Contains(line, "\"type\": \"tool_call\"") {
			// Try to extract tool name
			toolName := "unknown"
			if idx := strings.Index(line, "\"name\":\""); idx != -1 {
				nameStart := idx + len("\"name\":\"")
				nameEnd := strings.Index(line[nameStart:], "\"")
				if nameEnd > 0 {
					toolName = line[nameStart : nameStart+nameEnd]
				}
			} else if idx := strings.Index(line, "\"name\": \""); idx != -1 {
				nameStart := idx + len("\"name\": \"")
				nameEnd := strings.Index(line[nameStart:], "\"")
				if nameEnd > 0 {
					toolName = line[nameStart : nameStart+nameEnd]
				}
			}

			if existing, ok := toolCallMap[toolName]; ok {
				existing.CallCount++
			} else {
				toolCallMap[toolName] = &ToolCallInfo{
					Name:      toolName,
					CallCount: 1,
				}
			}
			currentSequence = append(currentSequence, toolName)
		}
	}

	// Convert tool call map to slice
	for _, info := range toolCallMap {
		metrics.ToolCalls = append(metrics.ToolCalls, *info)
	}

	if len(currentSequence) > 0 {
		metrics.ToolSequences = append(metrics.ToolSequences, currentSequence)
	}

	return metrics
}

// GetDefaultDetectionModel returns the default model for OpenClaw detection jobs
func (e *OpenClawEngine) GetDefaultDetectionModel() string {
	return ""
}
