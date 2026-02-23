//go:build !integration && !js && !wasm

package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateNpxPackages_SkipEnvVar(t *testing.T) {
	// When GH_AW_SKIP_NPX_VALIDATION=true the function must return nil regardless of
	// whether npm is available and regardless of which packages are referenced.
	t.Setenv("GH_AW_SKIP_NPX_VALIDATION", "true")

	compiler := NewCompiler()

	// Workflow that references an npx package – would normally trigger npm validation.
	workflowData := &WorkflowData{
		CustomSteps: "npx @mcp/inspector --some-flag",
	}

	err := compiler.validateNpxPackages(workflowData)
	assert.NoError(t, err, "validateNpxPackages should return nil when GH_AW_SKIP_NPX_VALIDATION=true")
}

func TestValidateNpxPackages_EmptyPackageList(t *testing.T) {
	// When there are no npx packages the function must return nil without needing npm.
	compiler := NewCompiler()

	workflowData := &WorkflowData{
		CustomSteps: "echo hello",
	}

	err := compiler.validateNpxPackages(workflowData)
	require.NoError(t, err, "validateNpxPackages should return nil when no npx packages are referenced")
}
