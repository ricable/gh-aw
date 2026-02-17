//go:build !integration

package workflow

import (
	"strings"
	"testing"

	"github.com/github/gh-aw/pkg/constants"
)

func TestOpenClawEngine(t *testing.T) {
	engine := NewOpenClawEngine()

	// Test basic properties
	if engine.GetID() != "openclaw" {
		t.Errorf("Expected ID 'openclaw', got '%s'", engine.GetID())
	}

	if engine.GetDisplayName() != "OpenClaw" {
		t.Errorf("Expected display name 'OpenClaw', got '%s'", engine.GetDisplayName())
	}

	if engine.GetDescription() != "Uses OpenClaw agent platform with ACP tool support" {
		t.Errorf("Expected correct description, got '%s'", engine.GetDescription())
	}

	if !engine.IsExperimental() {
		t.Error("OpenClaw engine should be experimental")
	}

	if engine.SupportsToolsAllowlist() {
		t.Error("OpenClaw engine should not support tools allowlist")
	}

	if engine.SupportsMaxTurns() {
		t.Error("OpenClaw engine should not support max turns")
	}

	if engine.SupportsWebFetch() {
		t.Error("OpenClaw engine should not support web fetch")
	}

	if engine.SupportsWebSearch() {
		t.Error("OpenClaw engine should not support web search")
	}

	if !engine.SupportsFirewall() {
		t.Error("OpenClaw engine should support firewall")
	}

	if engine.SupportsLLMGateway() != -1 {
		t.Errorf("Expected LLM gateway port -1, got %d", engine.SupportsLLMGateway())
	}
}

func TestOpenClawEngine_GetRequiredSecretNames(t *testing.T) {
	engine := NewOpenClawEngine()

	// Test basic secrets (no MCP servers)
	workflowData := &WorkflowData{}
	secrets := engine.GetRequiredSecretNames(workflowData)

	if len(secrets) != 2 {
		t.Errorf("Expected 2 secrets, got %d", len(secrets))
	}

	hasOpenClawKey := false
	hasAnthropicKey := false
	for _, s := range secrets {
		if s == "OPENCLAW_API_KEY" {
			hasOpenClawKey = true
		}
		if s == "ANTHROPIC_API_KEY" {
			hasAnthropicKey = true
		}
	}
	if !hasOpenClawKey {
		t.Error("Expected OPENCLAW_API_KEY in secrets")
	}
	if !hasAnthropicKey {
		t.Error("Expected ANTHROPIC_API_KEY in secrets")
	}

	// Test with MCP servers
	workflowData = &WorkflowData{
		Tools: map[string]any{
			"github": nil,
		},
		ParsedTools: &ToolsConfig{
			GitHub: &GitHubToolConfig{},
		},
	}
	secrets = engine.GetRequiredSecretNames(workflowData)

	hasMCPKey := false
	for _, s := range secrets {
		if s == "MCP_GATEWAY_API_KEY" {
			hasMCPKey = true
		}
	}
	if !hasMCPKey {
		t.Error("Expected MCP_GATEWAY_API_KEY in secrets when MCP servers are present")
	}
}

func TestOpenClawEngine_GetInstallationSteps(t *testing.T) {
	engine := NewOpenClawEngine()

	steps := engine.GetInstallationSteps(&WorkflowData{})
	expectedStepCount := 3 // Secret validation + Node.js setup + Install OpenClaw
	if len(steps) != expectedStepCount {
		t.Errorf("Expected %d installation steps, got %d", expectedStepCount, len(steps))
	}

	// Verify first step is secret validation
	if len(steps) > 0 && len(steps[0]) > 0 {
		if !strings.Contains(steps[0][0], "Validate OPENCLAW_API_KEY or ANTHROPIC_API_KEY secret") {
			t.Errorf("Expected first step to contain 'Validate OPENCLAW_API_KEY or ANTHROPIC_API_KEY secret', got '%s'", steps[0][0])
		}
	}

	// Verify second step is Node.js setup
	if len(steps) > 1 && len(steps[1]) > 0 {
		if !strings.Contains(steps[1][0], "Setup Node.js") {
			t.Errorf("Expected second step to contain 'Setup Node.js', got '%s'", steps[1][0])
		}
	}

	// Verify third step installs OpenClaw
	if len(steps) > 2 && len(steps[2]) > 0 {
		if !strings.Contains(steps[2][0], "Install OpenClaw") {
			t.Errorf("Expected third step to contain 'Install OpenClaw', got '%s'", steps[2][0])
		}
	}
}

