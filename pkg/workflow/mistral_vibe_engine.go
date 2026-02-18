package workflow

import (
	"fmt"
	"strings"

	"github.com/github/gh-aw/pkg/constants"
	"github.com/github/gh-aw/pkg/logger"
)

var mistralVibeLog = logger.New("workflow:mistral_vibe_engine")

// MistralVibeEngine represents the Mistral Vibe agentic engine
type MistralVibeEngine struct {
	BaseEngine
}

func NewMistralVibeEngine() *MistralVibeEngine {
	return &MistralVibeEngine{
		BaseEngine: BaseEngine{
			id:                     "mistral-vibe",
			displayName:            "Mistral Vibe CLI",
			description:            "Uses Mistral Vibe CLI with MCP server support and tool allow-listing",
			experimental:           true,
			supportsToolsAllowlist: true,
			supportsMaxTurns:       true,  // Vibe supports max-turns feature
			supportsWebFetch:       false, // Vibe does not have built-in web-fetch support
			supportsWebSearch:      false, // Vibe does not have built-in web-search support
			supportsFirewall:       true,  // Vibe supports network firewalling via AWF
			supportsPlugins:        false, // Vibe does not support plugin installation
			supportsLLMGateway:     false, // Vibe supports LLM gateway
		},
	}
}

// SupportsLLMGateway returns the LLM gateway port for Mistral Vibe engine
func (e *MistralVibeEngine) SupportsLLMGateway() int {
	return constants.MistralVibeLLMGatewayPort
}

