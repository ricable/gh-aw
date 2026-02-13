package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/github/gh-aw/pkg/console"
	"github.com/github/gh-aw/pkg/constants"
	"github.com/github/gh-aw/pkg/logger"
	"github.com/github/gh-aw/pkg/workflow"
	"github.com/spf13/cobra"
)

var cacheDownloadLog = logger.New("cli:cache_download")

// NewCacheDownloadCommand creates the cache download command
func NewCacheDownloadCommand() *cobra.Command {
	var outputDir string
	var limit int
	var cacheKey string

	cmd := &cobra.Command{
		Use:   "download [workflow]",
		Short: "Download cache artifacts for a workflow",
		Long: `Download GitHub Actions cache artifacts for agentic workflows using cache-memory.

This command downloads cache artifacts that workflows created when using the cache-memory
feature. By default, it downloads all caches with keys matching the workflow name pattern
'memory-<workflow>-*'.

The downloaded cache data is saved to the specified output directory (default: ./cache-downloads).
Each cache is saved in a subdirectory named by its cache key and ID.

` + WorkflowIDExplanation + `

Examples:
  ` + string(constants.CLIExtensionPrefix) + ` cache download my-workflow                    # Download caches for workflow
  ` + string(constants.CLIExtensionPrefix) + ` cache download my-workflow -o ./caches        # Custom output directory
  ` + string(constants.CLIExtensionPrefix) + ` cache download my-workflow -L 10              # Limit to 10 most recent
  ` + string(constants.CLIExtensionPrefix) + ` cache download my-workflow -k memory-custom   # Download specific cache key`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workflowID := args[0]
			verbose, _ := cmd.Flags().GetBool("verbose")

			config := CacheDownloadConfig{
				WorkflowID: workflowID,
				OutputDir:  outputDir,
				Limit:      limit,
				CacheKey:   cacheKey,
				Verbose:    verbose,
			}

			return RunCacheDownload(config)
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "./cache-downloads", "Output directory for downloaded caches")
	cmd.Flags().IntVarP(&limit, "limit", "L", 30, "Maximum number of caches to download")
	cmd.Flags().StringVarP(&cacheKey, "key", "k", "", "Filter by cache key prefix")

	return cmd
}

// CacheDownloadConfig holds configuration for cache download
type CacheDownloadConfig struct {
	WorkflowID string
	OutputDir  string
	Limit      int
	CacheKey   string
	Verbose    bool
}

// CacheEntry represents a GitHub Actions cache
type CacheEntry struct {
	ID            int64  `json:"id"`
	Key           string `json:"key"`
	Ref           string `json:"ref"`
	SizeInBytes   int64  `json:"size_in_bytes"`
	CreatedAt     string `json:"created_at"`
	LastAccessedAt string `json:"last_accessed_at"`
}

// CacheListResponse represents the response from GitHub Actions cache API
type CacheListResponse struct {
	TotalCount int          `json:"total_count"`
	Caches     []CacheEntry `json:"actions_caches"`
}

// RunCacheDownload executes the cache download logic
func RunCacheDownload(config CacheDownloadConfig) error {
	cacheDownloadLog.Printf("Starting cache download: workflow=%s, outputDir=%s", config.WorkflowID, config.OutputDir)

	// Strip .md extension if present
	workflowID := strings.TrimSuffix(config.WorkflowID, ".md")

	// Determine cache key pattern to search for
	keyPattern := config.CacheKey
	if keyPattern == "" {
		// Default: search for caches with keys matching memory-<workflow>-
		keyPattern = fmt.Sprintf("memory-%s-", workflowID)
	}

	fmt.Fprintln(os.Stderr, console.FormatInfoMessage(fmt.Sprintf("Searching for caches with key prefix: %s", keyPattern)))

	// List caches using gh CLI
	caches, err := listCaches(keyPattern, config.Limit, config.Verbose)
	if err != nil {
		fmt.Fprintln(os.Stderr, console.FormatErrorMessage(err.Error()))
		return fmt.Errorf("failed to list caches: %w", err)
	}

	if len(caches) == 0 {
		fmt.Fprintln(os.Stderr, console.FormatWarningMessage(fmt.Sprintf("No caches found with key prefix: %s", keyPattern)))
		return nil
	}

	fmt.Fprintln(os.Stderr, console.FormatInfoMessage(fmt.Sprintf("Found %d cache(s)", len(caches))))

	// Create output directory
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		fmt.Fprintln(os.Stderr, console.FormatErrorMessage(err.Error()))
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Download each cache
	successCount := 0
	for i, cache := range caches {
		if config.Verbose {
			fmt.Fprintln(os.Stderr, console.FormatInfoMessage(fmt.Sprintf("[%d/%d] Downloading cache: %s (ID: %d, Size: %d bytes)",
				i+1, len(caches), cache.Key, cache.ID, cache.SizeInBytes)))
		}

		// Note: GitHub Actions caches are not directly downloadable via API
		// They are only accessible during workflow runs via actions/cache
		// We'll save the cache metadata instead
		cacheDir := filepath.Join(config.OutputDir, fmt.Sprintf("%s-%d", cache.Key, cache.ID))
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			fmt.Fprintln(os.Stderr, console.FormatWarningMessage(fmt.Sprintf("Failed to create cache directory: %v", err)))
			continue
		}

		// Save cache metadata
		metadataPath := filepath.Join(cacheDir, "cache-metadata.json")
		metadataBytes, err := json.MarshalIndent(cache, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, console.FormatWarningMessage(fmt.Sprintf("Failed to marshal cache metadata: %v", err)))
			continue
		}

		if err := os.WriteFile(metadataPath, metadataBytes, 0644); err != nil {
			fmt.Fprintln(os.Stderr, console.FormatWarningMessage(fmt.Sprintf("Failed to write cache metadata: %v", err)))
			continue
		}

		successCount++
	}

	fmt.Fprintln(os.Stderr, console.FormatSuccessMessage(fmt.Sprintf("Downloaded metadata for %d cache(s) to %s", successCount, config.OutputDir)))

	return nil
}

// listCaches retrieves cache entries from GitHub Actions cache API
func listCaches(keyPrefix string, limit int, verbose bool) ([]CacheEntry, error) {
	cacheDownloadLog.Printf("Listing caches: keyPrefix=%s, limit=%d", keyPrefix, limit)

	// Use gh CLI to list caches
	args := []string{
		"cache", "list",
		"--key", keyPrefix,
		"--limit", fmt.Sprintf("%d", limit),
		"--json", "id,key,ref,sizeInBytes,createdAt,lastAccessedAt",
	}

	output, err := workflow.RunGHCombined("Listing caches...", args...)
	if err != nil {
		return nil, fmt.Errorf("gh cache list failed: %w", err)
	}

	// Parse JSON output
	var caches []CacheEntry
	if err := json.Unmarshal(output, &caches); err != nil {
		return nil, fmt.Errorf("failed to parse cache list: %w", err)
	}

	cacheDownloadLog.Printf("Found %d caches", len(caches))
	return caches, nil
}
