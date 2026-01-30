// This file provides validation for campaign orchestrator project requirements.
//
// # Campaign Project Validation
//
// This file ensures that workflows with campaign characteristics (such as campaign labels
// or campaign IDs) have a required GitHub Project URL configured for tracking their work.
//
// Campaign orchestrators coordinate multiple workflows and track progress on GitHub Project
// boards. Without a project URL, the orchestrator cannot track Dependabot PRs, bundle issues,
// or other campaign work items.
//
// # Detection Criteria
//
// A workflow is considered a campaign orchestrator if it has:
//   - Campaign labels in safe-outputs (agentic-campaign or z_campaign_*)
//   - Campaign ID configured in repo-memory tools
//
// # Validation Rules
//
// When campaign characteristics are detected:
//   - A project field must be present in frontmatter
//   - The project field must be a non-empty string or valid project config object
//
// # When to Update This File
//
// Update this validation when:
//   - New campaign detection patterns are added
//   - Project configuration requirements change
//   - Campaign orchestration patterns evolve

package workflow

import (
	"fmt"
	"strings"

	"github.com/githubnext/gh-aw/pkg/constants"
	"github.com/githubnext/gh-aw/pkg/logger"
)

var campaignProjectValidationLog = logger.New("workflow:campaign_project_validation")

// validateCampaignProject checks if a workflow with campaign characteristics has a project URL configured
func (c *Compiler) validateCampaignProject(frontmatter map[string]any) error {
	campaignProjectValidationLog.Print("Checking campaign project requirements")

	// Check if this workflow has campaign characteristics
	isCampaignWorkflow, campaignSource := detectCampaignWorkflow(frontmatter)
	if !isCampaignWorkflow {
		campaignProjectValidationLog.Print("Workflow is not a campaign orchestrator, skipping validation")
		return nil
	}

	campaignProjectValidationLog.Printf("Detected campaign workflow via %s", campaignSource)

	// Check if project field exists
	projectData, hasProject := frontmatter["project"]
	if !hasProject || projectData == nil {
		return fmt.Errorf("campaign orchestrator requires a GitHub Project URL to track work items. Please add a 'project' field to the frontmatter with a valid GitHub Project URL (e.g., project: https://github.com/orgs/myorg/projects/123). Campaign detected via: %s", campaignSource)
	}

	// Validate project field is not empty
	switch v := projectData.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return fmt.Errorf("campaign orchestrator requires a non-empty GitHub Project URL. Campaign detected via: %s", campaignSource)
		}
		campaignProjectValidationLog.Printf("Valid project URL found: %s", v)
	case map[string]any:
		// Check if object has a URL field
		if url, hasURL := v["url"]; !hasURL || url == nil {
			return fmt.Errorf("campaign orchestrator project configuration must include a 'url' field with a valid GitHub Project URL. Campaign detected via: %s", campaignSource)
		} else if urlStr, ok := url.(string); !ok || strings.TrimSpace(urlStr) == "" {
			return fmt.Errorf("campaign orchestrator project URL must be a non-empty string. Campaign detected via: %s", campaignSource)
		}
		campaignProjectValidationLog.Print("Valid project configuration object found")
	default:
		return fmt.Errorf("campaign orchestrator 'project' field must be a string URL or configuration object. Campaign detected via: %s", campaignSource)
	}

	campaignProjectValidationLog.Print("Campaign project validation passed")
	return nil
}

// detectCampaignWorkflow checks if a workflow has campaign characteristics
// Returns (isCampaign bool, source string) where source explains why it's detected as a campaign
func detectCampaignWorkflow(frontmatter map[string]any) (bool, string) {
	// Check for campaign labels in safe-outputs
	if hasCampaignLabels(frontmatter) {
		return true, "campaign labels in safe-outputs (agentic-campaign or z_campaign_*)"
	}

	// Check for campaign-id in repo-memory tools
	if hasCampaignID(frontmatter) {
		return true, "campaign-id in repo-memory configuration"
	}

	return false, ""
}

// hasCampaignLabels checks if safe-outputs configuration includes campaign labels
func hasCampaignLabels(frontmatter map[string]any) bool {
	safeOutputs, ok := frontmatter["safe-outputs"].(map[string]any)
	if !ok {
		return false
	}

	// Check all safe-output types that support labels
	labelConfigs := []string{
		"add-labels",
		"create-issue",
		"create-pull-request",
		"create-discussion",
	}

	for _, configKey := range labelConfigs {
		if hasLabelsInConfig(safeOutputs, configKey) {
			return true
		}
	}

	return false
}

// hasLabelsInConfig checks if a specific safe-output config contains campaign labels
func hasLabelsInConfig(safeOutputs map[string]any, configKey string) bool {
	config, ok := safeOutputs[configKey].(map[string]any)
	if !ok {
		return false
	}

	// Check for "allowed" field in add-labels
	if configKey == "add-labels" {
		if allowed, ok := config["allowed"].([]any); ok {
			for _, label := range allowed {
				if labelStr, ok := label.(string); ok && isCampaignLabel(labelStr) {
					return true
				}
			}
		}
	}

	// Check for "labels" field in other safe-outputs
	if labels, ok := config["labels"].([]any); ok {
		for _, label := range labels {
			if labelStr, ok := label.(string); ok && isCampaignLabel(labelStr) {
				return true
			}
		}
	}

	return false
}

// isCampaignLabel checks if a label string is a campaign label
func isCampaignLabel(label string) bool {
	// Check for exact match with AgenticCampaignLabel
	if label == string(constants.AgenticCampaignLabel) {
		return true
	}

	// Check for z_campaign_ prefix
	if strings.HasPrefix(label, string(constants.CampaignLabelPrefix)) {
		return true
	}

	return false
}

// hasCampaignID checks if tools.repo-memory configuration includes a campaign-id
func hasCampaignID(frontmatter map[string]any) bool {
	tools, ok := frontmatter["tools"].(map[string]any)
	if !ok {
		return false
	}

	repoMemory, ok := tools["repo-memory"]
	if !ok {
		return false
	}

	// repo-memory can be a single config object or an array of config objects
	switch v := repoMemory.(type) {
	case map[string]any:
		// Single repo-memory configuration
		if campaignID, exists := v["campaign-id"]; exists && campaignID != nil {
			if idStr, ok := campaignID.(string); ok && strings.TrimSpace(idStr) != "" {
				return true
			}
		}
	case []any:
		// Array of repo-memory configurations
		for _, item := range v {
			if itemMap, ok := item.(map[string]any); ok {
				if campaignID, exists := itemMap["campaign-id"]; exists && campaignID != nil {
					if idStr, ok := campaignID.(string); ok && strings.TrimSpace(idStr) != "" {
						return true
					}
				}
			}
		}
	}

	return false
}
