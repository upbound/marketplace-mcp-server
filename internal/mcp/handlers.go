package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pkg/errors"

	"github.com/upbound/marketplace-mcp-server/internal/marketplace"
)

// handleSearchPackages handles the search_packages tool.
func (s *Server) handleSearchPackages(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters using built-in methods
	query := req.GetString("query", "")
	family := req.GetString("family", "")
	packageType := req.GetString("package_type", "")
	accountName := req.GetString("account_name", "")
	tier := req.GetString("tier", "")
	size := req.GetInt("size", 20)
	page := req.GetInt("page", 0)
	useV1 := req.GetBool("use_v1", false)

	// Handle public parameter (optional boolean)
	var public *bool
	args := req.GetArguments()
	if val, ok := args["public"]; ok {
		if b, ok := val.(bool); ok {
			public = &b
		}
	}

	// Prepare search parameters
	params := marketplace.SearchParams{
		Query:       query,
		Family:      family,
		PackageType: packageType,
		AccountName: accountName,
		Tier:        tier,
		Size:        size,
		Page:        page,
		UseV1:       useV1,
		Public:      public,
	}

	// Perform search
	result, err := s.client.SearchPackages(ctx, params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Search failed: %v", err)), err
	}

	return mcp.NewToolResultText(formatSearchResults(result)), nil
}

// handleGetPackageMetadata handles the get_package_metadata tool.
func (s *Server) handleGetPackageMetadata(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required parameters
	account, err := req.RequireString("account")
	if err != nil {
		return mcp.NewToolResultError("account parameter is required"), err
	}

	repository, err := req.RequireString("repository")
	if err != nil {
		return mcp.NewToolResultError("repository parameter is required"), err
	}

	// Extract optional parameters
	version := req.GetString("version", "")
	useV1 := req.GetBool("use_v1", false)

	// Get package metadata
	metadata, err := s.client.GetPackageMetadata(ctx, account, repository, version, useV1)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get package metadata: %v", err)), err
	}

	return mcp.NewToolResultText(formatPackageMetadata(metadata)), nil
}

// handleGetPackageAssets handles the get_package_assets tool.
func (s *Server) handleGetPackageAssets(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required parameters
	account, err := req.RequireString("account")
	if err != nil {
		return mcp.NewToolResultError("account parameter is required"), err
	}

	repository, err := req.RequireString("repository")
	if err != nil {
		return mcp.NewToolResultError("repository parameter is required"), err
	}

	version, err := req.RequireString("version")
	if err != nil {
		return mcp.NewToolResultError("version parameter is required"), err
	}

	assetType, err := req.RequireString("asset_type")
	if err != nil {
		return mcp.NewToolResultError("asset_type parameter is required"), err
	}

	// Validate asset type
	validAssetTypes := map[string]bool{
		"docs":         true,
		"icon":         true,
		"readme":       true,
		"releaseNotes": true,
		"sbom":         true,
	}
	if !validAssetTypes[assetType] {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid asset_type: %s. Must be one of: docs, icon, readme, releaseNotes, sbom", assetType)), nil
	}

	// Get package assets
	assets, err := s.client.GetPackageAssets(ctx, account, repository, version, assetType)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get package assets: %v", err)), nil
	}

	return mcp.NewToolResultText(formatPackageAssets(assets, assetType)), nil
}

// handleGetRepositories handles the get_repositories tool.
func (s *Server) handleGetRepositories(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required parameters
	account, err := req.RequireString("account")
	if err != nil {
		return mcp.NewToolResultError("account parameter is required"), err
	}

	// Extract optional parameters
	filter := req.GetString("filter", "")
	size := req.GetInt("size", 20)
	page := req.GetInt("page", 0)
	useV1 := req.GetBool("use_v1", false)

	// Prepare repository parameters
	params := marketplace.RepositoryParams{
		Filter: filter,
		Size:   size,
		Page:   page,
		UseV1:  useV1,
	}

	// Get repositories
	repos, err := s.client.GetRepositories(ctx, account, params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get repositories: %v", err)), err
	}

	return mcp.NewToolResultText(formatRepositories(repos)), nil
}

