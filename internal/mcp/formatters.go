package mcp

import (
	"fmt"
	"strings"

	"github.com/upbound/marketplace-mcp-server/internal/marketplace"
)

// formatSearchResults formats search results for display
func (s *Server) formatSearchResults(result *marketplace.SearchResponse) string {
	if len(result.Packages) == 0 {
		return "No packages found."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d packages:\n\n", result.Total))

	for _, pkg := range result.Packages {
		sb.WriteString(fmt.Sprintf("**%s/%s**\n", pkg.Account, pkg.Repository))
		if pkg.Description != "" {
			sb.WriteString(fmt.Sprintf("Description: %s\n", pkg.Description))
		}
		if pkg.Type != "" {
			sb.WriteString(fmt.Sprintf("Type: %s\n", pkg.Type))
		}
		if pkg.Tier != "" {
			sb.WriteString(fmt.Sprintf("Tier: %s\n", pkg.Tier))
		}
		if pkg.Version != "" {
			sb.WriteString(fmt.Sprintf("Version: %s\n", pkg.Version))
		}
		sb.WriteString(fmt.Sprintf("Public: %t\n", pkg.Public))
		if pkg.Stars > 0 {
			sb.WriteString(fmt.Sprintf("Stars: %d\n", pkg.Stars))
		}
		if pkg.Downloads > 0 {
			sb.WriteString(fmt.Sprintf("Downloads: %d\n", pkg.Downloads))
		}
		if len(pkg.Tags) > 0 {
			sb.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(pkg.Tags, ", ")))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// formatPackageMetadata formats package metadata for display
func (s *Server) formatPackageMetadata(metadata *marketplace.PackageMetadata) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s/%s\n\n", metadata.Account, metadata.Repository))

	if metadata.Description != "" {
		sb.WriteString(fmt.Sprintf("**Description:** %s\n\n", metadata.Description))
	}

	sb.WriteString(fmt.Sprintf("**Type:** %s\n", metadata.Type))
	sb.WriteString(fmt.Sprintf("**Public:** %t\n", metadata.Public))
	if metadata.Tier != "" {
		sb.WriteString(fmt.Sprintf("**Tier:** %s\n", metadata.Tier))
	}
	if metadata.License != "" {
		sb.WriteString(fmt.Sprintf("**License:** %s\n", metadata.License))
	}

	if metadata.LatestVersion != "" {
		sb.WriteString(fmt.Sprintf("**Latest Version:** %s\n", metadata.LatestVersion))
	}

	if len(metadata.Versions) > 0 {
		sb.WriteString(fmt.Sprintf("**Available Versions:** %s\n", strings.Join(metadata.Versions, ", ")))
	}

	if metadata.Homepage != "" {
		sb.WriteString(fmt.Sprintf("**Homepage:** %s\n", metadata.Homepage))
	}

	if metadata.Documentation != "" {
		sb.WriteString(fmt.Sprintf("**Documentation:** %s\n", metadata.Documentation))
	}

	if len(metadata.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("**Tags:** %s\n", strings.Join(metadata.Tags, ", ")))
	}

	if len(metadata.Keywords) > 0 {
		sb.WriteString(fmt.Sprintf("**Keywords:** %s\n", strings.Join(metadata.Keywords, ", ")))
	}

	if len(metadata.Dependencies) > 0 {
		sb.WriteString("\n## Dependencies\n")
		for _, dep := range metadata.Dependencies {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", dep.Name, dep.Version))
		}
	}

	if len(metadata.CRDs) > 0 {
		sb.WriteString("\n## Custom Resource Definitions (CRDs)\n")
		for _, crd := range metadata.CRDs {
			sb.WriteString(fmt.Sprintf("- **%s** (%s/%s)\n", crd.Kind, crd.Group, crd.Version))
			if crd.Description != "" {
				sb.WriteString(fmt.Sprintf("  Description: %s\n", crd.Description))
			}
		}
	}

	if len(metadata.Examples) > 0 {
		sb.WriteString("\n## Examples\n")
		for _, example := range metadata.Examples {
			sb.WriteString(fmt.Sprintf("### %s\n", example.Name))
			if example.Description != "" {
				sb.WriteString(fmt.Sprintf("%s\n", example.Description))
			}
			sb.WriteString(fmt.Sprintf("```%s\n%s\n```\n\n", example.Type, example.Content))
		}
	}

	if len(metadata.Compositions) > 0 {
		sb.WriteString("\n## Compositions\n")
		for _, comp := range metadata.Compositions {
			sb.WriteString(fmt.Sprintf("### %s\n", comp.Name))
			if comp.Description != "" {
				sb.WriteString(fmt.Sprintf("%s\n", comp.Description))
			}
			if len(comp.Resources) > 0 {
				sb.WriteString("Resources:\n")
				for _, res := range comp.Resources {
					sb.WriteString(fmt.Sprintf("- %s (%s)\n", res.Name, res.Type))
				}
			}
		}
	}

	if len(metadata.Functions) > 0 {
		sb.WriteString("\n## Functions\n")
		for _, fn := range metadata.Functions {
			sb.WriteString(fmt.Sprintf("### %s\n", fn.Name))
			if fn.Description != "" {
				sb.WriteString(fmt.Sprintf("%s\n", fn.Description))
			}
			sb.WriteString(fmt.Sprintf("Version: %s\n", fn.Version))
			sb.WriteString(fmt.Sprintf("Image: %s\n", fn.Image))
		}
	}

	return sb.String()
}