func TestOpenClawEngine_GetInstallationSteps_CustomCommand(t *testing.T) {
	engine := NewOpenClawEngine()

	// Custom command should skip installation
	steps := engine.GetInstallationSteps(&WorkflowData{
		EngineConfig: &EngineConfig{
			Command: "/usr/bin/custom-openclaw",
		},
	})

	if len(steps) != 0 {
		t.Errorf("Expected 0 installation steps with custom command, got %d", len(steps))
	}
}

func TestOpenClawEngine_GetExecutionSteps(t *testing.T) {
	engine := NewOpenClawEngine()

	workflowData := &WorkflowData{
		Name: "test-workflow",
	}
	execSteps := engine.GetExecutionSteps(workflowData, "/tmp/gh-aw/agent-stdio.log")
	if len(execSteps) != 1 {
		t.Fatalf("Expected 1 step for OpenClaw execution, got %d", len(execSteps))
	}

	stepContent := strings.Join([]string(execSteps[0]), "\n")

	if !strings.Contains(stepContent, "name: Run OpenClaw") {
		t.Errorf("Expected step name 'Run OpenClaw' in step content")
	}

	if !strings.Contains(stepContent, "openclaw") {
		t.Errorf("Expected command to contain 'openclaw' in step content")
	}

	if !strings.Contains(stepContent, "agent") {
		t.Errorf("Expected command to contain 'agent' subcommand in step content")
	}

	if !strings.Contains(stepContent, "--local") {
		t.Errorf("Expected command to contain '--local' flag")
	}

	if !strings.Contains(stepContent, "--json") {
		t.Errorf("Expected command to contain '--json' flag")
	}

	if !strings.Contains(stepContent, "--no-color") {
		t.Errorf("Expected command to contain '--no-color' flag")
	}

	if !strings.Contains(stepContent, "--timeout") {
		t.Errorf("Expected command to contain '--timeout' flag")
	}

	if !strings.Contains(stepContent, "set -o pipefail") {
		t.Errorf("Expected command to contain 'set -o pipefail'")
	}

	// Check environment variables
	if !strings.Contains(stepContent, "OPENCLAW_API_KEY") {
		t.Errorf("Expected OPENCLAW_API_KEY environment variable")
	}

	if !strings.Contains(stepContent, "ANTHROPIC_API_KEY") {
		t.Errorf("Expected ANTHROPIC_API_KEY environment variable")
	}

	if !strings.Contains(stepContent, "OPENCLAW_STATE_DIR") {
		t.Errorf("Expected OPENCLAW_STATE_DIR environment variable")
	}

	if !strings.Contains(stepContent, "DISABLE_TELEMETRY") {
		t.Errorf("Expected DISABLE_TELEMETRY environment variable")
	}
}

func TestOpenClawEngine_GetExecutionSteps_WithFirewall(t *testing.T) {
	engine := NewOpenClawEngine()

	workflowData := &WorkflowData{
		Name: "test-workflow-firewall",
		NetworkPermissions: &NetworkPermissions{
			Allowed: []string{"defaults"},
			Firewall: &FirewallConfig{
				Enabled: true,
			},
		},
	}
	execSteps := engine.GetExecutionSteps(workflowData, "/tmp/gh-aw/agent-stdio.log")
	if len(execSteps) != 1 {
		t.Fatalf("Expected 1 step for OpenClaw execution with firewall, got %d", len(execSteps))
	}

	stepContent := strings.Join([]string(execSteps[0]), "\n")

	// Verify AWF command is present
	if !strings.Contains(stepContent, "awf") {
		t.Errorf("Expected AWF command in step content with firewall enabled")
	}

	if !strings.Contains(stepContent, "--allow-domains") {
		t.Errorf("Expected --allow-domains in step content with firewall enabled")
	}
}

func TestOpenClawEngine_GetExecutionSteps_WithAgentFile(t *testing.T) {
	engine := NewOpenClawEngine()

	workflowData := &WorkflowData{
		Name:      "test-workflow-agent-file",
		AgentFile: ".github/agents/my-agent.md",
	}
	execSteps := engine.GetExecutionSteps(workflowData, "/tmp/gh-aw/agent-stdio.log")
	if len(execSteps) != 1 {
		t.Fatalf("Expected 1 step for OpenClaw execution with agent file, got %d", len(execSteps))
	}

	stepContent := strings.Join([]string(execSteps[0]), "\n")

	// Verify agent file handling
	if !strings.Contains(stepContent, "AGENT_CONTENT") {
		t.Errorf("Expected AGENT_CONTENT variable in step content with agent file")
	}

	if !strings.Contains(stepContent, "awk") {
		t.Errorf("Expected awk command for frontmatter stripping in step content")
	}
}

