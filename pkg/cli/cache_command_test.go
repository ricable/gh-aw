//go:build !integration

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCacheCommand(t *testing.T) {
	cmd := NewCacheCommand()

	require.NotNil(t, cmd, "NewCacheCommand should not return nil")
	assert.Equal(t, "cache", cmd.Use, "Command use should be 'cache'")
	assert.Equal(t, "Manage GitHub Actions caches for workflows", cmd.Short, "Command short description should match")
	assert.Contains(t, cmd.Long, "cache-memory", "Command long description should mention cache-memory")
	assert.Contains(t, cmd.Long, "list", "Command long description should mention list subcommand")
	assert.Contains(t, cmd.Long, "download", "Command long description should mention download subcommand")
	assert.Contains(t, cmd.Long, "delete", "Command long description should mention delete subcommand")

	// Verify subcommands are registered
	subcommands := cmd.Commands()
	require.Len(t, subcommands, 3, "Should have exactly 3 subcommands")

	// Check that list, download and delete subcommands exist
	hasList := false
	hasDownload := false
	hasDelete := false
	for _, subcmd := range subcommands {
		switch subcmd.Name() {
		case "list":
			hasList = true
		case "download":
			hasDownload = true
		case "delete":
			hasDelete = true
		}
	}

	assert.True(t, hasList, "Should have 'list' subcommand")
	assert.True(t, hasDownload, "Should have 'download' subcommand")
	assert.True(t, hasDelete, "Should have 'delete' subcommand")
}

func TestNewCacheDownloadCommand(t *testing.T) {
	cmd := NewCacheDownloadCommand()

	require.NotNil(t, cmd, "NewCacheDownloadCommand should not return nil")
	assert.Equal(t, "download [workflow]", cmd.Use, "Command use should be 'download [workflow]'")
	assert.Equal(t, "Download cache artifacts for a workflow", cmd.Short, "Command short description should match")
	assert.Contains(t, cmd.Long, "cache-memory", "Command long description should mention cache-memory")
	assert.Contains(t, cmd.Long, "memory-<workflow>-", "Command long description should mention key pattern")

	// Verify flags are registered
	flags := cmd.Flags()

	// Check output flag
	outputFlag := flags.Lookup("output")
	require.NotNil(t, outputFlag, "Should have 'output' flag")
	assert.Equal(t, "o", outputFlag.Shorthand, "Output flag shorthand should be 'o'")
	assert.Equal(t, "./cache-downloads", outputFlag.DefValue, "Output flag default should be './cache-downloads'")

	// Check limit flag
	limitFlag := flags.Lookup("limit")
	require.NotNil(t, limitFlag, "Should have 'limit' flag")
	assert.Equal(t, "L", limitFlag.Shorthand, "Limit flag shorthand should be 'L'")
	assert.Equal(t, "30", limitFlag.DefValue, "Limit flag default should be '30'")

	// Check key flag
	keyFlag := flags.Lookup("key")
	require.NotNil(t, keyFlag, "Should have 'key' flag")
	assert.Equal(t, "k", keyFlag.Shorthand, "Key flag shorthand should be 'k'")
	assert.Empty(t, keyFlag.DefValue, "Key flag default should be empty")
}

func TestNewCacheDeleteCommand(t *testing.T) {
	cmd := NewCacheDeleteCommand()

	require.NotNil(t, cmd, "NewCacheDeleteCommand should not return nil")
	assert.Equal(t, "delete [workflow]", cmd.Use, "Command use should be 'delete [workflow]'")
	assert.Equal(t, "Delete cache artifacts for a workflow", cmd.Short, "Command short description should match")
	assert.Contains(t, cmd.Long, "cache-memory", "Command long description should mention cache-memory")
	assert.Contains(t, cmd.Long, "confirmation", "Command long description should mention confirmation")

	// Verify flags are registered
	flags := cmd.Flags()

	// Check all flag
	allFlag := flags.Lookup("all")
	require.NotNil(t, allFlag, "Should have 'all' flag")
	assert.Equal(t, "a", allFlag.Shorthand, "All flag shorthand should be 'a'")
	assert.Equal(t, "false", allFlag.DefValue, "All flag default should be 'false'")

	// Check key flag
	keyFlag := flags.Lookup("key")
	require.NotNil(t, keyFlag, "Should have 'key' flag")
	assert.Equal(t, "k", keyFlag.Shorthand, "Key flag shorthand should be 'k'")
	assert.Empty(t, keyFlag.DefValue, "Key flag default should be empty")

	// Check ref flag
	refFlag := flags.Lookup("ref")
	require.NotNil(t, refFlag, "Should have 'ref' flag")
	assert.Equal(t, "r", refFlag.Shorthand, "Ref flag shorthand should be 'r'")
	assert.Empty(t, refFlag.DefValue, "Ref flag default should be empty")

	// Check force flag
	forceFlag := flags.Lookup("force")
	require.NotNil(t, forceFlag, "Should have 'force' flag")
	assert.Equal(t, "f", forceFlag.Shorthand, "Force flag shorthand should be 'f'")
	assert.Equal(t, "false", forceFlag.DefValue, "Force flag default should be 'false'")
}

func TestCacheDownloadConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   CacheDownloadConfig
		expected string
	}{
		{
			name: "basic config",
			config: CacheDownloadConfig{
				WorkflowID: "my-workflow",
				OutputDir:  "./caches",
				Limit:      10,
				Verbose:    false,
			},
			expected: "my-workflow",
		},
		{
			name: "config with .md extension",
			config: CacheDownloadConfig{
				WorkflowID: "my-workflow.md",
				OutputDir:  "./caches",
				Limit:      10,
				Verbose:    true,
			},
			expected: "my-workflow.md",
		},
		{
			name: "config with custom cache key",
			config: CacheDownloadConfig{
				WorkflowID: "test",
				OutputDir:  "/tmp/caches",
				Limit:      5,
				CacheKey:   "memory-custom-",
				Verbose:    false,
			},
			expected: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.config.WorkflowID, "WorkflowID should match")
		})
	}
}

func TestCacheDeleteConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   CacheDeleteConfig
		expected string
	}{
		{
			name: "basic config",
			config: CacheDeleteConfig{
				WorkflowID: "my-workflow",
				DeleteAll:  false,
				Force:      false,
				Verbose:    false,
			},
			expected: "my-workflow",
		},
		{
			name: "config with all flags",
			config: CacheDeleteConfig{
				WorkflowID: "test-workflow",
				DeleteAll:  true,
				CacheKey:   "memory-test-",
				Ref:        "refs/heads/main",
				Force:      true,
				Verbose:    true,
			},
			expected: "test-workflow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.config.WorkflowID, "WorkflowID should match")
		})
	}
}

func TestNewCacheListCommand(t *testing.T) {
	cmd := NewCacheListCommand()

	require.NotNil(t, cmd, "NewCacheListCommand should not return nil")
	assert.Equal(t, "list [workflow]", cmd.Use, "Command use should be 'list [workflow]'")
	assert.Equal(t, "List cache artifacts for a workflow", cmd.Short, "Command short description should match")
	assert.Contains(t, cmd.Long, "cache-memory", "Command long description should mention cache-memory")
	assert.Contains(t, cmd.Long, "memory-<workflow>-", "Command long description should mention key pattern")

	// Verify flags are registered
	flags := cmd.Flags()

	// Check limit flag
	limitFlag := flags.Lookup("limit")
	require.NotNil(t, limitFlag, "Should have 'limit' flag")
	assert.Equal(t, "L", limitFlag.Shorthand, "Limit flag shorthand should be 'L'")
	assert.Equal(t, "30", limitFlag.DefValue, "Limit flag default should be '30'")

	// Check key flag
	keyFlag := flags.Lookup("key")
	require.NotNil(t, keyFlag, "Should have 'key' flag")
	assert.Equal(t, "k", keyFlag.Shorthand, "Key flag shorthand should be 'k'")
	assert.Empty(t, keyFlag.DefValue, "Key flag default should be empty")

	// Check ref flag
	refFlag := flags.Lookup("ref")
	require.NotNil(t, refFlag, "Should have 'ref' flag")
	assert.Equal(t, "r", refFlag.Shorthand, "Ref flag shorthand should be 'r'")
	assert.Empty(t, refFlag.DefValue, "Ref flag default should be empty")
}

func TestCacheListConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   CacheListConfig
		expected string
	}{
		{
			name: "basic config",
			config: CacheListConfig{
				WorkflowID: "my-workflow",
				Limit:      30,
				Verbose:    false,
			},
			expected: "my-workflow",
		},
		{
			name: "config with all flags",
			config: CacheListConfig{
				WorkflowID: "test-workflow",
				Limit:      10,
				CacheKey:   "memory-test-",
				Ref:        "refs/heads/main",
				Verbose:    true,
			},
			expected: "test-workflow",
		},
		{
			name: "config without workflow",
			config: CacheListConfig{
				WorkflowID: "",
				Limit:      50,
				Verbose:    false,
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.config.WorkflowID, "WorkflowID should match")
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"zero bytes", 0, "0 B"},
		{"bytes", 512, "512 B"},
		{"kilobytes", 1536, "1.5 KB"},
		{"megabytes", 1048576, "1.0 MB"},
		{"gigabytes", 2147483648, "2.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBytes(tt.bytes)
			assert.Equal(t, tt.expected, result, "Formatted bytes should match")
		})
	}
}