// formatAssetResponse formats asset response for display
func (s *Server) formatAssetResponse(asset *marketplace.AssetResponse, assetType string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s Asset\n\n", strings.Title(assetType)))

	if asset.URL != "" {
		sb.WriteString(fmt.Sprintf("**Asset URL:** %s\n\n", asset.URL))
		sb.WriteString("Use this URL to download the asset directly.\n\n")
	}

	if asset.Content != "" {
		sb.WriteString("**Content:**\n")
		sb.WriteString(fmt.Sprintf("```%s\n%s\n```\n", asset.Type, asset.Content))
	}

	// Add helpful information based on asset type
	switch assetType {
	case "crds":
		sb.WriteString("\n**About CRDs:**\n")
		sb.WriteString("Custom Resource Definitions define the schema for custom resources in Kubernetes.\n")
		sb.WriteString("These CRDs can be applied to your cluster to enable the resources provided by this package.\n")
	case "examples":
		sb.WriteString("\n**About Examples:**\n")
		sb.WriteString("These examples show how to use the resources provided by this package.\n")
		sb.WriteString("You can use these as templates for creating your own resources.\n")
	case "docs":
		sb.WriteString("\n**About Documentation:**\n")
		sb.WriteString("This documentation provides detailed information about the package and its usage.\n")
	case "package":
		sb.WriteString("\n**About Package:**\n")
		sb.WriteString("This is the complete package file that can be installed in your Crossplane cluster.\n")
	}

	return sb.String()
}

// formatRepositoryResponse formats repository response for display
func (s *Server) formatRepositoryResponse(result *marketplace.RepositoryResponse) string {
	if len(result.Repositories) == 0 {
		return "No repositories found."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d repositories:\n\n", result.Total))

	for _, repo := range result.Repositories {
		sb.WriteString(fmt.Sprintf("**%s/%s**\n", repo.Account, repo.Name))
		if repo.Description != "" {
			sb.WriteString(fmt.Sprintf("Description: %s\n", repo.Description))
		}
		if repo.Type != "" {
			sb.WriteString(fmt.Sprintf("Type: %s\n", repo.Type))
		}
		sb.WriteString(fmt.Sprintf("Public: %t\n", repo.Public))
		if repo.Policy != "" {
			sb.WriteString(fmt.Sprintf("Policy: %s\n", repo.Policy))
		}
		if repo.PackageCount > 0 {
			sb.WriteString(fmt.Sprintf("Packages: %d\n", repo.PackageCount))
		}
		if !repo.CreatedAt.IsZero() {
			sb.WriteString(fmt.Sprintf("Created: %s\n", repo.CreatedAt.Format("2006-01-02")))
		}
		if !repo.UpdatedAt.IsZero() {
			sb.WriteString(fmt.Sprintf("Updated: %s\n", repo.UpdatedAt.Format("2006-01-02")))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
