//go:build !integration

package cli

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUpdateWorkflowsWithExtensionCheckContext_Cancellation tests that the context cancellation is respected
func TestUpdateWorkflowsWithExtensionCheckContext_Cancellation(t *testing.T) {
	// Create a context that is already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Call the function with the cancelled context
	err := UpdateWorkflowsWithExtensionCheckContext(
		ctx,
		[]string{}, // workflowNames
		false,      // allowMajor
		false,      // force
		false,      // verbose
		"",         // engineOverride
		false,      // createPR
		"",         // workflowsDir
		false,      // noStopAfter
		"",         // stopAfter
		false,      // merge
		false,      // noActions
	)

	// Verify that the error is context.Canceled
	require.Error(t, err, "Expected an error when context is cancelled")
	assert.Equal(t, context.Canceled, err, "Expected context.Canceled error")
}

// TestUpdateWorkflowsWithExtensionCheckContext_Timeout tests timeout handling
func TestUpdateWorkflowsWithExtensionCheckContext_Timeout(t *testing.T) {
	// Create a context with a very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait a moment to ensure the timeout is triggered
	time.Sleep(10 * time.Millisecond)

	// Call the function with the timed-out context
	err := UpdateWorkflowsWithExtensionCheckContext(
		ctx,
		[]string{}, // workflowNames
		false,      // allowMajor
		false,      // force
		false,      // verbose
		"",         // engineOverride
		false,      // createPR
		"",         // workflowsDir
		false,      // noStopAfter
		"",         // stopAfter
		false,      // merge
		false,      // noActions
	)

	// Verify that the error is context.DeadlineExceeded
	require.Error(t, err, "Expected an error when context times out")
	assert.Equal(t, context.DeadlineExceeded, err, "Expected context.DeadlineExceeded error")
}

// TestCaptureStderr_Basic tests basic stderr capture functionality
func TestCaptureStderr_Basic(t *testing.T) {
	// Test capturing a simple function that writes to stderr
	output, err := captureStderr(func() error {
		return nil
	})

	require.NoError(t, err, "Expected no error from function")
	assert.Empty(t, output, "Expected empty output when nothing is written to stderr")
}

// TestCaptureStderr_Error tests that captureStderr returns both output and error
func TestCaptureStderr_Error(t *testing.T) {
	expectedErr := assert.AnError

	output, err := captureStderr(func() error {
		return expectedErr
	})

	require.Error(t, err, "Expected error from function")
	assert.Equal(t, expectedErr, err, "Expected the same error to be returned")
	assert.Empty(t, output, "Expected empty output when nothing is written to stderr")
}

// TestUpdateActionsContext_Cancellation tests UpdateActionsContext respects cancellation
func TestUpdateActionsContext_Cancellation(t *testing.T) {
	// Create a context that is already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Call the function with the cancelled context
	err := UpdateActionsContext(ctx, false, false)

	// Verify that the error is context.Canceled
	require.Error(t, err, "Expected an error when context is cancelled")
	assert.Equal(t, context.Canceled, err, "Expected context.Canceled error")
}

// TestUpdateWorkflowsContext_Cancellation tests UpdateWorkflowsContext respects cancellation
func TestUpdateWorkflowsContext_Cancellation(t *testing.T) {
	// Create a context that is already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Call the function with the cancelled context
	err := UpdateWorkflowsContext(
		ctx,
		[]string{}, // workflowNames
		false,      // allowMajor
		false,      // force
		false,      // verbose
		"",         // engineOverride
		"",         // workflowsDir
		false,      // noStopAfter
		"",         // stopAfter
		false,      // merge
	)

	// Verify that the error is context.Canceled
	require.Error(t, err, "Expected an error when context is cancelled")
	assert.Equal(t, context.Canceled, err, "Expected context.Canceled error")
}

// TestRunFixContext_Cancellation tests RunFixContext respects cancellation
func TestRunFixContext_Cancellation(t *testing.T) {
	// Create a context that is already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	config := FixConfig{
		WorkflowIDs: []string{},
		Write:       false,
		Verbose:     false,
		WorkflowDir: "",
	}

	// Call the function with the cancelled context
	err := RunFixContext(ctx, config)

	// Verify that the error is context.Canceled
	require.Error(t, err, "Expected an error when context is cancelled")
	assert.Equal(t, context.Canceled, err, "Expected context.Canceled error")
}

// TestCheckExtensionUpdateContext_Cancellation tests checkExtensionUpdateContext respects cancellation
func TestCheckExtensionUpdateContext_Cancellation(t *testing.T) {
	// Create a context that is already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Call the function with the cancelled context
	err := checkExtensionUpdateContext(ctx, false)

	// Verify that the error is context.Canceled
	require.Error(t, err, "Expected an error when context is cancelled")
	assert.Equal(t, context.Canceled, err, "Expected context.Canceled error")
}
