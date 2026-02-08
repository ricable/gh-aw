package workflow

import "github.com/github/gh-aw/pkg/logger"

var permissionsBuilderLog = logger.New("workflow:permissions_builder")

// PermissionsBuilder provides a fluent API for building Permissions objects.
//
// The builder pattern replaces the previous factory explosion approach (23 constructors)
// with a composable, type-safe API that scales to any permission combination.
//
// # Basic Usage
//
//	// Simple single permission
//	perms := NewPermissionsBuilder().
//		WithContents(PermissionRead).
//		Build()
//
//	// Multiple permissions
//	perms := NewPermissionsBuilder().
//		WithContents(PermissionRead).
//		WithIssues(PermissionWrite).
//		WithPullRequests(PermissionWrite).
//		Build()
//
// # Advanced Usage
//
//	// Complex permission combinations
//	perms := NewPermissionsBuilder().
//		WithActions(PermissionWrite).
//		WithContents(PermissionWrite).
//		WithIssues(PermissionWrite).
//		WithPullRequests(PermissionWrite).
//		WithDiscussions(PermissionWrite).
//		WithSecurityEvents(PermissionWrite).
//		Build()
//
// # Backward Compatibility
//
// Existing factory functions remain available as deprecated wrappers:
//
//	// Old style (still works, but deprecated)
//	perms := NewPermissionsContentsReadIssuesWrite()
//
//	// New style (preferred)
//	perms := NewPermissionsBuilder().
//		WithContents(PermissionRead).
//		WithIssues(PermissionWrite).
//		Build()
//
// Both produce identical results, but the builder pattern is more flexible
// and doesn't require a new factory function for each combination.
//
// # Available Permission Scopes
//
// The builder provides methods for all GitHub Actions permission scopes:
//   - WithActions(level)          - actions permission
//   - WithAttestations(level)     - attestations permission
//   - WithChecks(level)           - checks permission
//   - WithContents(level)         - contents permission
//   - WithDeployments(level)      - deployments permission
//   - WithDiscussions(level)      - discussions permission
//   - WithIdToken(level)          - id-token permission
//   - WithIssues(level)           - issues permission
//   - WithMetadata(level)         - metadata permission
//   - WithModels(level)           - models permission
//   - WithPackages(level)         - packages permission
//   - WithPages(level)            - pages permission
//   - WithPullRequests(level)     - pull-requests permission
//   - WithRepositoryProjects(level)      - repository-projects permission
//   - WithOrganizationProjects(level)    - organization-projects permission
//   - WithSecurityEvents(level)   - security-events permission
//   - WithStatuses(level)         - statuses permission
//
// # Permission Levels
//
// Each scope accepts one of three levels:
//   - PermissionRead  - read access
//   - PermissionWrite - write access (implies read)
//   - PermissionNone  - explicitly no access
type PermissionsBuilder struct {
	perms *Permissions
}

// NewPermissionsBuilder creates a new PermissionsBuilder with an empty permissions map
func NewPermissionsBuilder() *PermissionsBuilder {
	if permissionsBuilderLog.Enabled() {
		permissionsBuilderLog.Print("Creating new PermissionsBuilder")
	}
	return &PermissionsBuilder{
		perms: NewPermissions(),
	}
}

// WithActions sets the actions permission level
func (pb *PermissionsBuilder) WithActions(level PermissionLevel) *PermissionsBuilder {
	pb.perms.Set(PermissionActions, level)
	return pb
}

// WithAttestations sets the attestations permission level
func (pb *PermissionsBuilder) WithAttestations(level PermissionLevel) *PermissionsBuilder {
	pb.perms.Set(PermissionAttestations, level)
	return pb
}

// WithChecks sets the checks permission level
func (pb *PermissionsBuilder) WithChecks(level PermissionLevel) *PermissionsBuilder {
	pb.perms.Set(PermissionChecks, level)
	return pb
}

// WithContents sets the contents permission level
func (pb *PermissionsBuilder) WithContents(level PermissionLevel) *PermissionsBuilder {
	pb.perms.Set(PermissionContents, level)
	return pb
}

// WithDeployments sets the deployments permission level
func (pb *PermissionsBuilder) WithDeployments(level PermissionLevel) *PermissionsBuilder {
	pb.perms.Set(PermissionDeployments, level)
	return pb
}

// WithDiscussions sets the discussions permission level
func (pb *PermissionsBuilder) WithDiscussions(level PermissionLevel) *PermissionsBuilder {
	pb.perms.Set(PermissionDiscussions, level)
	return pb
}

// WithIdToken sets the id-token permission level
func (pb *PermissionsBuilder) WithIdToken(level PermissionLevel) *PermissionsBuilder {
	pb.perms.Set(PermissionIdToken, level)
	return pb
}

// WithIssues sets the issues permission level
func (pb *PermissionsBuilder) WithIssues(level PermissionLevel) *PermissionsBuilder {
	pb.perms.Set(PermissionIssues, level)
	return pb
}

// WithMetadata sets the metadata permission level
func (pb *PermissionsBuilder) WithMetadata(level PermissionLevel) *PermissionsBuilder {
	pb.perms.Set(PermissionMetadata, level)
	return pb
}

// WithModels sets the models permission level
func (pb *PermissionsBuilder) WithModels(level PermissionLevel) *PermissionsBuilder {
	pb.perms.Set(PermissionModels, level)
	return pb
}

// WithPackages sets the packages permission level
func (pb *PermissionsBuilder) WithPackages(level PermissionLevel) *PermissionsBuilder {
	pb.perms.Set(PermissionPackages, level)
	return pb
}

// WithPages sets the pages permission level
func (pb *PermissionsBuilder) WithPages(level PermissionLevel) *PermissionsBuilder {
	pb.perms.Set(PermissionPages, level)
	return pb
}

// WithPullRequests sets the pull-requests permission level
func (pb *PermissionsBuilder) WithPullRequests(level PermissionLevel) *PermissionsBuilder {
	pb.perms.Set(PermissionPullRequests, level)
	return pb
}

// WithRepositoryProjects sets the repository-projects permission level
func (pb *PermissionsBuilder) WithRepositoryProjects(level PermissionLevel) *PermissionsBuilder {
	pb.perms.Set(PermissionRepositoryProj, level)
	return pb
}

// WithOrganizationProjects sets the organization-projects permission level
// Note: organization-projects is only valid for GitHub App tokens, not workflow permissions
func (pb *PermissionsBuilder) WithOrganizationProjects(level PermissionLevel) *PermissionsBuilder {
	pb.perms.Set(PermissionOrganizationProj, level)
	return pb
}

// WithSecurityEvents sets the security-events permission level
func (pb *PermissionsBuilder) WithSecurityEvents(level PermissionLevel) *PermissionsBuilder {
	pb.perms.Set(PermissionSecurityEvents, level)
	return pb
}

// WithStatuses sets the statuses permission level
func (pb *PermissionsBuilder) WithStatuses(level PermissionLevel) *PermissionsBuilder {
	pb.perms.Set(PermissionStatuses, level)
	return pb
}

// Build returns the constructed Permissions object
func (pb *PermissionsBuilder) Build() *Permissions {
	if permissionsBuilderLog.Enabled() {
		permissionsBuilderLog.Printf("Building permissions: scope_count=%d", len(pb.perms.permissions))
	}
	return pb.perms
}
