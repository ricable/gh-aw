package cli

import (
	"github.com/github/gh-aw/pkg/logger"
	"github.com/spf13/cobra"
)

var cacheCommandLog = logger.New("cli:cache")

// NewCacheCommand creates the main cache command with subcommands
func NewCacheCommand() *cobra.Command {
	cacheCommandLog.Print("Creating cache command with subcommands")
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage GitHub Actions caches for workflows",
		Long: `Manage GitHub Actions caches used by agentic workflows with cache-memory.

When workflows use cache-memory, they store data in GitHub Actions caches with keys
like 'memory-<workflow>-<run-id>'. This command helps manage these caches by listing,
downloading, and deleting them.

Available subcommands:
  • download   - Download cache artifacts for a workflow
  • delete     - Delete cache artifacts for a workflow

Examples:
  gh aw cache download my-workflow           # Download caches for a workflow
  gh aw cache delete my-workflow             # Delete caches for a workflow
  gh aw cache delete my-workflow --all       # Delete all caches for a workflow`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Add subcommands
	cmd.AddCommand(NewCacheDownloadCommand())
	cmd.AddCommand(NewCacheDeleteCommand())

	return cmd
}
