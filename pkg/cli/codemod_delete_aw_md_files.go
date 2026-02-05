package cli

// getDeleteAgenticWorkflowPromptFilesCodemod creates a codemod for deleting .github/aw/*.md files
func getDeleteAgenticWorkflowPromptFilesCodemod() Codemod {
	return Codemod{
		ID:           "delete-aw-md-files",
		Name:         "Delete agentic workflow prompt markdown files",
		Description:  "Deletes all .github/aw/*.md files which are now downloaded from GitHub on demand",
		IntroducedIn: "0.7.0",
		Apply: func(content string, frontmatter map[string]any) (string, bool, error) {
			// This codemod is handled by the fix command itself (see runFixCommand)
			// It doesn't modify workflow files, so we just return content unchanged
			return content, false, nil
		},
	}
}
