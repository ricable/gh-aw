package workflow

import (
	"fmt"
	"strings"
)

// CheckoutManager handles the generation of checkout steps from frontmatter configuration.
// It encapsulates all checkout logic, merging the user-defined checkout configuration
// with automatically generated checkout steps.
type CheckoutManager struct {
	// customCheckouts holds user-defined checkout configurations from frontmatter
	customCheckouts []CheckoutConfig
	// trialMode indicates whether the workflow is running in trial mode
	trialMode bool
	// trialLogicalRepoSlug holds the target repository slug for trial mode
	trialLogicalRepoSlug string
}

// NewCheckoutManager creates a new CheckoutManager with the given configuration.
func NewCheckoutManager(customCheckouts []CheckoutConfig, trialMode bool, trialLogicalRepoSlug string) *CheckoutManager {
	return &CheckoutManager{
		customCheckouts:      customCheckouts,
		trialMode:            trialMode,
		trialLogicalRepoSlug: trialLogicalRepoSlug,
	}
}

// GenerateMainCheckoutStep generates the main repository checkout step YAML lines.
// It merges user-defined checkout configuration with the default checkout behaviour.
//
// When no custom checkouts are specified, it falls back to the default checkout
// (persist-credentials: false, optional trial-mode fields).
//
// When a single custom checkout is specified without a path, it is treated as an
// override for the main repository checkout – its fields are merged on top of the
// defaults.
//
// When an array of custom checkouts is specified, only the first element (if it has
// no explicit path) is used as the main checkout override; remaining entries are
// returned by GenerateAdditionalCheckoutSteps.
func (m *CheckoutManager) GenerateMainCheckoutStep() []string {
	var lines []string
	lines = append(lines, "      - name: Checkout repository\n")
	lines = append(lines, fmt.Sprintf("        uses: %s\n", GetActionPin("actions/checkout")))
	lines = append(lines, "        with:\n")

	// Start from defaults.
	// persist-credentials is always false and cannot be overridden by user config.
	persistCredentials := false

	// Collect overrides from user-specified first checkout (if it has no explicit path,
	// meaning it targets the main repository rather than a secondary checkout).
	var override *CheckoutConfig
	if len(m.customCheckouts) > 0 && m.customCheckouts[0].Path == "" {
		override = &m.customCheckouts[0]
	}

	// Apply overrides from the user config.
	var repository, ref, token, sshKey, filter, sparseCheckout, submodules, gitHubServerURL string
	var fetchDepth *int
	var fetchTags, showProgress, lfs, setSafeDirectory, sparseCheckoutConeMode *bool
	var clean *bool

	if override != nil {
		repository = override.Repository
		ref = override.Ref
		token = override.Token
		sshKey = override.SSHKey
		filter = override.Filter
		sparseCheckout = override.SparseCheckout
		submodules = override.Submodules
		gitHubServerURL = override.GitConfigURL
		fetchDepth = override.FetchDepth
		fetchTags = override.FetchTags
		showProgress = override.ShowProgress
		lfs = override.Lfs
		setSafeDirectory = override.SetSafeDirectory
		sparseCheckoutConeMode = override.SparseCheckoutConeMode
		clean = override.Clean
	}

	// Trial mode overrides: set repository and token when running in trial mode.
	if m.trialMode {
		if repository == "" && m.trialLogicalRepoSlug != "" {
			repository = m.trialLogicalRepoSlug
		}
		if token == "" {
			token = getEffectiveGitHubToken("")
		}
	}

	// Emit required fields.
	if repository != "" {
		lines = append(lines, fmt.Sprintf("          repository: %s\n", repository))
	}
	if ref != "" {
		lines = append(lines, fmt.Sprintf("          ref: %s\n", ref))
	}
	if token != "" {
		lines = append(lines, fmt.Sprintf("          token: %s\n", token))
	}
	if sshKey != "" {
		lines = append(lines, fmt.Sprintf("          ssh-key: %s\n", sshKey))
	}
	lines = append(lines, fmt.Sprintf("          persist-credentials: %v\n", persistCredentials))
	if clean != nil {
		lines = append(lines, fmt.Sprintf("          clean: %v\n", *clean))
	}
	if filter != "" {
		lines = append(lines, fmt.Sprintf("          filter: %s\n", filter))
	}
	if sparseCheckout != "" {
		lines = append(lines, "          sparse-checkout: |\n")
		for _, pattern := range strings.Split(strings.TrimSpace(sparseCheckout), "\n") {
			lines = append(lines, fmt.Sprintf("            %s\n", strings.TrimSpace(pattern)))
		}
	}
	if sparseCheckoutConeMode != nil {
		lines = append(lines, fmt.Sprintf("          sparse-checkout-cone-mode: %v\n", *sparseCheckoutConeMode))
	}
	if fetchDepth != nil {
		lines = append(lines, fmt.Sprintf("          fetch-depth: %d\n", *fetchDepth))
	}
	if fetchTags != nil {
		lines = append(lines, fmt.Sprintf("          fetch-tags: %v\n", *fetchTags))
	}
	if showProgress != nil {
		lines = append(lines, fmt.Sprintf("          show-progress: %v\n", *showProgress))
	}
	if lfs != nil {
		lines = append(lines, fmt.Sprintf("          lfs: %v\n", *lfs))
	}
	if submodules != "" {
		lines = append(lines, fmt.Sprintf("          submodules: %s\n", submodules))
	}
	if setSafeDirectory != nil {
		lines = append(lines, fmt.Sprintf("          set-safe-directory: %v\n", *setSafeDirectory))
	}
	if gitHubServerURL != "" {
		lines = append(lines, fmt.Sprintf("          github-server-url: %s\n", gitHubServerURL))
	}

	return lines
}

