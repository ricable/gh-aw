//go:build !integration

package workflow

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckoutManager_NoCustomCheckouts(t *testing.T) {
	mgr := NewCheckoutManager(nil, false, "")
	lines := mgr.GenerateMainCheckoutStep()
	result := strings.Join(lines, "")

	assert.Contains(t, result, "name: Checkout repository", "should have default step name")
	assert.Contains(t, result, "uses: actions/checkout", "should use actions/checkout")
	assert.Contains(t, result, "persist-credentials: false", "should default to persist-credentials false")
	assert.NotContains(t, result, "repository:", "should not have repository without custom config")
	assert.NotContains(t, result, "ref:", "should not have ref without custom config")
}

func TestCheckoutManager_SingleCheckout_NoPath(t *testing.T) {
	// A single checkout without a path overrides the main checkout.
	ref := "my-feature-branch"
	co := CheckoutConfig{Ref: ref}

	mgr := NewCheckoutManager([]CheckoutConfig{co}, false, "")
	lines := mgr.GenerateMainCheckoutStep()
	result := strings.Join(lines, "")

	assert.Contains(t, result, "name: Checkout repository", "step name should still be 'Checkout repository'")
	assert.Contains(t, result, "ref: my-feature-branch", "should include user-specified ref")
	assert.Contains(t, result, "persist-credentials: false", "should keep persist-credentials false")

	// No additional steps because it was consumed as main checkout.
	additional := mgr.GenerateAdditionalCheckoutSteps()
	assert.Empty(t, additional, "no additional checkouts expected")
}

func TestCheckoutManager_SingleCheckout_WithPath(t *testing.T) {
	// A single checkout WITH a path is treated as an additional checkout (not the main one).
	co := CheckoutConfig{
		Repository: "org/repo",
		Ref:        "main",
		Path:       "myrepo",
	}

	mgr := NewCheckoutManager([]CheckoutConfig{co}, false, "")

	mainLines := mgr.GenerateMainCheckoutStep()
	mainResult := strings.Join(mainLines, "")
	// Main checkout should be default (no custom fields except persist-credentials)
	assert.Contains(t, mainResult, "name: Checkout repository", "default main checkout")
	assert.Contains(t, mainResult, "persist-credentials: false", "default persist-credentials")
	assert.NotContains(t, mainResult, "org/repo", "main checkout should not reference additional repo")

	// Additional checkout should have the custom settings.
	additional := mgr.GenerateAdditionalCheckoutSteps()
	addResult := strings.Join(additional, "")
	assert.Contains(t, addResult, "repository: org/repo", "should include repository")
	assert.Contains(t, addResult, "ref: main", "should include ref")
	assert.Contains(t, addResult, "path: myrepo", "should include path")
	assert.Contains(t, addResult, "persist-credentials: false", "should default persist-credentials to false")
}

func TestCheckoutManager_ArrayCheckout_FirstNoPath(t *testing.T) {
	// Array: first entry has no path → becomes main checkout override.
	// Remaining entries are additional checkouts.
	checkouts := []CheckoutConfig{
		{Ref: "my-branch"},
		{Repository: "org/repo2", Path: "repo2"},
	}

	mgr := NewCheckoutManager(checkouts, false, "")

	mainLines := mgr.GenerateMainCheckoutStep()
	mainResult := strings.Join(mainLines, "")
	assert.Contains(t, mainResult, "ref: my-branch", "main checkout should use first entry's ref")
	assert.NotContains(t, mainResult, "org/repo2", "main checkout should not include second entry")

	additional := mgr.GenerateAdditionalCheckoutSteps()
	addResult := strings.Join(additional, "")
	assert.Contains(t, addResult, "repository: org/repo2", "additional checkout should include second entry")
	assert.Contains(t, addResult, "path: repo2", "additional checkout should include path")
}

func TestCheckoutManager_ArrayCheckout_AllWithPath(t *testing.T) {
	// Array: all entries have paths → all are additional checkouts, main is default.
	checkouts := []CheckoutConfig{
		{Repository: "org/repo1", Path: "repo1"},
		{Repository: "org/repo2", Path: "repo2"},
	}

	mgr := NewCheckoutManager(checkouts, false, "")

	mainLines := mgr.GenerateMainCheckoutStep()
	mainResult := strings.Join(mainLines, "")
	assert.Contains(t, mainResult, "name: Checkout repository", "main checkout should be default")
	assert.NotContains(t, mainResult, "org/repo1", "main checkout should not include custom repos")

	additional := mgr.GenerateAdditionalCheckoutSteps()
	addResult := strings.Join(additional, "")
	assert.Contains(t, addResult, "repository: org/repo1", "should include first additional checkout")
	assert.Contains(t, addResult, "path: repo1", "should include first checkout path")
	assert.Contains(t, addResult, "repository: org/repo2", "should include second additional checkout")
	assert.Contains(t, addResult, "path: repo2", "should include second checkout path")
}

