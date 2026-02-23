//go:build windows

package cli

// augmentEnvPath is a no-op on Windows; PATH augmentation is not needed.
func augmentEnvPath(env []string) []string {
	return env
}
