package workflow

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"time"
)

// computeCreatorPromptHash computes a SHA-256 hash of the creator's prompt content
func computeCreatorPromptHash(prompt string) string {
	hash := sha256.Sum256([]byte(prompt))
	// Return first 16 characters of hex hash for brevity
	return hex.EncodeToString(hash[:])[:16]
}

// PromptVersion represents a version for a system prompt file.
// Versions use YYYY-MM-DD format for simplicity and clarity.
type PromptVersion string

// String returns the string representation of the prompt version
func (v PromptVersion) String() string {
	return string(v)
}

// IsValid returns true if the version is non-empty
func (v PromptVersion) IsValid() bool {
	return v != ""
}

// System prompt versions
// Each constant represents the version of a system prompt file in actions/setup/md/
// Versions are date-based (YYYY-MM-DD) to indicate when the prompt was last substantially modified
const (
	// XPIAPromptVersion is the version of the XPIA security policy prompt
	// File: actions/setup/md/xpia.md
	// Last modified: Initial version
	XPIAPromptVersion PromptVersion = "2026-02-17"

	// TempFolderPromptVersion is the version of the temporary folder instructions prompt
	// File: actions/setup/md/temp_folder_prompt.md
	// Last modified: Initial version
	TempFolderPromptVersion PromptVersion = "2026-02-17"

	// MarkdownPromptVersion is the version of the markdown generation instructions prompt
	// File: actions/setup/md/markdown.md
	// Last modified: Initial version
	MarkdownPromptVersion PromptVersion = "2026-02-17"

	// PlaywrightPromptVersion is the version of the Playwright tool instructions prompt
	// File: actions/setup/md/playwright_prompt.md
	// Last modified: Initial version
	PlaywrightPromptVersion PromptVersion = "2026-02-17"

	// PRContextPromptVersion is the version of the PR context instructions prompt
	// File: actions/setup/md/pr_context_prompt.md
	// Last modified: Initial version
	PRContextPromptVersion PromptVersion = "2026-02-17"

	// CacheMemoryPromptVersion is the version of the cache memory instructions prompt
	// File: actions/setup/md/cache_memory_prompt.md
	// Last modified: Initial version
	CacheMemoryPromptVersion PromptVersion = "2026-02-17"

	// CacheMemoryPromptMultiVersion is the version of the multi-cache memory instructions prompt
	// File: actions/setup/md/cache_memory_prompt_multi.md
	// Last modified: Initial version
	CacheMemoryPromptMultiVersion PromptVersion = "2026-02-17"

	// GitHubContextPromptVersion is the version of the GitHub context prompt (embedded)
	// File: pkg/workflow/prompts/github_context_prompt.md
	// Last modified: Initial version
	GitHubContextPromptVersion PromptVersion = "2026-02-17"

	// ThreatDetectionPromptVersion is the version of the threat detection prompt
	// File: actions/setup/md/threat_detection.md
	// Last modified: Initial version
	ThreatDetectionPromptVersion PromptVersion = "2026-02-17"
)

// PromptVersionManifest contains all system prompt versions used in a workflow
type PromptVersionManifest struct {
	// GeneratedAt is the timestamp when the manifest was generated
	GeneratedAt time.Time
	// SystemPrompts maps prompt file names to their versions
	SystemPrompts map[string]PromptVersion
	// CreatorPromptHash is a hash of the user's workflow markdown content
	CreatorPromptHash string
}

// NewPromptVersionManifest creates a new prompt version manifest with current versions
func NewPromptVersionManifest() *PromptVersionManifest {
	return &PromptVersionManifest{
		GeneratedAt: time.Now(),
		SystemPrompts: map[string]PromptVersion{
			"xpia.md":                      XPIAPromptVersion,
			"temp_folder_prompt.md":        TempFolderPromptVersion,
			"markdown.md":                  MarkdownPromptVersion,
			"playwright_prompt.md":         PlaywrightPromptVersion,
			"pr_context_prompt.md":         PRContextPromptVersion,
			"cache_memory_prompt.md":       CacheMemoryPromptVersion,
			"cache_memory_prompt_multi.md": CacheMemoryPromptMultiVersion,
			"github_context_prompt.md":     GitHubContextPromptVersion,
			"threat_detection.md":          ThreatDetectionPromptVersion,
		},
	}
}

// GetVersion returns the version for a specific prompt file
func (m *PromptVersionManifest) GetVersion(filename string) (PromptVersion, bool) {
	version, ok := m.SystemPrompts[filename]
	return version, ok
}

// ToYAMLComment generates a YAML comment block with version information
func (m *PromptVersionManifest) ToYAMLComment() string {
	comment := "# System Prompt Versions:\n"
	comment += fmt.Sprintf("# Generated: %s\n", m.GeneratedAt.Format(time.RFC3339))

	// Sort the keys for consistent output
	keys := make([]string, 0, len(m.SystemPrompts))
	for key := range m.SystemPrompts {
		keys = append(keys, key)
	}

	// Use sort.Strings for alphabetical ordering
	sort.Strings(keys)

	for _, key := range keys {
		version := m.SystemPrompts[key]
		comment += fmt.Sprintf("# - %s: %s\n", key, version)
	}

	if m.CreatorPromptHash != "" {
		comment += fmt.Sprintf("# Creator Prompt Hash: %s\n", m.CreatorPromptHash)
	}

	return comment
}
