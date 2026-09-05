// /*
// Copyright 2025 The Upbound Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// */

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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

	if len(result.Packages) == 0 {
		return "Search Results (0 matches)\n" +
			"=====================================\n\n" +
			"No packages matched the search. Note that package_type must be one of " +
			"Provider, Configuration, Function or Addon, tier one of official, partner " +
			"or community, and that all filters are combined with AND."
	}

	output := fmt.Sprintf("Search Results (%d matches, showing %d)\n", result.Count, len(result.Packages))
	output += "=====================================\n\n"

	for i, pkg := range result.Packages {
		output += fmt.Sprintf("%d. %s/%s\n", i+1, pkg.Account, pkg.Repository)
		if pkg.Name != "" {
			output += fmt.Sprintf("   Name: %s\n", pkg.Name)
		}
		if pkg.Description != "" {
			output += fmt.Sprintf("   Description: %s\n", strings.TrimSpace(pkg.Description))
		}
		if pkg.Version != "" {
			output += fmt.Sprintf("   Version: %s\n", pkg.Version)
		}
		if pkg.PackageType != "" {
			output += fmt.Sprintf("   Type: %s\n", pkg.PackageType)
		}
		if pkg.Tier != "" {
			output += fmt.Sprintf("   Tier: %s\n", pkg.Tier)
		}
		if pkg.FamilyRepoKey != "" {
			output += fmt.Sprintf("   Family: %s\n", pkg.FamilyRepoKey)
		}
		if pkg.PkgDigest != "" {
			output += fmt.Sprintf("   Digest: %s\n", pkg.PkgDigest)
		}
		if pkg.DownloadCount > 0 {
			output += fmt.Sprintf("   Downloads: %d\n", pkg.DownloadCount)
		}
		if !pkg.UpdatedAt.IsZero() {
			output += fmt.Sprintf("   Updated: %s\n", pkg.UpdatedAt.Format("2006-01-02 15:04:05"))
		}
		output += "\n"
	}

	if result.Count > len(result.Packages) {
		output += fmt.Sprintf("%d more matches are available; request the next page or a larger size (max 500).\n", result.Count-len(result.Packages))
	}

	return output
}

// formatPackageMetadata formats package metadata for display.
func formatPackageMetadata(metadata *marketplace.PackageMetadata) string {
	if metadata == nil {
		return "No package metadata"
	}

	output := fmt.Sprintf("Package: %s\n", metadata.RepoKey)
	output += "=====================================\n\n"

	if metadata.Type != "" {
		output += fmt.Sprintf("Type: %s\n", metadata.Type)
	}
	if metadata.Tier != "" {
		output += fmt.Sprintf("Tier: %s\n", metadata.Tier)
	}
	output += fmt.Sprintf("Public: %t\n", metadata.Public)
	if metadata.CurrentVersion != "" {
		output += fmt.Sprintf("Current Version: %s\n", metadata.CurrentVersion)
	}
	if metadata.FamilyRepoKey != "" {
		output += fmt.Sprintf("Family: %s\n", metadata.FamilyRepoKey)
	}
	if metadata.Digest != "" {
		output += fmt.Sprintf("Digest: %s\n", metadata.Digest)
	}
	if metadata.PublishedAt != "" {
		output += fmt.Sprintf("Published: %s\n", metadata.PublishedAt)
	}
	if metadata.EndOfSupport != "" {
		output += fmt.Sprintf("End of Support: %s\n", metadata.EndOfSupport)
	}
	if metadata.EndOfLife != "" {
		output += fmt.Sprintf("End of Life: %s\n", metadata.EndOfLife)
	}
	if metadata.Languages != "" {
		output += fmt.Sprintf("Languages: %s\n", metadata.Languages)
	}

	if len(metadata.Versions) > 0 {
		output += fmt.Sprintf("\nVersions (%d):\n", len(metadata.Versions))
		output += "-----------------------------------\n"
		for _, v := range metadata.Versions {
			output += fmt.Sprintf("- %s", v.Version)
			if v.Digest != "" {
				output += fmt.Sprintf(" (%s)", v.Digest)
			}
			output += "\n"
		}
	}

	if metadata.RelatedRepository != nil && metadata.RelatedRepository.RepoKey != "" {
		output += fmt.Sprintf("\nRelated Repository: %s", metadata.RelatedRepository.RepoKey)
		if metadata.RelatedRepository.SubscriptionTier != "" {
			output += fmt.Sprintf(" (subscription tier: %s)", metadata.RelatedRepository.SubscriptionTier)
		}
		output += "\n"
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
