// This file provides sandbox configuration for agentic workflows.
//
// This file handles:
//   - Sandbox type definitions (AWF)
//   - Sandbox configuration structures and parsing
//   - Sandbox runtime config generation
//
// # Validation Functions
//
// Domain-specific validation functions for sandbox configuration are located in
// sandbox_validation.go following the validation architecture pattern.
// See validation.go for the validation architecture documentation.

package workflow

import (
	"github.com/github/gh-aw/pkg/logger"
)

var sandboxLog = logger.New("workflow:sandbox")

// SandboxType represents the type of sandbox to use
type SandboxType string

const (
	SandboxTypeAWF     SandboxType = "awf"     // Uses AWF (Agent Workflow Firewall)
	SandboxTypeDefault SandboxType = "default" // Alias for AWF (backward compat)
)

// SandboxConfig represents the top-level sandbox configuration from front matter
// New format: { agent: "awf"|{id, config}, mcp: {port, command, ...} }
// Legacy format: "default"|"awf" or { type, config }
type SandboxConfig struct {
	// New fields
	Agent *AgentSandboxConfig      `yaml:"agent,omitempty"` // Agent sandbox configuration
	MCP   *MCPGatewayRuntimeConfig `yaml:"mcp,omitempty"`   // MCP gateway configuration

	// Legacy fields (for backward compatibility)
	Type   SandboxType           `yaml:"type,omitempty"`   // Sandbox type: "default" or "awf"
	Config *SandboxRuntimeConfig `yaml:"config,omitempty"` // Custom SRT config (optional)
}

// AgentSandboxConfig represents the agent sandbox configuration
type AgentSandboxConfig struct {
	ID       string                `yaml:"id,omitempty"`      // Agent ID: "awf" (AWF is the only supported sandbox)
	Type     SandboxType           `yaml:"type,omitempty"`    // Sandbox type: "awf" (legacy field, use ID instead)
	Disabled bool                  `yaml:"-"`                 // True when agent is explicitly set to false (disables firewall). This is a runtime flag, not serialized to YAML.
	Config   *SandboxRuntimeConfig `yaml:"config,omitempty"`  // Deprecated: Custom sandbox config (no longer used)
	Command  string                `yaml:"command,omitempty"` // Custom command to replace AWF installation
	Args     []string              `yaml:"args,omitempty"`    // Additional arguments to append to the command
	Env      map[string]string     `yaml:"env,omitempty"`     // Environment variables to set on the step
	Mounts   []string              `yaml:"mounts,omitempty"`  // Container mounts to add for AWF (format: "source:dest:mode")
}

// SandboxRuntimeConfig represents the Anthropic Sandbox Runtime configuration
// This matches the TypeScript SandboxRuntimeConfig interface
// Note: Network configuration is controlled by the top-level 'network' field, not this struct
type SandboxRuntimeConfig struct {
	// Network is only used internally for generating SRT settings JSON output.
	// It is NOT user-configurable from sandbox.agent.config (yaml:"-" prevents parsing).
	// The json tag is needed for output serialization to .srt-settings.json.
	Network                   *SRTNetworkConfig    `yaml:"-" json:"network,omitempty"`
	Filesystem                *SRTFilesystemConfig `yaml:"filesystem,omitempty" json:"filesystem,omitempty"`
	IgnoreViolations          map[string][]string  `yaml:"ignoreViolations,omitempty" json:"ignoreViolations,omitempty"`
	EnableWeakerNestedSandbox bool                 `yaml:"enableWeakerNestedSandbox" json:"enableWeakerNestedSandbox"`
}

// SRTNetworkConfig represents network configuration for SRT
type SRTNetworkConfig struct {
	AllowedDomains      []string `yaml:"allowedDomains,omitempty" json:"allowedDomains,omitempty"`
	BlockedDomains      []string `yaml:"blockedDomains,omitempty" json:"blockedDomains"`
	AllowUnixSockets    []string `yaml:"allowUnixSockets,omitempty" json:"allowUnixSockets,omitempty"`
	AllowLocalBinding   bool     `yaml:"allowLocalBinding" json:"allowLocalBinding"`
	AllowAllUnixSockets bool     `yaml:"allowAllUnixSockets" json:"allowAllUnixSockets"`
}

