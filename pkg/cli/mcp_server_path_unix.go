//go:build !windows

package cli

import (
	"os"
	"strings"
)

// augmentEnvPath returns env (defaulting to os.Environ() when nil) with
// /usr/local/bin appended to PATH if not already present, so that tools like
// npm are reachable even when the MCP server was started with a restricted PATH.
func augmentEnvPath(env []string) []string {
	if env == nil {
		env = os.Environ()
	}
	const dir = "/usr/local/bin"
	for i, e := range env {
		if suffix, ok := strings.CutPrefix(e, "PATH="); ok {
			if !strings.Contains(":"+suffix+":", ":"+dir+":") {
				env[i] = "PATH=" + suffix + ":" + dir
			}
			return env
		}
	}
	// No PATH entry found; create one with the common dir.
	env = append(env, "PATH="+dir)
	return env
}
