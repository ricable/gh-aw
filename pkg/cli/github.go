package cli

import (
	"os"
	"strings"

	"github.com/github/gh-aw/pkg/logger"
)

var githubLog = logger.New("cli:github")

// getGitHubHost returns the GitHub host URL from environment variables.
// It checks GITHUB_SERVER_URL first (GitHub Actions standard),
// then falls back to GH_HOST (gh CLI standard),
// then derives from GITHUB_API_URL if available,
// and finally defaults to https://github.com
func getGitHubHost() string {
	host := os.Getenv("GITHUB_SERVER_URL")
	if host == "" {
		host = os.Getenv("GH_HOST")
	}
	if host == "" {
		// Try to derive from GITHUB_API_URL
		if apiURL := os.Getenv("GITHUB_API_URL"); apiURL != "" {
			// Convert API URL to server URL
			// https://api.github.com -> https://github.com
			// https://github.enterprise.com/api/v3 -> https://github.enterprise.com
			host = strings.Replace(apiURL, "://api.", "://", 1)
			host = strings.TrimSuffix(host, "/api/v3")
			host = strings.TrimSuffix(host, "/api")
			githubLog.Printf("Derived GitHub host from GITHUB_API_URL: %s", host)
		}
	}
	if host == "" {
		host = "https://github.com"
		githubLog.Print("Using default GitHub host: https://github.com")
	} else {
		githubLog.Printf("Resolved GitHub host: %s", host)
	}

	// Ensure https:// prefix if only hostname is provided (from GH_HOST)
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "https://" + host
	}

	// Remove trailing slash for consistency
	return strings.TrimSuffix(host, "/")
}
