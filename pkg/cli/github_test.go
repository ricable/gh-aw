//go:build !integration

package cli

import (
	"testing"
)

func TestGetGitHubHost(t *testing.T) {
	tests := []struct {
		name         string
		serverURL    string
		ghHost       string
		apiURL       string
		expectedHost string
	}{
		{
			name:         "defaults to github.com",
			serverURL:    "",
			ghHost:       "",
			apiURL:       "",
			expectedHost: "https://github.com",
		},
		{
			name:         "uses GITHUB_SERVER_URL when set",
			serverURL:    "https://github.enterprise.com",
			ghHost:       "",
			apiURL:       "",
			expectedHost: "https://github.enterprise.com",
		},
		{
			name:         "uses GH_HOST when GITHUB_SERVER_URL not set",
			serverURL:    "",
			ghHost:       "https://github.company.com",
			apiURL:       "",
			expectedHost: "https://github.company.com",
		},
		{
			name:         "adds https:// prefix to GH_HOST if missing",
			serverURL:    "",
			ghHost:       "github.company.com",
			apiURL:       "",
			expectedHost: "https://github.company.com",
		},
		{
			name:         "GITHUB_SERVER_URL takes precedence over GH_HOST",
			serverURL:    "https://github.enterprise.com",
			ghHost:       "https://github.company.com",
			apiURL:       "",
			expectedHost: "https://github.enterprise.com",
		},
		{
			name:         "removes trailing slash from GITHUB_SERVER_URL",
			serverURL:    "https://github.enterprise.com/",
			ghHost:       "",
			apiURL:       "",
			expectedHost: "https://github.enterprise.com",
		},
		{
			name:         "removes trailing slash from GH_HOST",
			serverURL:    "",
			ghHost:       "https://github.company.com/",
			apiURL:       "",
			expectedHost: "https://github.company.com",
		},
		{
			name:         "derives from GITHUB_API_URL when others not set (api subdomain)",
			serverURL:    "",
			ghHost:       "",
			apiURL:       "https://api.github.com",
			expectedHost: "https://github.com",
		},
		{
			name:         "derives from GITHUB_API_URL (enterprise api subdomain)",
			serverURL:    "",
			ghHost:       "",
			apiURL:       "https://api.github.enterprise.com",
			expectedHost: "https://github.enterprise.com",
		},
		{
			name:         "derives from GITHUB_API_URL (path-based API)",
			serverURL:    "",
			ghHost:       "",
			apiURL:       "https://github.enterprise.com/api/v3",
			expectedHost: "https://github.enterprise.com",
		},
		{
			name:         "GITHUB_SERVER_URL takes precedence over GITHUB_API_URL",
			serverURL:    "https://github.primary.com",
			ghHost:       "",
			apiURL:       "https://api.github.secondary.com",
			expectedHost: "https://github.primary.com",
		},
		{
			name:         "GH_HOST takes precedence over GITHUB_API_URL",
			serverURL:    "",
			ghHost:       "github.primary.com",
			apiURL:       "https://api.github.secondary.com",
			expectedHost: "https://github.primary.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set test env vars (always set to ensure clean state)
			t.Setenv("GITHUB_SERVER_URL", tt.serverURL)
			t.Setenv("GH_HOST", tt.ghHost)
			t.Setenv("GITHUB_API_URL", tt.apiURL)

			// Test
			host := getGitHubHost()
			if host != tt.expectedHost {
				t.Errorf("Expected host '%s', got '%s'", tt.expectedHost, host)
			}
		})
	}
}
