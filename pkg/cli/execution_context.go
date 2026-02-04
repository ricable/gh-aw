package cli

import (
	"fmt"
	"sync"
)

// executionContext tracks the current command execution context
// This is used to enforce security boundaries, such as preventing
// secret modifications during the upgrade command.
type executionContext struct {
	mu              sync.RWMutex
	currentCommand  string
	allowSecretMods bool
}

var globalExecutionContext = &executionContext{
	allowSecretMods: true, // Default: allow secret modifications
}

// SetUpgradeContext marks that we're running the upgrade command
// This disables secret modifications as a security measure
func SetUpgradeContext() func() {
	globalExecutionContext.mu.Lock()
	defer globalExecutionContext.mu.Unlock()

	// Save previous state
	prevCommand := globalExecutionContext.currentCommand
	prevAllow := globalExecutionContext.allowSecretMods

	// Set upgrade context
	globalExecutionContext.currentCommand = "upgrade"
	globalExecutionContext.allowSecretMods = false

	// Return cleanup function to restore previous state
	return func() {
		globalExecutionContext.mu.Lock()
		defer globalExecutionContext.mu.Unlock()
		globalExecutionContext.currentCommand = prevCommand
		globalExecutionContext.allowSecretMods = prevAllow
	}
}

// CheckSecretModificationAllowed checks if secret modifications are allowed
// in the current execution context. Returns an error if they are not allowed.
func CheckSecretModificationAllowed() error {
	globalExecutionContext.mu.RLock()
	defer globalExecutionContext.mu.RUnlock()

	if !globalExecutionContext.allowSecretMods {
		return fmt.Errorf("secret modifications are not allowed during %s command execution", globalExecutionContext.currentCommand)
	}
	return nil
}

// GetCurrentCommand returns the currently executing command name
func GetCurrentCommand() string {
	globalExecutionContext.mu.RLock()
	defer globalExecutionContext.mu.RUnlock()
	return globalExecutionContext.currentCommand
}