func TestOpenClawEngine_GetExecutionSteps_WithModel(t *testing.T) {
	engine := NewOpenClawEngine()

	workflowData := &WorkflowData{
		Name: "test-workflow-model",
		EngineConfig: &EngineConfig{
			Model: "custom-agent",
		},
	}
	execSteps := engine.GetExecutionSteps(workflowData, "/tmp/gh-aw/agent-stdio.log")
	if len(execSteps) != 1 {
		t.Fatalf("Expected 1 step for OpenClaw execution with model, got %d", len(execSteps))
	}

	stepContent := strings.Join([]string(execSteps[0]), "\n")

	if !strings.Contains(stepContent, "--agent") {
		t.Errorf("Expected '--agent' flag when model is specified")
	}

	if !strings.Contains(stepContent, "custom-agent") {
		t.Errorf("Expected 'custom-agent' value in step content")
	}
}

func TestOpenClawEngine_GetLogParserScriptId(t *testing.T) {
	engine := NewOpenClawEngine()

	script := engine.GetLogParserScriptId()
	if script != "parse_openclaw_log" {
		t.Errorf("Expected log parser script 'parse_openclaw_log', got '%s'", script)
	}
}

func TestOpenClawEngine_GetDeclaredOutputFiles(t *testing.T) {
	engine := NewOpenClawEngine()

	files := engine.GetDeclaredOutputFiles()
	if len(files) != 0 {
		t.Errorf("Expected 0 declared output files, got %d", len(files))
	}
}

func TestOpenClawEngine_EngineRegistered(t *testing.T) {
	registry := NewEngineRegistry()

	engine, err := registry.GetEngine("openclaw")
	if err != nil {
		t.Fatalf("Expected openclaw engine to be registered, got error: %v", err)
	}

	if engine.GetID() != "openclaw" {
		t.Errorf("Expected engine ID 'openclaw', got '%s'", engine.GetID())
	}
}

func TestOpenClawEngine_Constants(t *testing.T) {
	// Verify constants are defined
	if string(constants.OpenClawEngine) != "openclaw" {
		t.Errorf("Expected OpenClawEngine constant to be 'openclaw', got '%s'", string(constants.OpenClawEngine))
	}

	if string(constants.DefaultOpenClawVersion) == "" {
		t.Error("Expected DefaultOpenClawVersion to be non-empty")
	}

	// Verify openclaw is in AgenticEngines
	found := false
	for _, engine := range constants.AgenticEngines {
		if engine == "openclaw" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'openclaw' to be in AgenticEngines")
	}

	// Verify EngineOptions contains openclaw
	option := constants.GetEngineOption("openclaw")
	if option == nil {
		t.Error("Expected openclaw to be in EngineOptions")
	}
	if option != nil && option.Label != "OpenClaw" {
		t.Errorf("Expected EngineOption label 'OpenClaw', got '%s'", option.Label)
	}
}

func TestOpenClawEngine_ParseLogMetrics(t *testing.T) {
	engine := NewOpenClawEngine()

	// Test with empty log
	metrics := engine.ParseLogMetrics("", false)
	if len(metrics.ToolCalls) != 0 {
		t.Errorf("Expected 0 tool calls for empty log, got %d", len(metrics.ToolCalls))
	}

	// Test with tool calls in JSON
	logContent := `{"type":"tool_call","name":"test_tool"}
{"type":"message","content":"thinking..."}
{"type":"tool_call","name":"another_tool"}`
	metrics = engine.ParseLogMetrics(logContent, false)
	if len(metrics.ToolCalls) != 2 {
		t.Errorf("Expected 2 tool call types, got %d", len(metrics.ToolCalls))
	}

	// Test with repeated tool calls
	logContent2 := `{"type":"tool_call","name":"test_tool"}
{"type":"tool_call","name":"test_tool"}
{"type":"tool_call","name":"other_tool"}`
	metrics2 := engine.ParseLogMetrics(logContent2, false)
	if len(metrics2.ToolCalls) != 2 {
		t.Errorf("Expected 2 unique tool types, got %d", len(metrics2.ToolCalls))
	}
	// Check total call count
	totalCalls := 0
	for _, tc := range metrics2.ToolCalls {
		totalCalls += tc.CallCount
	}
	if totalCalls != 3 {
		t.Errorf("Expected 3 total tool calls, got %d", totalCalls)
	}
}

func TestOpenClawEngine_GetDefaultDetectionModel(t *testing.T) {
	engine := NewOpenClawEngine()

	model := engine.GetDefaultDetectionModel()
	if model != "" {
		t.Errorf("Expected empty default detection model, got '%s'", model)
	}
}
