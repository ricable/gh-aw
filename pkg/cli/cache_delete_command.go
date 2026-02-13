package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/github/gh-aw/pkg/console"
	"github.com/github/gh-aw/pkg/constants"
	"github.com/github/gh-aw/pkg/logger"
	"github.com/github/gh-aw/pkg/workflow"
	"github.com/spf13/cobra"
)

var cacheDeleteLog = logger.New("cli:cache_delete")

// NewCacheDeleteCommand creates the cache delete command
func NewCacheDeleteCommand() *cobra.Command {
	var deleteAll bool
	var cacheKey string
	var ref string
	var force bool

	cmd := &cobra.Command{
		Use:   "delete [workflow]",
		Short: "Delete cache artifacts for a workflow",
		Long: `Delete GitHub Actions cache artifacts for agentic workflows using cache-memory.

This command deletes cache artifacts that workflows created when using the cache-memory
feature. By default, it prompts for confirmation before deleting caches. Use --force to
skip the confirmation prompt.

Without --all, this command deletes a single cache matching the workflow name pattern.
With --all, it deletes all caches matching the pattern.

` + WorkflowIDExplanation + `

Examples:
  ` + string(constants.CLIExtensionPrefix) + ` cache delete my-workflow                      # Delete first matching cache (with confirmation)
  ` + string(constants.CLIExtensionPrefix) + ` cache delete my-workflow --all                # Delete all matching caches
  ` + string(constants.CLIExtensionPrefix) + ` cache delete my-workflow --all --force        # Delete all without confirmation
  ` + string(constants.CLIExtensionPrefix) + ` cache delete my-workflow -k memory-custom     # Delete specific cache key
  ` + string(constants.CLIExtensionPrefix) + ` cache delete my-workflow -r refs/heads/main   # Delete caches for specific ref`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workflowID := args[0]
			verbose, _ := cmd.Flags().GetBool("verbose")

			config := CacheDeleteConfig{
				WorkflowID: workflowID,
				DeleteAll:  deleteAll,
				CacheKey:   cacheKey,
				Ref:        ref,
				Force:      force,
				Verbose:    verbose,
			}

			return RunCacheDelete(config)
		},
	}

	cmd.Flags().BoolVarP(&deleteAll, "all", "a", false, "Delete all caches matching the pattern")
	cmd.Flags().StringVarP(&cacheKey, "key", "k", "", "Filter by cache key prefix")
	cmd.Flags().StringVarP(&ref, "ref", "r", "", "Filter by ref (e.g., refs/heads/main)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}

// CacheDeleteConfig holds configuration for cache deletion
type CacheDeleteConfig struct {
	WorkflowID string
	DeleteAll  bool
	CacheKey   string
	Ref        string
	Force      bool
	Verbose    bool
}

// RunCacheDelete executes the cache deletion logic
func RunCacheDelete(config CacheDeleteConfig) error {
	cacheDeleteLog.Printf("Starting cache delete: workflow=%s, deleteAll=%t", config.WorkflowID, config.DeleteAll)

	// Strip .md extension if present
	workflowID := strings.TrimSuffix(config.WorkflowID, ".md")

	// Determine cache key pattern to search for
	keyPattern := config.CacheKey
	if keyPattern == "" {
		// Default: search for caches with keys matching memory-<workflow>-
		keyPattern = fmt.Sprintf("memory-%s-", workflowID)
	}

	fmt.Fprintln(os.Stderr, console.FormatInfoMessage(fmt.Sprintf("Searching for caches with key prefix: %s", keyPattern)))

	// List caches to see what we're about to delete
	caches, err := listCaches(keyPattern, 100, config.Verbose)
	if err != nil {
		fmt.Fprintln(os.Stderr, console.FormatErrorMessage(err.Error()))
		return fmt.Errorf("failed to list caches: %w", err)
	}

	if len(caches) == 0 {
		fmt.Fprintln(os.Stderr, console.FormatWarningMessage(fmt.Sprintf("No caches found with key prefix: %s", keyPattern)))
		return nil
	}

	// Determine which caches to delete
	cachesToDelete := caches
	if !config.DeleteAll {
		// Only delete the first cache
		cachesToDelete = caches[:1]
	}

	// Confirmation prompt (unless --force is used)
	if !config.Force {
		fmt.Fprintln(os.Stderr, console.FormatWarningMessage(fmt.Sprintf("About to delete %d cache(s):", len(cachesToDelete))))
		for _, cache := range cachesToDelete {
			fmt.Fprintf(os.Stderr, "  - %s (ID: %d, Size: %d bytes)\n", cache.Key, cache.ID, cache.SizeInBytes)
		}
		fmt.Fprintf(os.Stderr, "\n%s", console.FormatPromptMessage("Proceed with deletion? [y/N]: "))

		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Fprintln(os.Stderr, console.FormatInfoMessage("Deletion cancelled"))
			return nil
		}
	}

	// Delete caches
	successCount := 0
	for i, cache := range cachesToDelete {
		if config.Verbose {
			fmt.Fprintln(os.Stderr, console.FormatInfoMessage(fmt.Sprintf("[%d/%d] Deleting cache: %s (ID: %d)",
				i+1, len(cachesToDelete), cache.Key, cache.ID)))
		}

		if err := deleteCache(cache.ID, config.Ref, config.Verbose); err != nil {
			fmt.Fprintln(os.Stderr, console.FormatWarningMessage(fmt.Sprintf("Failed to delete cache %d: %v", cache.ID, err)))
			continue
		}

		successCount++
	}

	if successCount > 0 {
		fmt.Fprintln(os.Stderr, console.FormatSuccessMessage(fmt.Sprintf("Successfully deleted %d cache(s)", successCount)))
	} else {
		fmt.Fprintln(os.Stderr, console.FormatWarningMessage("No caches were deleted"))
	}

	return nil
}

// deleteCache deletes a single cache by ID
func deleteCache(cacheID int64, ref string, verbose bool) error {
	cacheDeleteLog.Printf("Deleting cache: cacheID=%d, ref=%s", cacheID, ref)

	args := []string{
		"cache", "delete",
		fmt.Sprintf("%d", cacheID),
	}

	if ref != "" {
		args = append(args, "--ref", ref)
	}

	_, err := workflow.RunGHCombined("Deleting cache...", args...)
	if err != nil {
		return fmt.Errorf("gh cache delete failed: %w", err)
	}

	cacheDeleteLog.Printf("Successfully deleted cache: cacheID=%d", cacheID)
	return nil
}
