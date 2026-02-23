//go:build !integration && !windows

package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAugmentEnvPath(t *testing.T) {
	tests := []struct {
		name      string
		env       []string
		wantPaths []string // directories that must appear in the resulting PATH
	}{
		{
			name:      "nil env gets /usr/local/bin appended",
			env:       nil,
			wantPaths: []string{"/usr/local/bin"},
		},
		{
			name:      "existing PATH gets /usr/local/bin appended",
			env:       []string{"PATH=/usr/bin:/bin", "HOME=/root"},
			wantPaths: []string{"/usr/bin", "/bin", "/usr/local/bin"},
		},
		{
			name:      "PATH already containing /usr/local/bin is not duplicated",
			env:       []string{"PATH=/usr/local/bin:/usr/bin"},
			wantPaths: []string{"/usr/local/bin", "/usr/bin"},
		},
		{
			name:      "env without PATH entry gets PATH added",
			env:       []string{"HOME=/root", "USER=runner"},
			wantPaths: []string{"/usr/local/bin"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := augmentEnvPath(tt.env)
			require.NotEmpty(t, result, "augmentEnvPath should return non-empty env")

			// Find the PATH entry
			var pathValue string
			for _, e := range result {
				if v, ok := strings.CutPrefix(e, "PATH="); ok {
					pathValue = v
					break
				}
			}
			require.NotEmpty(t, pathValue, "PATH entry must be present in result")

			// Verify all expected directories are present
			pathDirs := strings.Split(pathValue, ":")
			pathSet := make(map[string]bool, len(pathDirs))
			for _, d := range pathDirs {
				pathSet[d] = true
			}
			for _, want := range tt.wantPaths {
				assert.True(t, pathSet[want], "PATH should contain %q, got: %s", want, pathValue)
			}
		})
	}
}

func TestAugmentEnvPath_NoDuplicates(t *testing.T) {
	// Verify that calling augmentEnvPath twice does not duplicate directories.
	env := []string{"PATH=/usr/bin"}
	once := augmentEnvPath(env)
	twice := augmentEnvPath(append([]string(nil), once...))

	var pathOnce, pathTwice string
	for _, e := range once {
		if v, ok := strings.CutPrefix(e, "PATH="); ok {
			pathOnce = v
		}
	}
	for _, e := range twice {
		if v, ok := strings.CutPrefix(e, "PATH="); ok {
			pathTwice = v
		}
	}

	assert.Equal(t, pathOnce, pathTwice, "calling augmentEnvPath twice should not duplicate PATH entries")
}