// handleGetPackagesAccountRepositoryVersionResources handles the get_repositories tool.
func (s *Server) handleGetPackagesAccountRepositoryVersionResources(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required parameters
	account, err := req.RequireString("account")
	if err != nil {
		return mcp.NewToolResultError("account parameter is required"), err
	}
	repositoryName, err := req.RequireString("repository_name")
	if err != nil {
		return mcp.NewToolResultError("repository_name parameter is required"), err
	}
	version, err := req.RequireString("version")
	if err != nil {
		return mcp.NewToolResultError("version parameter is required"), err
	}

	// Get repositories
	repos, err := s.client.GetV1PackagesAccountRepositoryVersionResources(ctx, account, repositoryName, version)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get repositories: %v", err)), err
	}

	b, err := json.Marshal(repos)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal response")
	}

	return mcp.NewToolResultText(string(b)), nil
}

// handleGetPackagesAccountRepositoryVersionResources handles the get_repositories tool.
func (s *Server) handleGetPackagesAccountRepositoryVersionResourcesGroupKindComposition(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required parameters
	account, err := req.RequireString("account")
	if err != nil {
		return mcp.NewToolResultError("account parameter is required"), err
	}
	repositoryName, err := req.RequireString("repository_name")
	if err != nil {
		return mcp.NewToolResultError("repository_name parameter is required"), err
	}
	version, err := req.RequireString("version")
	if err != nil {
		return mcp.NewToolResultError("version parameter is required"), err
	}
	resourceGroup, err := req.RequireString("resource_group")
	if err != nil {
		return mcp.NewToolResultError("resource_group parameter is required"), err
	}
	resourceKind, err := req.RequireString("resource_kind")
	if err != nil {
		return mcp.NewToolResultError("resource_kind parameter is required"), err
	}
	compositionName, err := req.RequireString("composition_name")
	if err != nil {
		return mcp.NewToolResultError("composition_name parameter is required"), err
	}

	// Get repositories
	raw, err := s.client.GetV1PackagesAccountRepositoryVersionResourcesGroupKindComposition(ctx, account, repositoryName, version, resourceGroup, resourceKind, compositionName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get repositories: %v", err)), err
	}

	return mcp.NewToolResultText(raw), nil
}

// handleGetPackagesAccountRepositoryVersionResourcesGroupKind handles the get_repositories tool.
func (s *Server) handleGetPackagesAccountRepositoryVersionResourcesGroupKind(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required parameters
	account, err := req.RequireString("account")
	if err != nil {
		return mcp.NewToolResultError("account parameter is required"), err
	}
	repositoryName, err := req.RequireString("repository_name")
	if err != nil {
		return mcp.NewToolResultError("repository_name parameter is required"), err
	}
	version, err := req.RequireString("version")
	if err != nil {
		return mcp.NewToolResultError("version parameter is required"), err
	}
	resourceGroup, err := req.RequireString("resource_group")
	if err != nil {
		return mcp.NewToolResultError("resource_group parameter is required"), err
	}
	resourceKind, err := req.RequireString("resource_kind")
	if err != nil {
		return mcp.NewToolResultError("resource_kind parameter is required"), err
	}

	// Get repositories
	raw, err := s.client.GetV1PackagesAccountRepositoryVersionResourcesGroupKind(ctx, account, repositoryName, version, resourceGroup, resourceKind)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get repositories: %v", err)), err
	}

	return mcp.NewToolResultText(raw), nil
}

