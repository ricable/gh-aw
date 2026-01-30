// This file provides post-processing operations for workflow compilation.
//
// This file contains functions that perform post-compilation operations such as
// generating Dependabot manifests, maintenance workflows, and validating campaigns.
//
// # Organization Rationale
//
// These post-processing functions are grouped here because they:
//   - Run after workflow compilation completes
//   - Generate auxiliary files and manifests
//   - Have a clear domain focus (post-compilation processing)
//   - Keep the main orchestrator focused on coordination
//
// # Key Functions
//
// Generation:
//   - generateDependabotManifestsWrapper() - Generate Dependabot manifests
//   - generateMaintenanceWorkflowWrapper() - Generate maintenance workflow
//
// Validation:
//   - validateCampaignsWrapper() - Validate campaign specs
//
// Statistics:
//   - collectWorkflowStatisticsWrapper() - Collect workflow statistics
//
// These functions abstract post-processing operations, allowing the main compile
// orchestrator to focus on coordination while these handle generation and validation.

package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/githubnext/gh-aw/pkg/console"
	"github.com/githubnext/gh-aw/pkg/logger"
	"github.com/githubnext/gh-aw/pkg/stringutil"
	"github.com/githubnext/gh-aw/pkg/workflow"
)

var compilePostProcessingLog = logger.New("cli:compile_post_processing")

// generateDependabotManifestsWrapper generates Dependabot manifests for compiled workflows
func generateDependabotManifestsWrapper(
	compiler *workflow.Compiler,
	workflowDataList []*workflow.WorkflowData,
	workflowsDir string,
	forceOverwrite bool,
	strict bool,
) error {
	compilePostProcessingLog.Print("Generating Dependabot manifests for compiled workflows")

	if err := compiler.GenerateDependabotManifests(workflowDataList, workflowsDir, forceOverwrite); err != nil {
		if strict {
			return fmt.Errorf("failed to generate Dependabot manifests: %w", err)
		}
		// Non-strict mode: just report as warning
		fmt.Fprintln(os.Stderr, console.FormatWarningMessage(fmt.Sprintf("Failed to generate Dependabot manifests: %v", err)))
	}

	return nil
}

// generateMaintenanceWorkflowWrapper generates maintenance workflow if any workflow uses expires field
func generateMaintenanceWorkflowWrapper(
	compiler *workflow.Compiler,
	workflowDataList []*workflow.WorkflowData,
	workflowsDir string,
	verbose bool,
	strict bool,
) error {
	compilePostProcessingLog.Print("Generating maintenance workflow")

	if err := workflow.GenerateMaintenanceWorkflow(workflowDataList, workflowsDir, compiler.GetVersion(), compiler.GetActionMode(), verbose); err != nil {
		if strict {
			return fmt.Errorf("failed to generate maintenance workflow: %w", err)
		}
		// Non-strict mode: just report as warning
		fmt.Fprintln(os.Stderr, console.FormatWarningMessage(fmt.Sprintf("Failed to generate maintenance workflow: %v", err)))
	}

	return nil
}

// validateCampaignsWrapper validates campaign specs if they exist
func validateCampaignsWrapper(workflowDir string, verbose bool, strict bool, campaignFiles []string) error {
	compilePostProcessingLog.Print("Validating campaign specs")

	if err := validateCampaigns(workflowDir, verbose, campaignFiles); err != nil {
		if strict {
			return fmt.Errorf("campaign validation failed: %w", err)
		}
		// Non-strict mode: just report as warning
		fmt.Fprintln(os.Stderr, console.FormatWarningMessage(fmt.Sprintf("Campaign validation: %v", err)))
	}

	return nil
}

// collectWorkflowStatisticsWrapper collects and returns workflow statistics
func collectWorkflowStatisticsWrapper(markdownFiles []string) []*WorkflowStats {
	compilePostProcessingLog.Printf("Collecting workflow statistics for %d files", len(markdownFiles))

	var statsList []*WorkflowStats
	for _, file := range markdownFiles {
		resolvedFile, err := resolveWorkflowFile(file, false)
		if err != nil {
			continue // Skip files that couldn't be resolved
		}
		lockFile := stringutil.MarkdownToLockFile(resolvedFile)
		if workflowStats, err := collectWorkflowStats(lockFile); err == nil {
			statsList = append(statsList, workflowStats)
		}
	}

	compilePostProcessingLog.Printf("Collected statistics for %d workflows", len(statsList))
	return statsList
}

