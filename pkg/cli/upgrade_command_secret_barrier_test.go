//go:build !integration

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpgradeCommand_BlocksSecretModification(t *testing.T) {
	// Ensure we start in a clean state
	globalExecutionContext.currentCommand = ""
	globalExecutionContext.allowSecretMods = true

	// Verify that secrets can be modified before upgrade context
	err := CheckSecretModificationAllowed()
	require.NoError(t, err, "Should allow secret modifications before upgrade")

	// Set upgrade context
	cleanup := SetUpgradeContext()
	defer cleanup()

	// Verify that secret modifications are now blocked
	err = CheckSecretModificationAllowed()
	require.Error(t, err, "Should block secret modifications during upgrade")
	assert.Contains(t, err.Error(), "upgrade", "Error message should mention upgrade")
	assert.Contains(t, err.Error(), "not allowed", "Error message should indicate operation is not allowed")
}

func TestSetRepoSecret_BlockedDuringUpgrade(t *testing.T) {
	// This test verifies that setRepoSecret checks the execution context
	// We can't easily test the full function without mocking the GitHub API,
	// but we can verify the guard is in place by setting upgrade context
	// and checking that the error propagates

	// Ensure we start in a clean state
	globalExecutionContext.currentCommand = ""
	globalExecutionContext.allowSecretMods = true

	// Set upgrade context
	cleanup := SetUpgradeContext()
	defer cleanup()

	// Attempt to call setRepoSecret - it should fail immediately with context check
	// Note: We pass nil for client which would normally cause an error later,
	// but the context check should fail first
	err := setRepoSecret(nil, "owner", "repo", "TEST_SECRET", "value")

	require.Error(t, err, "Should fail when trying to set secret during upgrade")
	assert.Contains(t, err.Error(), "upgrade", "Error should mention upgrade command")
}

func TestAttemptSetSecret_BlockedDuringUpgrade(t *testing.T) {
	// This test verifies that attemptSetSecret checks the execution context

	// Ensure we start in a clean state
	globalExecutionContext.currentCommand = ""
	globalExecutionContext.allowSecretMods = true

	// Set upgrade context
	cleanup := SetUpgradeContext()
	defer cleanup()

	// Attempt to call attemptSetSecret - it should fail immediately with context check
	err := attemptSetSecret("TEST_SECRET", "owner/repo", false)

	require.Error(t, err, "Should fail when trying to set secret during upgrade")
	assert.Contains(t, err.Error(), "upgrade", "Error should mention upgrade command")
}

func TestUpgradeCommand_ContextCleanupOnError(t *testing.T) {
	// Ensure context is cleaned up even if upgrade fails
	// We'll simulate this by setting the context and cleaning up

	// Ensure we start in a clean state
	globalExecutionContext.currentCommand = ""
	globalExecutionContext.allowSecretMods = true

	// Set upgrade context
	cleanup := SetUpgradeContext()

	// Verify upgrade context is active
	err := CheckSecretModificationAllowed()
	require.Error(t, err, "Should block secrets during upgrade")

	// Simulate cleanup (what defer does in runUpgradeCommand)
	cleanup()

	// Verify context is cleaned up
	err = CheckSecretModificationAllowed()
	assert.NoError(t, err, "Should allow secrets after cleanup")
}