// handleGetPackagesAccountRepositoryVersionResourcesGroupKind handles the get_repositories tool.
func (s *Server) handleGetPackagesAccountRepositoryVersionResourcesGroupKindExamples(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract required parameters
	account, err := req.RequireString("account")
	if err != nil {
		return mcp.NewToolResultError("account parameter is required"), err
	}
	repositoryName, err := req.RequireString("repository_name")
	if err != nil {
		return mcp.NewToolResultError("repository_name parameter is required"), err
	}
	version, err := req.RequireString("version")
	if err != nil {
		return mcp.NewToolResultError("version parameter is required"), err
	}
	resourceGroup, err := req.RequireString("resource_group")
	if err != nil {
		return mcp.NewToolResultError("resource_group parameter is required"), err
	}
	resourceKind, err := req.RequireString("resource_kind")
	if err != nil {
		return mcp.NewToolResultError("resource_kind parameter is required"), err
	}

	// Get repositories
	exs, err := s.client.GetV1PackagesAccountRepositoryVersionResourcesGroupKindExamples(ctx, account, repositoryName, version, resourceGroup, resourceKind)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get repositories: %v", err)), err
	}

	b, err := json.Marshal(exs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal response")
	}

	return mcp.NewToolResultText(string(b)), nil
}

// handleReloadAuth handles the reload_auth tool.
func (s *Server) handleReloadAuth(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Try to reload authentication token from UP CLI config
	token, err := s.authManager.GetCurrentToken()
	if err == nil {
		s.client.SetToken(token.AccessToken)

		// Also reload server URL
		if serverURL, err := s.authManager.GetCurrentServerURL(); err == nil {
			s.client.SetBaseURL(serverURL)
			return mcp.NewToolResultText(fmt.Sprintf("Successfully reloaded authentication and server configuration.\nServer URL: %s\nAuthentication: Loaded from UP CLI profile", serverURL)), nil
		}
		return mcp.NewToolResultText("Successfully reloaded authentication token from UP CLI profile, but failed to reload server URL."), nil
	}
	return mcp.NewToolResultError(fmt.Sprintf("Failed to reload authentication from UP CLI: %v. Please ensure you are logged in with 'up login'.", err)), nil
}

// formatSearchResults formats search results for display.
func formatSearchResults(result *marketplace.SearchResponse) string {
	if result == nil {
		return "No search results"
	}

	output := fmt.Sprintf("Search Results (Total: %d)\n", result.Total)
	output += "=====================================\n\n"

	for i, pkg := range result.Packages {
		output += fmt.Sprintf("%d. %s/%s\n", i+1, pkg.Account, pkg.Repository)
		if pkg.Name != "" {
			output += fmt.Sprintf("   Name: %s\n", pkg.Name)
		}
		if pkg.Description != "" {
			output += fmt.Sprintf("   Description: %s\n", pkg.Description)
		}
		if pkg.Version != "" {
			output += fmt.Sprintf("   Version: %s\n", pkg.Version)
		}
		if pkg.Type != "" {
			output += fmt.Sprintf("   Type: %s\n", pkg.Type)
		}
		if pkg.Tier != "" {
			output += fmt.Sprintf("   Tier: %s\n", pkg.Tier)
		}
		if len(pkg.Tags) > 0 {
			output += fmt.Sprintf("   Tags: %v\n", pkg.Tags)
		}
		output += "\n"
	}

	return output
}

