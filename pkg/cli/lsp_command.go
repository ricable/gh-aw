package cli

import (
	"fmt"
	"os"

	"github.com/github/gh-aw/pkg/console"
	"github.com/github/gh-aw/pkg/logger"
	"github.com/github/gh-aw/pkg/lsp"
	"github.com/spf13/cobra"
)

var lspLog = logger.New("cli:lsp")

// NewLSPCommand creates the lsp command
func NewLSPCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lsp",
		Short: "Start the Language Server Protocol server for workflow files",
		Long: `Start a Language Server Protocol (LSP) server that provides IDE support
for agentic workflow Markdown files.

The LSP server communicates over stdio using JSON-RPC 2.0 and provides:
- Diagnostics (YAML syntax errors and schema validation)
- Hover information (schema descriptions for frontmatter keys)
- Completions (frontmatter keys, values, and workflow snippets)

The server is stateless (session-only memory, no daemon, no disk state).

Examples:
  gh aw lsp                           # Start LSP server on stdio
  echo '...' | gh aw lsp              # Pipe LSP messages`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunLSP()
		},
	}

	return cmd
}

// RunLSP starts the LSP server on stdio.
func RunLSP() error {
	lspLog.Print("Starting LSP server on stdio")

	server, err := lsp.NewServer(os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, console.FormatErrorMessage(fmt.Sprintf("failed to start LSP server: %s", err.Error())))
		return fmt.Errorf("failed to start LSP server: %w", err)
	}

	if err := server.Run(); err != nil {
		fmt.Fprintln(os.Stderr, console.FormatErrorMessage(fmt.Sprintf("LSP server error: %s", err.Error())))
		return fmt.Errorf("LSP server error: %w", err)
	}

	return nil
}