func TestCheckoutManager_AdditionalCheckout_AutoPath(t *testing.T) {
	// When an additional checkout has no path, it is auto-derived from the repository slug.
	checkouts := []CheckoutConfig{
		{Path: "main-repo"},       // first has path → becomes additional (not main override)
		{Repository: "org/mylib"}, // second has no path → auto-derived
	}

	mgr := NewCheckoutManager(checkouts, false, "")
	additional := mgr.GenerateAdditionalCheckoutSteps()
	addResult := strings.Join(additional, "")

	// First additional entry
	assert.Contains(t, addResult, "path: main-repo", "should keep explicit path for first entry")
	// Second additional entry: path derived from "org/mylib" → "mylib"
	assert.Contains(t, addResult, "path: mylib", "should auto-derive path from repo slug")
}

func TestCheckoutManager_TrialMode(t *testing.T) {
	mgr := NewCheckoutManager(nil, true, "owner/target-repo")
	lines := mgr.GenerateMainCheckoutStep()
	result := strings.Join(lines, "")

	assert.Contains(t, result, "repository: owner/target-repo", "should include trial logical repo")
	assert.Contains(t, result, "token:", "should include token in trial mode")
}

func TestCheckoutManager_CustomPersistCredentials(t *testing.T) {
	co := CheckoutConfig{PersistCredentials: boolPtr(true)}
	mgr := NewCheckoutManager([]CheckoutConfig{co}, false, "")
	lines := mgr.GenerateMainCheckoutStep()
	result := strings.Join(lines, "")

	assert.Contains(t, result, "persist-credentials: true", "should respect user-specified persist-credentials")
}

func TestCheckoutManager_FetchDepth(t *testing.T) {
	co := CheckoutConfig{FetchDepth: intPtr(0)}
	mgr := NewCheckoutManager([]CheckoutConfig{co}, false, "")
	lines := mgr.GenerateMainCheckoutStep()
	result := strings.Join(lines, "")

	assert.Contains(t, result, "fetch-depth: 0", "should include fetch-depth 0 for full history")
}

func TestCheckoutManager_SparseCheckout(t *testing.T) {
	co := CheckoutConfig{SparseCheckout: "src/\ntest/"}
	mgr := NewCheckoutManager([]CheckoutConfig{co}, false, "")
	lines := mgr.GenerateMainCheckoutStep()
	result := strings.Join(lines, "")

	assert.Contains(t, result, "sparse-checkout: |", "should include sparse-checkout block")
	assert.Contains(t, result, "src/", "should include sparse-checkout patterns")
	assert.Contains(t, result, "test/", "should include all sparse-checkout patterns")
}

func TestParseCheckoutConfig_SingleObject(t *testing.T) {
	input := map[string]any{
		"ref":                 "my-branch",
		"fetch-depth":         float64(1),
		"persist-credentials": false,
	}

	checkouts, err := parseCheckoutConfig(input)
	require.NoError(t, err, "should parse single object without error")
	assert.Len(t, checkouts, 1, "should return 1-element slice for single object")
	assert.Equal(t, "my-branch", checkouts[0].Ref, "should parse ref correctly")
	assert.NotNil(t, checkouts[0].FetchDepth, "should parse fetch-depth")
	assert.Equal(t, 1, *checkouts[0].FetchDepth, "should have fetch-depth 1")
}

func TestParseCheckoutConfig_Array(t *testing.T) {
	input := []any{
		map[string]any{"ref": "branch1"},
		map[string]any{"repository": "org/repo2", "path": "repo2"},
	}

	checkouts, err := parseCheckoutConfig(input)
	require.NoError(t, err, "should parse array without error")
	assert.Len(t, checkouts, 2, "should return 2-element slice for array input")
	assert.Equal(t, "branch1", checkouts[0].Ref)
	assert.Equal(t, "org/repo2", checkouts[1].Repository)
	assert.Equal(t, "repo2", checkouts[1].Path)
}

func TestParseCheckoutConfig_InvalidInput(t *testing.T) {
	_, err := parseCheckoutConfig("not-an-object")
	assert.Error(t, err, "should return error for invalid input type")
}