// GetRequiredSecretNames returns the list of secrets required by the Mistral Vibe engine
// This includes MISTRAL_API_KEY and optionally MCP_GATEWAY_API_KEY
func (e *MistralVibeEngine) GetRequiredSecretNames(workflowData *WorkflowData) []string {
	secrets := []string{"MISTRAL_API_KEY"}

	// Add MCP gateway API key if MCP servers are present (gateway is always started with MCP servers)
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

func (e *MistralVibeEngine) GetInstallationSteps(workflowData *WorkflowData) []GitHubActionStep {
	mistralVibeLog.Printf("Generating installation steps for Mistral Vibe engine: workflow=%s", workflowData.Name)

	// Skip installation if custom command is specified
	if workflowData.EngineConfig != nil && workflowData.EngineConfig.Command != "" {
		mistralVibeLog.Printf("Skipping installation steps: custom command specified (%s)", workflowData.EngineConfig.Command)
		return []GitHubActionStep{}
	}

	var steps []GitHubActionStep

	// Add secret validation step
	secretValidation := GenerateMultiSecretValidationStep(
		[]string{"MISTRAL_API_KEY"},
		"Mistral Vibe",
		"https://github.github.com/gh-aw/reference/engines/#mistral-vibe",
	)
	steps = append(steps, secretValidation)

	// Determine Mistral Vibe version
	vibeVersion := string(constants.DefaultMistralVibeVersion)
	if workflowData.EngineConfig != nil && workflowData.EngineConfig.Version != "" {
		vibeVersion = workflowData.EngineConfig.Version
	}

	// Add Mistral Vibe installation step using curl/uv
	installStep := GitHubActionStep{
		"      - name: Install Mistral Vibe CLI",
		"        id: install_mistral_vibe",
		"        run: |",
		"          # Install Mistral Vibe using official install script",
		fmt.Sprintf("          curl -fsSL https://raw.githubusercontent.com/mistralai/vibe/v%s/install.sh | bash", vibeVersion),
		"          # Add vibe to PATH",
		"          echo \"$HOME/.local/bin\" >> $GITHUB_PATH",
		"          # Verify installation",
		"          vibe --version",
	}
	steps = append(steps, installStep)

	// Add AWF installation if firewall is enabled
	if isFirewallEnabled(workflowData) {
		firewallConfig := getFirewallConfig(workflowData)
		agentConfig := getAgentConfig(workflowData)
		var awfVersion string
		if firewallConfig != nil {
			awfVersion = firewallConfig.Version
		}

		// Install AWF binary (or skip if custom command is specified)
		awfInstall := generateAWFInstallationStep(awfVersion, agentConfig)
		if len(awfInstall) > 0 {
			steps = append(steps, awfInstall)
		}
	}

	return steps
}

// GetDeclaredOutputFiles returns the output files that Mistral Vibe may produce
func (e *MistralVibeEngine) GetDeclaredOutputFiles() []string {
	return []string{}
}

// GetExecutionSteps returns the GitHub Actions steps for executing Mistral Vibe
func (e *MistralVibeEngine) GetExecutionSteps(workflowData *WorkflowData, logFile string) []GitHubActionStep {
	mistralVibeLog.Printf("Generating execution steps for Mistral Vibe engine: workflow=%s, firewall=%v", workflowData.Name, isFirewallEnabled(workflowData))

	// Handle custom steps if they exist in engine config
	steps := InjectCustomEngineSteps(workflowData, e.convertStepToYAML)

	// Build vibe CLI arguments based on configuration
	var vibeArgs []string

	// Add programmatic mode flag for non-interactive execution (auto-approves)
	vibeArgs = append(vibeArgs, "-p")

	// Add model if specified via engine config
	modelConfigured := workflowData.EngineConfig != nil && workflowData.EngineConfig.Model != ""

	// Add max_turns if specified (in CLI it's max-turns)
	if workflowData.EngineConfig != nil && workflowData.EngineConfig.MaxTurns != "" {
		mistralVibeLog.Printf("Setting max turns: %s", workflowData.EngineConfig.MaxTurns)
		vibeArgs = append(vibeArgs, "--max-turns", workflowData.EngineConfig.MaxTurns)
	}

	// Add output format for structured output
	vibeArgs = append(vibeArgs, "--output", "streaming")

	// Add enabled tools configuration if MCP servers are present
	if HasMCPServers(workflowData) {
		enabledTools := e.computeEnabledVibeToolsString(workflowData.Tools, workflowData.SafeOutputs, workflowData.CacheMemoryConfig)
		if enabledTools != "" {
			vibeArgs = append(vibeArgs, "--enabled-tools", enabledTools)
		}
	}

	// Add custom args from engine configuration before the prompt
	if workflowData.EngineConfig != nil && len(workflowData.EngineConfig.Args) > 0 {
		vibeArgs = append(vibeArgs, workflowData.EngineConfig.Args...)
	}

	// Build the agent command - prepend custom agent file content if specified (via imports)
	var promptSetup string
	var promptCommand string
	if workflowData.AgentFile != "" {
		agentPath := ResolveAgentFilePath(workflowData.AgentFile)
		mistralVibeLog.Printf("Using custom agent file: %s", workflowData.AgentFile)
		// Extract markdown body from custom agent file and prepend to prompt
		promptSetup = fmt.Sprintf(`# Extract markdown body from custom agent file (skip frontmatter)
          AGENT_CONTENT="$(awk 'BEGIN{skip=1} /^---$/{if(skip){skip=0;next}else{skip=1;next}} !skip' %s)"
          # Combine agent content with prompt
          PROMPT_TEXT="$(printf '%%s\n\n%%s' "$AGENT_CONTENT" "$(cat /tmp/gh-aw/aw-prompts/prompt.txt)")"`, agentPath)
		promptCommand = "\"$PROMPT_TEXT\""
	} else {
		promptCommand = "\"$(cat /tmp/gh-aw/aw-prompts/prompt.txt)\""
	}

	// Build the command string with proper argument formatting
	// Determine which command to use
	var commandName string
	if workflowData.EngineConfig != nil && workflowData.EngineConfig.Command != "" {
		commandName = workflowData.EngineConfig.Command
		mistralVibeLog.Printf("Using custom command: %s", commandName)
	} else {
		commandName = "vibe"
	}

	commandParts := []string{commandName}
	commandParts = append(commandParts, vibeArgs...)
	commandParts = append(commandParts, promptCommand)

	// Join command parts with proper escaping using shellJoinArgs helper
	vibeCommand := shellJoinArgs(commandParts)

	// Add conditional model flag if not explicitly configured
	// Check if this is a detection job (has no SafeOutputs config)
	isDetectionJob := workflowData.SafeOutputs == nil
	var modelEnvVar string
	if isDetectionJob {
		modelEnvVar = "GH_AW_MODEL_DETECTION_MISTRAL_VIBE"
	} else {
		modelEnvVar = "GH_AW_MODEL_AGENT_MISTRAL_VIBE"
	}

	// Build the full command based on whether firewall is enabled
	var command string
	if isFirewallEnabled(workflowData) {
		// Build the AWF-wrapped command using helper function
		// Get allowed domains (Mistral defaults + network permissions + HTTP MCP server URLs + runtime ecosystem domains)
		allowedDomains := GetMistralVibeAllowedDomainsWithToolsAndRuntimes(workflowData.NetworkPermissions, workflowData.Tools, workflowData.Runtimes)

		// Enable API proxy sidecar if this engine supports LLM gateway
		llmGatewayPort := e.SupportsLLMGateway()
		usesAPIProxy := llmGatewayPort > 0

		// Prepend PATH setup for vibe CLI
		pathSetup := "export PATH=\"$HOME/.local/bin:$PATH\""
		vibeCommandWithPath := fmt.Sprintf("%s && %s", pathSetup, vibeCommand)

		// Add config.toml setup before vibe command
		configSetup := e.generateConfigTOML(workflowData, modelConfigured)
		fullCommand := fmt.Sprintf("%s\n%s", configSetup, vibeCommandWithPath)

		// Add prompt setup if using custom agent file
		if promptSetup != "" {
			fullCommand = fmt.Sprintf("%s\n%s", promptSetup, fullCommand)
		}

		command = BuildAWFCommand(AWFCommandConfig{
			EngineName:     "mistral-vibe",
			EngineCommand:  fullCommand,
			LogFile:        logFile,
			WorkflowData:   workflowData,
			UsesTTY:        false,
			UsesAPIProxy:   usesAPIProxy,
			AllowedDomains: allowedDomains,
		})
	} else {
		// Non-firewall mode - simpler command without AWF wrapping
		configSetup := e.generateConfigTOML(workflowData, modelConfigured)

		if promptSetup != "" {
			command = fmt.Sprintf(`set -o pipefail
%s
%s
%s 2>&1 | tee %s`, promptSetup, configSetup, vibeCommand, logFile)
		} else {
			command = fmt.Sprintf(`set -o pipefail
%s
%s 2>&1 | tee %s`, configSetup, vibeCommand, logFile)
		}
	}

	// Build environment variables
	env := map[string]string{
		"MISTRAL_API_KEY":  "${{ secrets.MISTRAL_API_KEY }}",
		"GH_AW_PROMPT":     "/tmp/gh-aw/aw-prompts/prompt.txt",
		"GITHUB_WORKSPACE": "${{ github.workspace }}",
	}

	// Add VIBE_HOME for config location
	if HasMCPServers(workflowData) {
		env["VIBE_HOME"] = "/tmp/gh-aw/vibe-config"
	}

	// Add safe outputs env
	applySafeOutputEnvToMap(env, workflowData)

	// Add model env var if not explicitly configured
	if !modelConfigured {
		env[modelEnvVar] = fmt.Sprintf("${{ vars.%s || '' }}", modelEnvVar)
	}

	// Generate the execution step
	stepLines := []string{
		"      - name: Run Mistral Vibe",
		"        id: agentic_execution",
	}

	// Filter environment variables for security
	allowedSecrets := e.GetRequiredSecretNames(workflowData)
	filteredEnv := FilterEnvForSecrets(env, allowedSecrets)

	// Format step with command and env
	stepLines = FormatStepWithCommandAndEnv(stepLines, command, filteredEnv)

	steps = append(steps, GitHubActionStep(stepLines))
	return steps
}

// generateConfigTOML generates the Vibe config.toml file with MCP servers and model configuration
func (e *MistralVibeEngine) generateConfigTOML(workflowData *WorkflowData, modelConfigured bool) string {
	var configLines []string

	configLines = append(configLines, "# Generate Vibe config.toml")
	configLines = append(configLines, "mkdir -p \"${VIBE_HOME:-$HOME/.vibe}\"")
	configLines = append(configLines, "cat > \"${VIBE_HOME:-$HOME/.vibe}/config.toml\" << 'VIBE_CONFIG_EOF'")

	// Add model configuration if specified
	if modelConfigured && workflowData.EngineConfig != nil {
		configLines = append(configLines, fmt.Sprintf("active_model = \"%s\"", workflowData.EngineConfig.Model))
		configLines = append(configLines, "")
	}

	// Add MCP servers configuration if present
	if HasMCPServers(workflowData) {
		// MCP config will be added via RenderMCPConfig
		configLines = append(configLines, "# MCP servers will be configured via RenderMCPConfig")
	}

	configLines = append(configLines, "VIBE_CONFIG_EOF")

	return strings.Join(configLines, "\n")
}

// computeEnabledVibeToolsString computes the --enabled-tools string for Vibe CLI
// Vibe supports glob patterns and regex patterns (re:^pattern$)
func (e *MistralVibeEngine) computeEnabledVibeToolsString(tools map[string]any, safeOutputs *SafeOutputsConfig, cacheMemory *CacheMemoryConfig) string {
	var enabledTools []string

	// Always enable basic file operations
	enabledTools = append(enabledTools, "bash", "read_file", "write_file")

	// Add MCP tools with glob patterns
	if _, hasGitHub := tools["github"]; hasGitHub {
		enabledTools = append(enabledTools, "github_*")
	}

	// Add playwright if present
	if _, hasPlaywright := tools["playwright"]; hasPlaywright {
		enabledTools = append(enabledTools, "playwright_*")
	}

	return strings.Join(enabledTools, " ")
}

// GetMistralVibeAllowedDomainsWithToolsAndRuntimes returns allowed domains for Mistral Vibe
// Returns a deduplicated, sorted, comma-separated string suitable for AWF's --allow-domains flag
func GetMistralVibeAllowedDomainsWithToolsAndRuntimes(networkPerms *NetworkPermissions, tools map[string]any, runtimes map[string]any) string {
	// Start with Mistral API domain
	mistralDefaults := []string{
		"api.mistral.ai",
	}

	return mergeDomainsWithNetworkToolsAndRuntimes(mistralDefaults, networkPerms, tools, runtimes)
}
