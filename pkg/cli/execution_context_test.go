//go:build !integration

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecutionContext_Default(t *testing.T) {
	// Default state should allow secret modifications
	err := CheckSecretModificationAllowed()
	require.NoError(t, err, "Should allow secret modifications by default")

	cmd := GetCurrentCommand()
	assert.Empty(t, cmd, "Should have no current command by default")
}

func TestExecutionContext_SetUpgradeContext(t *testing.T) {
	// Ensure we start in a clean state
	globalExecutionContext.currentCommand = ""
	globalExecutionContext.allowSecretMods = true

	// Set upgrade context
	cleanup := SetUpgradeContext()
	defer cleanup()

	// Check that secret modifications are blocked
	err := CheckSecretModificationAllowed()
	require.Error(t, err, "Should not allow secret modifications in upgrade context")
	assert.Contains(t, err.Error(), "upgrade", "Error should mention upgrade command")

	cmd := GetCurrentCommand()
	assert.Equal(t, "upgrade", cmd, "Should be in upgrade command context")
}

func TestExecutionContext_RestoreAfterCleanup(t *testing.T) {
	// Ensure we start in a clean state
	globalExecutionContext.currentCommand = ""
	globalExecutionContext.allowSecretMods = true

	// Set upgrade context and immediately clean up
	cleanup := SetUpgradeContext()

	// Verify upgrade context is set
	err := CheckSecretModificationAllowed()
	require.Error(t, err, "Should block secrets in upgrade context")

	// Clean up
	cleanup()

	// Verify we're back to default state
	err = CheckSecretModificationAllowed()
	require.NoError(t, err, "Should allow secret modifications after cleanup")

	cmd := GetCurrentCommand()
	assert.Empty(t, cmd, "Should have no current command after cleanup")
}

func TestExecutionContext_NestedContexts(t *testing.T) {
	// Ensure we start in a clean state
	globalExecutionContext.currentCommand = ""
	globalExecutionContext.allowSecretMods = true

	// First context
	cleanup1 := SetUpgradeContext()
	defer cleanup1()

	err := CheckSecretModificationAllowed()
	require.Error(t, err, "Should block secrets in upgrade context")

	// Second context (nested)
	cleanup2 := SetUpgradeContext()

	err = CheckSecretModificationAllowed()
	require.Error(t, err, "Should still block secrets in nested context")

	// Clean up second context
	cleanup2()

	// Should still be in first context
	err = CheckSecretModificationAllowed()
	require.Error(t, err, "Should still block secrets after cleaning up nested context")

	// Clean up first context
	cleanup1()

	// Now should be back to default
	err = CheckSecretModificationAllowed()
	assert.NoError(t, err, "Should allow secret modifications after all cleanups")
}
