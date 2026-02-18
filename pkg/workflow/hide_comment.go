package workflow

import (
	"github.com/github/gh-aw/pkg/logger"
)

var hideCommentLog = logger.New("workflow:hide_comment")

// HideCommentConfig holds configuration for hiding comments from agent output
type HideCommentConfig struct {
	BaseSafeOutputConfig   `yaml:",inline"`
	SafeOutputTargetConfig `yaml:",inline"`
	Discussion             *bool    `yaml:"discussion,omitempty"` // Enable discussion comment support (default: true). Set to false to disable.
	AllowedReasons         []string `yaml:"allowed-reasons,omitempty"` // List of allowed reasons for hiding comments (default: all reasons allowed)
}

// parseHideCommentConfig handles hide-comment configuration
func (c *Compiler) parseHideCommentConfig(outputMap map[string]any) *HideCommentConfig {
	// Check if the key exists
	if _, exists := outputMap["hide-comment"]; !exists {
		return nil
	}

	hideCommentLog.Print("Parsing hide-comment configuration")

	// Unmarshal into typed config struct
	var config HideCommentConfig
	if err := unmarshalConfig(outputMap, "hide-comment", &config, hideCommentLog); err != nil {
		hideCommentLog.Printf("Failed to unmarshal config: %v", err)
		// For backward compatibility, handle nil/empty config
		config = HideCommentConfig{}
	}

	// Set default max if not specified
	if config.Max == 0 {
		config.Max = 5
	}

	// Validate target-repo (wildcard "*" is not allowed)
	if validateTargetRepoSlug(config.TargetRepoSlug, hideCommentLog) {
		return nil // Invalid configuration, return nil to cause validation error
	}

	return &config
}