// updateGitAttributes ensures .gitattributes marks .lock.yml files as generated
func updateGitAttributes(successCount int, actionCache *workflow.ActionCache, verbose bool) error {
	compilePostProcessingLog.Printf("Updating .gitattributes (compiled=%d, actionCache=%v)", successCount, actionCache != nil)

	hasActionCacheEntries := actionCache != nil && len(actionCache.Entries) > 0

	// Only update if we successfully compiled workflows or have action cache entries
	if successCount > 0 || hasActionCacheEntries {
		compilePostProcessingLog.Printf("Updating .gitattributes (compiled=%d, actionCache=%v)", successCount, hasActionCacheEntries)
		if err := ensureGitAttributes(); err != nil {
			compilePostProcessingLog.Printf("Failed to update .gitattributes: %v", err)
			if verbose {
				fmt.Fprintln(os.Stderr, console.FormatWarningMessage(fmt.Sprintf("Failed to update .gitattributes: %v", err)))
			}
			return err
		}
		compilePostProcessingLog.Printf("Successfully updated .gitattributes")
		if verbose {
			fmt.Fprintln(os.Stderr, console.FormatSuccessMessage("Updated .gitattributes to mark .lock.yml files as generated"))
		}
	} else {
		compilePostProcessingLog.Print("Skipping .gitattributes update (no compiled workflows and no action cache entries)")
	}

	return nil
}

// ensureGitAttributes ensures that .gitattributes contains the entry to mark .lock.yml files as generated
func ensureGitAttributes() error {
	compilePostProcessingLog.Print("Ensuring .gitattributes is updated")
	gitRoot, err := findGitRoot()
	if err != nil {
		return err // Not in a git repository, skip
	}

	gitAttributesPath := filepath.Join(gitRoot, ".gitattributes")
	lockYmlEntry := ".github/workflows/*.lock.yml linguist-generated=true merge=ours"
	requiredEntries := []string{lockYmlEntry}

	// Read existing .gitattributes file if it exists
	var lines []string
	if content, err := os.ReadFile(gitAttributesPath); err == nil {
		lines = strings.Split(string(content), "\n")
		compilePostProcessingLog.Printf("Read existing .gitattributes with %d lines", len(lines))
	} else {
		compilePostProcessingLog.Print("No existing .gitattributes file found")
	}

	modified := false
	for _, required := range requiredEntries {
		found := false
		for i, line := range lines {
			trimmedLine := strings.TrimSpace(line)
			if trimmedLine == required {
				found = true
				break
			}
			// Check for old format entries that need updating
			if strings.HasPrefix(trimmedLine, ".github/workflows/*.lock.yml") && required == lockYmlEntry {
				compilePostProcessingLog.Print("Updating old .gitattributes entry format")
				lines[i] = lockYmlEntry
				found = true
				modified = true
				break
			}
		}

		if !found {
			compilePostProcessingLog.Printf("Adding new .gitattributes entry: %s", required)
			if len(lines) > 0 && lines[len(lines)-1] != "" {
				lines = append(lines, "")
			}
			lines = append(lines, required)
			modified = true
		}
	}

	// Remove old campaign.g.md entries if they exist (they're now in .gitignore)
	for i := len(lines) - 1; i >= 0; i-- {
		trimmedLine := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmedLine, ".github/workflows/*.campaign.g.md") {
			compilePostProcessingLog.Print("Removing obsolete .campaign.g.md .gitattributes entry")
			lines = append(lines[:i], lines[i+1:]...)
			modified = true
		}
	}

	if !modified {
		compilePostProcessingLog.Print(".gitattributes already contains required entries")
		return nil
	}

	// Write back to file with owner-only read/write permissions (0600) for security best practices
	content := strings.Join(lines, "\n")
	if err := os.WriteFile(gitAttributesPath, []byte(content), 0600); err != nil {
		compilePostProcessingLog.Printf("Failed to write .gitattributes: %v", err)
		return fmt.Errorf("failed to write .gitattributes: %w", err)
	}

	compilePostProcessingLog.Print("Successfully updated .gitattributes")
	return nil
}

// stageGitAttributesIfChanged stages .gitattributes if it was modified
func stageGitAttributesIfChanged() error {
	gitRoot, err := findGitRoot()
	if err != nil {
		return err
	}
	gitAttributesPath := filepath.Join(gitRoot, ".gitattributes")
	return exec.Command("git", "-C", gitRoot, "add", gitAttributesPath).Run()
}

// saveActionCache saves the action cache after all compilations
func saveActionCache(actionCache *workflow.ActionCache, verbose bool) error {
	if actionCache == nil {
		return nil
	}

	compilePostProcessingLog.Print("Saving action cache")

	if err := actionCache.Save(); err != nil {
		compilePostProcessingLog.Printf("Failed to save action cache: %v", err)
		if verbose {
			fmt.Fprintln(os.Stderr, console.FormatWarningMessage(fmt.Sprintf("Failed to save action cache: %v", err)))
		}
		return err
	}

	compilePostProcessingLog.Print("Action cache saved successfully")
	if verbose {
		fmt.Fprintln(os.Stderr, console.FormatSuccessMessage(fmt.Sprintf("Action cache saved to %s", actionCache.GetCachePath())))
	}

	return nil
}

// getAbsoluteWorkflowDir converts a relative workflow dir to absolute path
func getAbsoluteWorkflowDir(workflowDir string, gitRoot string) string {
	absWorkflowDir := workflowDir
	if !filepath.IsAbs(absWorkflowDir) {
		absWorkflowDir = filepath.Join(gitRoot, workflowDir)
	}
	return absWorkflowDir
}
