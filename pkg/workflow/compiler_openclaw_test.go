//go:build integration

package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/github/gh-aw/pkg/stringutil"
	"github.com/github/gh-aw/pkg/testutil"
)

func TestCompileOpenClawBasicWorkflow(t *testing.T) {
	tmpDir := testutil.TempDir(t, "openclaw-compile-test")

	testContent := `---
on: push
timeout-minutes: 10
permissions:
  contents: read
  issues: write
  pull-requests: read
engine: openclaw
strict: false
features:
  dangerous-permissions-write: true
  experimental-engines: true
tools:
  github:
    allowed: [list_issues, create_issue]
  bash: ["echo", "ls"]
---

# OpenClaw Test Workflow

This is a test workflow using the OpenClaw engine.
`

	testFile := filepath.Join(tmpDir, "openclaw-basic.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler()
	err := compiler.CompileWorkflow(testFile)
	if err != nil {
		t.Fatalf("Expected OpenClaw workflow to compile successfully, got error: %v", err)
	}

	// Verify lock file was created
	lockFile := stringutil.MarkdownToLockFile(testFile)
	lockContent, err := os.ReadFile(lockFile)
	if err != nil {
		t.Fatalf("Failed to read lock file: %v", err)
	}

	lockStr := string(lockContent)

	// Verify the workflow name is present
	if !strings.Contains(lockStr, "OpenClaw Test Workflow") {
		t.Error("Expected workflow name 'OpenClaw Test Workflow' in lock file")
	}

	// Verify OpenClaw installation step
	if !strings.Contains(lockStr, "Install OpenClaw") {
		t.Error("Expected 'Install OpenClaw' step in lock file")
	}

	// Verify OpenClaw execution step
	if !strings.Contains(lockStr, "Run OpenClaw") {
		t.Error("Expected 'Run OpenClaw' step in lock file")
	}

	// Verify openclaw CLI command is present
	if !strings.Contains(lockStr, "openclaw") {
		t.Error("Expected 'openclaw' command in lock file")
	}

	// Verify agent subcommand
	if !strings.Contains(lockStr, "agent") {
		t.Error("Expected 'agent' subcommand in lock file")
	}

	// Verify key CLI flags
	if !strings.Contains(lockStr, "--local") {
		t.Error("Expected '--local' flag in lock file")
	}

	if !strings.Contains(lockStr, "--json") {
		t.Error("Expected '--json' flag in lock file")
	}

	if !strings.Contains(lockStr, "--no-color") {
		t.Error("Expected '--no-color' flag in lock file")
	}

	// Verify environment variables
	if !strings.Contains(lockStr, "OPENCLAW_API_KEY") {
		t.Error("Expected OPENCLAW_API_KEY in lock file")
	}

	if !strings.Contains(lockStr, "ANTHROPIC_API_KEY") {
		t.Error("Expected ANTHROPIC_API_KEY in lock file")
	}

	if !strings.Contains(lockStr, "OPENCLAW_STATE_DIR") {
		t.Error("Expected OPENCLAW_STATE_DIR in lock file")
	}

	// Verify log parser reference
	if !strings.Contains(lockStr, "parse_openclaw_log") {
		t.Error("Expected 'parse_openclaw_log' log parser reference in lock file")
	}
}

func TestCompileOpenClawEngineObject(t *testing.T) {
	tmpDir := testutil.TempDir(t, "openclaw-engine-obj-test")

	testContent := `---
on: push
timeout-minutes: 15
permissions:
  contents: read
  issues: write
  pull-requests: read
engine:
  id: openclaw
  model: custom-agent-v2
  args:
    - "--verbose"
    - "--timeout"
    - "1800"
strict: false
features:
  dangerous-permissions-write: true
  experimental-engines: true
tools:
  bash: ["*"]
---

# OpenClaw Custom Agent Workflow

Test workflow with engine object configuration including model and args.
`

	testFile := filepath.Join(tmpDir, "openclaw-engine-obj.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler()
	err := compiler.CompileWorkflow(testFile)
	if err != nil {
		t.Fatalf("Expected OpenClaw workflow with engine object to compile successfully, got error: %v", err)
	}

	lockFile := stringutil.MarkdownToLockFile(testFile)
	lockContent, err := os.ReadFile(lockFile)
	if err != nil {
		t.Fatalf("Failed to read lock file: %v", err)
	}

	lockStr := string(lockContent)

	// Verify custom model/agent is present
	if !strings.Contains(lockStr, "--agent") {
		t.Error("Expected '--agent' flag in lock file for custom model")
	}

	if !strings.Contains(lockStr, "custom-agent-v2") {
		t.Error("Expected 'custom-agent-v2' agent name in lock file")
	}

	// Verify custom args are passed through
	if !strings.Contains(lockStr, "--verbose") {
		t.Error("Expected '--verbose' custom arg in lock file")
	}
}

func TestCompileOpenClawWithFirewall(t *testing.T) {
	tmpDir := testutil.TempDir(t, "openclaw-firewall-test")

	testContent := `---
on: push
timeout-minutes: 10
permissions:
  contents: read
  issues: read
  pull-requests: read
engine: openclaw
strict: false
features:
  experimental-engines: true
network:
  allowed:
    - defaults
    - github
tools:
  github:
    toolsets: [repos]
  bash: ["echo"]
---

# OpenClaw Firewall Workflow

Test workflow with firewall enabled.
`

	testFile := filepath.Join(tmpDir, "openclaw-firewall.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler()
	err := compiler.CompileWorkflow(testFile)
	if err != nil {
		t.Fatalf("Expected OpenClaw workflow with firewall to compile successfully, got error: %v", err)
	}

	lockFile := stringutil.MarkdownToLockFile(testFile)
	lockContent, err := os.ReadFile(lockFile)
	if err != nil {
		t.Fatalf("Failed to read lock file: %v", err)
	}

	lockStr := string(lockContent)

	// Verify AWF (firewall) command is present
	if !strings.Contains(lockStr, "awf") {
		t.Error("Expected 'awf' command in lock file when firewall is enabled")
	}

	// Verify allowed domains flag
	if !strings.Contains(lockStr, "--allow-domains") {
		t.Error("Expected '--allow-domains' flag in lock file when firewall is enabled")
	}

	// Verify the openclaw command is still in the wrapped command
	if !strings.Contains(lockStr, "openclaw") {
		t.Error("Expected 'openclaw' command inside AWF wrapper")
	}
}

func TestCompileOpenClawWithSafeOutputs(t *testing.T) {
	tmpDir := testutil.TempDir(t, "openclaw-safe-outputs-test")

	testContent := `---
on: push
timeout-minutes: 10
permissions:
  contents: read
  issues: write
  pull-requests: write
engine: openclaw
strict: false
features:
  dangerous-permissions-write: true
  experimental-engines: true
tools:
  github:
    toolsets: [repos]
  bash: ["echo"]
safe-outputs:
  add-comment:
    max: 2
  create-issue:
    expires: 2h
---

# OpenClaw Safe Outputs Workflow

Test workflow with safe outputs configured.
`

	testFile := filepath.Join(tmpDir, "openclaw-safe-outputs.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler()
	err := compiler.CompileWorkflow(testFile)
	if err != nil {
		t.Fatalf("Expected OpenClaw workflow with safe outputs to compile successfully, got error: %v", err)
	}

	lockFile := stringutil.MarkdownToLockFile(testFile)
	lockContent, err := os.ReadFile(lockFile)
	if err != nil {
		t.Fatalf("Failed to read lock file: %v", err)
	}

	lockStr := string(lockContent)

	// Verify safe outputs environment variable
	if !strings.Contains(lockStr, "GH_AW_SAFE_OUTPUTS") {
		t.Error("Expected GH_AW_SAFE_OUTPUTS in lock file when safe-outputs is configured")
	}
}

func TestCompileOpenClawWithCustomCommand(t *testing.T) {
	tmpDir := testutil.TempDir(t, "openclaw-custom-cmd-test")

	testContent := `---
on: push
timeout-minutes: 10
permissions:
  contents: read
  issues: read
  pull-requests: read
engine:
  id: openclaw
  command: /usr/local/bin/my-openclaw
strict: false
features:
  experimental-engines: true
tools:
  bash: ["echo"]
---

# OpenClaw Custom Command Workflow

Test workflow with a custom openclaw command path.
`

	testFile := filepath.Join(tmpDir, "openclaw-custom-cmd.md")
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}

	compiler := NewCompiler()
	err := compiler.CompileWorkflow(testFile)
	if err != nil {
		t.Fatalf("Expected OpenClaw workflow with custom command to compile successfully, got error: %v", err)
	}

	lockFile := stringutil.MarkdownToLockFile(testFile)
	lockContent, err := os.ReadFile(lockFile)
	if err != nil {
		t.Fatalf("Failed to read lock file: %v", err)
	}

	lockStr := string(lockContent)

	// Verify custom command is used instead of default 'openclaw'
	if !strings.Contains(lockStr, "/usr/local/bin/my-openclaw") {
		t.Error("Expected custom command '/usr/local/bin/my-openclaw' in lock file")
	}

	// With custom command, installation steps should be skipped
	if strings.Contains(lockStr, "Install OpenClaw") {
		t.Error("Expected 'Install OpenClaw' step to be absent when custom command is specified")
	}
}
