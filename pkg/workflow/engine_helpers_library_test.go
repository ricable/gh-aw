//go:build !integration

package workflow

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetCommonBinaryPaths verifies that the common binary paths list contains expected binaries
func TestGetCommonBinaryPaths(t *testing.T) {
	binaries := GetCommonBinaryPaths()

	// Should return non-empty list
	assert.NotEmpty(t, binaries, "GetCommonBinaryPaths should return a non-empty list")

	// Should include common utilities
	expectedBinaries := []string{
		"/usr/bin/curl",
		"/usr/bin/jq",
		"/usr/bin/cat",
		"/usr/bin/grep",
	}

	for _, expected := range expectedBinaries {
		assert.Contains(t, binaries, expected, "Should contain common binary: %s", expected)
	}

	// All paths should start with /usr/bin
	for _, binary := range binaries {
		assert.True(t, strings.HasPrefix(binary, "/usr/bin"), "Binary path should start with /usr/bin: %s", binary)
	}

	// Should not contain duplicates
	uniqueBinaries := make(map[string]bool)
	for _, binary := range binaries {
		assert.False(t, uniqueBinaries[binary], "Should not contain duplicates: %s", binary)
		uniqueBinaries[binary] = true
	}
}

// TestGenerateLibraryMountArgsCommand verifies library mount command generation
func TestGenerateLibraryMountArgsCommand(t *testing.T) {
	tests := []struct {
		name      string
		binaries  []string
		expectCmd string
	}{
		{
			name:      "empty binaries list",
			binaries:  []string{},
			expectCmd: "echo ''",
		},
		{
			name:      "single binary",
			binaries:  []string{"/usr/bin/curl"},
			expectCmd: "/opt/gh-aw/scripts/detect-library-deps.sh --cache-file=/tmp/gh-aw-lib-deps-cache.txt --format=awf-mounts /usr/bin/curl 2>/dev/null || echo ''",
		},
		{
			name:      "multiple binaries",
			binaries:  []string{"/usr/bin/curl", "/usr/bin/jq"},
			expectCmd: "/opt/gh-aw/scripts/detect-library-deps.sh --cache-file=/tmp/gh-aw-lib-deps-cache.txt --format=awf-mounts /usr/bin/curl /usr/bin/jq 2>/dev/null || echo ''",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateLibraryMountArgsCommand(tt.binaries)

			assert.Equal(t, tt.expectCmd, result, "Generated command should match expected")

			// Should always include cache file path
			if len(tt.binaries) > 0 {
				assert.Contains(t, result, "/tmp/gh-aw-lib-deps-cache.txt", "Should include cache file path")
			}

			// Should always use awf-mounts format
			if len(tt.binaries) > 0 {
				assert.Contains(t, result, "--format=awf-mounts", "Should use awf-mounts format")
			}

			// Should include error handling
			if len(tt.binaries) > 0 {
				assert.Contains(t, result, "2>/dev/null || echo ''", "Should include error handling")
			}
		})
	}
}

// TestGenerateLibraryMountArgsCommand_EscapingHandling tests that binary paths with special characters are handled correctly
func TestGenerateLibraryMountArgsCommand_EscapingHandling(t *testing.T) {
	// Test with a binary path that might need escaping (though in practice these paths are well-formed)
	binaries := []string{"/usr/bin/test-binary"}
	result := GenerateLibraryMountArgsCommand(binaries)

	// Should contain the binary path
	assert.Contains(t, result, "test-binary", "Should include the binary name")

	// Should not crash or produce empty result
	assert.NotEmpty(t, result, "Should produce non-empty result")
}

// TestGetCommonBinaryPaths_Consistency verifies consistent output
func TestGetCommonBinaryPaths_Consistency(t *testing.T) {
	// Call multiple times to ensure consistent output
	first := GetCommonBinaryPaths()
	second := GetCommonBinaryPaths()

	assert.Equal(t, first, second, "GetCommonBinaryPaths should return consistent results")
}

// TestGenerateLibraryMountArgsCommand_Consistency verifies consistent command generation
func TestGenerateLibraryMountArgsCommand_Consistency(t *testing.T) {
	binaries := []string{"/usr/bin/curl", "/usr/bin/jq"}

	// Call multiple times to ensure consistent output
	first := GenerateLibraryMountArgsCommand(binaries)
	second := GenerateLibraryMountArgsCommand(binaries)

	assert.Equal(t, first, second, "GenerateLibraryMountArgsCommand should return consistent results")
}

// TestGenerateLibraryMountArgsCommand_ScriptPath verifies correct script path
func TestGenerateLibraryMountArgsCommand_ScriptPath(t *testing.T) {
	binaries := []string{"/usr/bin/curl"}
	result := GenerateLibraryMountArgsCommand(binaries)

	// Should reference the installed script path
	assert.Contains(t, result, "/opt/gh-aw/scripts/detect-library-deps.sh", "Should reference installed script path")
}
