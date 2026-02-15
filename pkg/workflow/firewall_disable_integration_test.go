//go:build integration

package workflow

import (
	"strings"
	"testing"
)

func TestFirewallDisableIntegration(t *testing.T) {
	t.Run("firewall disable with allowed domains warns", func(t *testing.T) {
		frontmatter := map[string]any{
			"on":     "workflow_dispatch",
			"engine": "copilot",
			"network": map[string]any{
				"allowed":  []any{"example.com"},
				"firewall": "disable",
			},
			"tools": map[string]any{
				"web-fetch": nil,
			},
		}

		compiler := NewCompiler(
			WithVersion("test"),
			WithSkipValidation(true),
		)

		// Extract network permissions
		networkPerms := compiler.extractNetworkPermissions(frontmatter)
		if networkPerms == nil {
			t.Fatal("Expected network permissions to be extracted")
		}

		// Check firewall config - should be nil when disabled
		if networkPerms.Firewall != nil {
			t.Error("Firewall should be nil (disabled) when set to 'disable'")
		}

		// Check validation - no warnings since checkFirewallDisable is now a no-op
		engine := NewCopilotEngine()
		initialWarnings := compiler.warningCount
		err := compiler.checkFirewallDisable(engine, networkPerms)
		if err != nil {
			t.Errorf("Expected no error in non-strict mode, got: %v", err)
		}
		// checkFirewallDisable is now a no-op since we removed Enabled field
		if compiler.warningCount != initialWarnings {
			t.Error("Should NOT emit warning since checkFirewallDisable is now a no-op")
		}
	})

	t.Run("firewall disable in strict mode - no longer validates in checkFirewallDisable", func(t *testing.T) {
		frontmatter := map[string]any{
			"on":     "workflow_dispatch",
			"engine": "copilot",
			"strict": true,
			"network": map[string]any{
				"allowed":  []any{"example.com"},
				"firewall": "disable",
			},
		}

		compiler := NewCompiler()
		compiler.strictMode = true
		compiler.SetSkipValidation(true)

		networkPerms := compiler.extractNetworkPermissions(frontmatter)
		if networkPerms == nil {
			t.Fatal("Expected network permissions to be extracted")
		}

		// Firewall should be nil when disabled
		if networkPerms.Firewall != nil {
			t.Error("Firewall should be nil when set to 'disable'")
		}

		// checkFirewallDisable is now a no-op, no error expected
		engine := NewCopilotEngine()
		err := compiler.checkFirewallDisable(engine, networkPerms)
		if err != nil {
			t.Errorf("checkFirewallDisable is now a no-op, expected no error, got: %v", err)
		}
	})
}