// GenerateAdditionalCheckoutSteps generates YAML lines for any extra checkout entries
// beyond the first/main one. Each additional checkout must have an explicit path.
// If a checkout in the array has no path, a path is automatically derived from the
// repository slug (last path segment) to ensure each checkout is in its own subfolder.
func (m *CheckoutManager) GenerateAdditionalCheckoutSteps() []string {
	if len(m.customCheckouts) == 0 {
		return nil
	}

	// Determine which entries are "additional" (not the main checkout).
	// The first entry is additional only when it already has an explicit path
	// (meaning it does NOT override the main checkout).
	startIdx := 0
	if len(m.customCheckouts) > 0 && m.customCheckouts[0].Path == "" {
		// First entry was used as main checkout override – skip it here.
		startIdx = 1
	}

	var lines []string
	for i := startIdx; i < len(m.customCheckouts); i++ {
		co := m.customCheckouts[i]
		checkoutLines := m.generateAdditionalCheckoutStep(co, i)
		lines = append(lines, checkoutLines...)
	}
	return lines
}

// generateAdditionalCheckoutStep generates YAML lines for a single additional checkout.
// The index is used to create a unique default path when no path is specified.
func (m *CheckoutManager) generateAdditionalCheckoutStep(co CheckoutConfig, index int) []string {
	// Derive the step name.
	name := "Checkout"
	if co.Repository != "" {
		name = fmt.Sprintf("Checkout %s", co.Repository)
		if co.Ref != "" {
			name = fmt.Sprintf("Checkout %s@%s", co.Repository, co.Ref)
		}
	}

	// Derive the path if not explicitly set.
	path := co.Path
	if path == "" {
		if co.Repository != "" {
			// Use the repo name (last segment of owner/repo).
			parts := strings.Split(co.Repository, "/")
			path = parts[len(parts)-1]
		} else {
			// Fallback to a numbered path.
			path = fmt.Sprintf("checkout-%d", index)
		}
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("      - name: %s\n", name))
	lines = append(lines, fmt.Sprintf("        uses: %s\n", GetActionPin("actions/checkout")))
	lines = append(lines, "        with:\n")

	if co.Repository != "" {
		lines = append(lines, fmt.Sprintf("          repository: %s\n", co.Repository))
	}
	if co.Ref != "" {
		lines = append(lines, fmt.Sprintf("          ref: %s\n", co.Ref))
	}
	if co.Token != "" {
		lines = append(lines, fmt.Sprintf("          token: %s\n", co.Token))
	}
	if co.SSHKey != "" {
		lines = append(lines, fmt.Sprintf("          ssh-key: %s\n", co.SSHKey))
	}
	lines = append(lines, fmt.Sprintf("          path: %s\n", path))

	// persist-credentials is always false and cannot be overridden.
	lines = append(lines, "          persist-credentials: false\n")

	if co.Clean != nil {
		lines = append(lines, fmt.Sprintf("          clean: %v\n", *co.Clean))
	}
	if co.Filter != "" {
		lines = append(lines, fmt.Sprintf("          filter: %s\n", co.Filter))
	}
	if co.SparseCheckout != "" {
		lines = append(lines, "          sparse-checkout: |\n")
		for _, pattern := range strings.Split(strings.TrimSpace(co.SparseCheckout), "\n") {
			lines = append(lines, fmt.Sprintf("            %s\n", strings.TrimSpace(pattern)))
		}
	}
	if co.SparseCheckoutConeMode != nil {
		lines = append(lines, fmt.Sprintf("          sparse-checkout-cone-mode: %v\n", *co.SparseCheckoutConeMode))
	}
	if co.FetchDepth != nil {
		lines = append(lines, fmt.Sprintf("          fetch-depth: %d\n", *co.FetchDepth))
	}
	if co.FetchTags != nil {
		lines = append(lines, fmt.Sprintf("          fetch-tags: %v\n", *co.FetchTags))
	}
	if co.ShowProgress != nil {
		lines = append(lines, fmt.Sprintf("          show-progress: %v\n", *co.ShowProgress))
	}
	if co.Lfs != nil {
		lines = append(lines, fmt.Sprintf("          lfs: %v\n", *co.Lfs))
	}
	if co.Submodules != "" {
		lines = append(lines, fmt.Sprintf("          submodules: %s\n", co.Submodules))
	}
	if co.SetSafeDirectory != nil {
		lines = append(lines, fmt.Sprintf("          set-safe-directory: %v\n", *co.SetSafeDirectory))
	}
	if co.GitConfigURL != "" {
		lines = append(lines, fmt.Sprintf("          github-server-url: %s\n", co.GitConfigURL))
	}

	return lines
}