// SRTFilesystemConfig represents filesystem configuration for SRT
type SRTFilesystemConfig struct {
	DenyRead   []string `yaml:"denyRead" json:"denyRead"`
	AllowWrite []string `yaml:"allowWrite,omitempty" json:"allowWrite,omitempty"`
	DenyWrite  []string `yaml:"denyWrite" json:"denyWrite"`
}

// getAgentType returns the effective agent type from AgentSandboxConfig
// Prefers ID field (new format) over Type field (legacy)
// Returns "awf" for AWF sandbox, "default" for default, or "" if not set
func getAgentType(agent *AgentSandboxConfig) SandboxType {
	if agent == nil {
		return ""
	}
	// New format: use ID field if set
	if agent.ID != "" {
		return SandboxType(agent.ID)
	}
	// Legacy format: use Type field
	return agent.Type
}

// isSupportedSandboxType checks if a sandbox type is valid/supported
// Only "awf" and "default" (alias for awf) are supported
func isSupportedSandboxType(sandboxType SandboxType) bool {
	return sandboxType == SandboxTypeAWF ||
		sandboxType == SandboxTypeDefault
}

// migrateSRTToAWF converts any deprecated sandbox configuration to AWF
// This is a codemod that automatically migrates workflows to use AWF
func migrateSRTToAWF(sandboxConfig *SandboxConfig) *SandboxConfig {
	if sandboxConfig == nil {
		return nil
	}

	// Migrate legacy Type field from deprecated values to AWF
	if sandboxConfig.Type == "srt" || sandboxConfig.Type == "sandbox-runtime" {
		sandboxLog.Printf("Migrating legacy sandbox type from %s to awf", sandboxConfig.Type)
		sandboxConfig.Type = SandboxTypeAWF
	}

	// Migrate Agent.Type field from deprecated values to AWF
	if sandboxConfig.Agent != nil {
		if sandboxConfig.Agent.Type == "srt" || sandboxConfig.Agent.Type == "sandbox-runtime" {
			sandboxLog.Printf("Migrating agent type from %s to awf", sandboxConfig.Agent.Type)
			sandboxConfig.Agent.Type = SandboxTypeAWF
		}
		// Migrate Agent.ID field from deprecated values to AWF
		if sandboxConfig.Agent.ID == "srt" || sandboxConfig.Agent.ID == "sandbox-runtime" {
			sandboxLog.Printf("Migrating agent ID from %s to awf", sandboxConfig.Agent.ID)
			sandboxConfig.Agent.ID = "awf"
		}
	}

	return sandboxConfig
}

// applySandboxDefaults applies default values to sandbox configuration
// If no sandbox config exists, creates one with AWF as default agent
// If sandbox config exists but has no agent, sets agent to AWF (unless agent is explicitly disabled)
func applySandboxDefaults(sandboxConfig *SandboxConfig, engineConfig *EngineConfig) *SandboxConfig {
	// First, migrate any deprecated references to AWF (codemod)
	sandboxConfig = migrateSRTToAWF(sandboxConfig)

	// If agent sandbox is explicitly disabled (sandbox.agent: false), preserve that setting
	if sandboxConfig != nil && sandboxConfig.Agent != nil && sandboxConfig.Agent.Disabled {
		sandboxLog.Print("Agent sandbox explicitly disabled with sandbox.agent: false, preserving disabled state")
		return sandboxConfig
	}

	// If no sandbox config exists, create one with AWF as default
	if sandboxConfig == nil {
		sandboxLog.Print("No sandbox config found, creating default with agent: awf")
		return &SandboxConfig{
			Agent: &AgentSandboxConfig{
				Type: SandboxTypeAWF,
			},
		}
	}

	// If sandbox config exists with legacy Type field set, don't override with AWF default
	// The legacy Type field indicates explicit sandbox configuration
	if sandboxConfig.Type != "" {
		sandboxLog.Printf("Sandbox config uses legacy Type field: %s, preserving it", sandboxConfig.Type)
		return sandboxConfig
	}

	// If sandbox config exists but has no agent, set agent to AWF
	if sandboxConfig.Agent == nil {
		sandboxLog.Print("Sandbox config exists without agent, setting default agent: awf")
		sandboxConfig.Agent = &AgentSandboxConfig{
			Type: SandboxTypeAWF,
		}
	}

	return sandboxConfig
}