// formatPackageMetadata formats package metadata for display.
func formatPackageMetadata(metadata *marketplace.PackageMetadata) string {
	if metadata == nil {
		return "No package metadata"
	}

	output := fmt.Sprintf("Package: %s/%s\n", metadata.Account, metadata.Repository)
	output += "=====================================\n\n"

	if metadata.Name != "" {
		output += fmt.Sprintf("Name: %s\n", metadata.Name)
	}
	if metadata.Description != "" {
		output += fmt.Sprintf("Description: %s\n", metadata.Description)
	}
	if metadata.Version != "" {
		output += fmt.Sprintf("Version: %s\n", metadata.Version)
	}
	if metadata.Type != "" {
		output += fmt.Sprintf("Type: %s\n", metadata.Type)
	}
	if metadata.Tier != "" {
		output += fmt.Sprintf("Tier: %s\n", metadata.Tier)
	}
	if len(metadata.Tags) > 0 {
		output += fmt.Sprintf("Tags: %v\n", metadata.Tags)
	}
	if !metadata.CreatedAt.IsZero() {
		output += fmt.Sprintf("Created: %s\n", metadata.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	if !metadata.UpdatedAt.IsZero() {
		output += fmt.Sprintf("Updated: %s\n", metadata.UpdatedAt.Format("2006-01-02 15:04:05"))
	}
	if metadata.Downloads > 0 {
		output += fmt.Sprintf("Downloads: %d\n", metadata.Downloads)
	}

	// Add CRDs information if available
	if len(metadata.CRDs) > 0 {
		output += "\nCustom Resource Definitions (CRDs):\n"
		output += "-----------------------------------\n"
		for i, crd := range metadata.CRDs {
			output += fmt.Sprintf("%d. %s\n", i+1, crd.Name)
			if crd.Group != "" {
				output += fmt.Sprintf("   Group: %s\n", crd.Group)
			}
			if crd.Version != "" {
				output += fmt.Sprintf("   Version: %s\n", crd.Version)
			}
			if crd.Kind != "" {
				output += fmt.Sprintf("   Kind: %s\n", crd.Kind)
			}
			if crd.Description != "" {
				output += fmt.Sprintf("   Description: %s\n", crd.Description)
			}
			output += "\n"
		}
	}

	return output
}

// formatPackageAssets formats package assets for display.
func formatPackageAssets(assets *marketplace.AssetResponse, assetType string) string {
	if assets == nil {
		return fmt.Sprintf("No %s assets found", assetType)
	}

	output := fmt.Sprintf("Package Assets (%s):\n", assetType)
	output += "=====================================\n\n"

	switch assetType {
	case "docs", "readme", "releaseNotes":
		switch {
		case assets.Content != "":
			output += assets.Content
		case assets.URL != "":
			output += fmt.Sprintf("Asset URL: %s", assets.URL)
		default:
			output += "No content available"
		}
	case "icon":
		switch {
		case assets.URL != "":
			output += fmt.Sprintf("Icon URL: %s", assets.URL)
		default:
			output += "Icon asset retrieved (binary data)"
		}
	case "sbom":
		switch {
		case assets.Content != "":
			output += assets.Content
		case assets.URL != "":
			output += fmt.Sprintf("SBOM URL: %s", assets.URL)
		default:
			output += "No SBOM content available"
		}
	default:
		switch {
		case assets.Content != "":
			output += assets.Content
		case assets.URL != "":
			output += fmt.Sprintf("Asset URL: %s", assets.URL)
		default:
			output += "No asset content available"
		}
	}

	return output
}

// formatRepositories formats repositories for display.
func formatRepositories(repos *marketplace.RepositoryResponse) string {
	if repos == nil {
		return "No repositories found"
	}

	output := fmt.Sprintf("Repositories (Count: %d)\n", repos.Count)
	output += "=====================================\n\n"

	for i, repo := range repos.Repositories {
		output += fmt.Sprintf("%d. %s\n", i+1, repo.Name)
		if repo.Description != "" {
			output += fmt.Sprintf("   Description: %s\n", repo.Description)
		}
		if repo.Type != "" {
			output += fmt.Sprintf("   Type: %s\n", repo.Type)
		}
		if !repo.CreatedAt.IsZero() {
			output += fmt.Sprintf("   Created: %s\n", repo.CreatedAt.Format("2006-01-02 15:04:05"))
		}
		if !repo.UpdatedAt.IsZero() {
			output += fmt.Sprintf("   Updated: %s\n", repo.UpdatedAt.Format("2006-01-02 15:04:05"))
		}
		if repo.PackageCount > 0 {
			output += fmt.Sprintf("   Package Count: %d\n", repo.PackageCount)
		}
		output += "\n"
	}

	return output
}
