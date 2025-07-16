package mcp

import (
	"context"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/upbound/marketplace-mcp-server/internal/auth"
	"github.com/upbound/marketplace-mcp-server/internal/marketplace"
)

// Server represents the MCP server.
type Server struct {
	mcpServer   *server.MCPServer
	client      *marketplace.Client
	authManager *auth.Manager
}

// NewServer creates a new MCP server using mcp-go framework.
func NewServer(client *marketplace.Client) *Server {
	// Initialize auth manager
	authManager := auth.NewManager()

	// Try to load and set server URL from UP CLI profile
	if serverURL, err := authManager.GetCurrentServerURL(); err == nil {
		client.SetBaseURL(serverURL)
		log.Printf("Loaded server URL from UP CLI profile: %s", serverURL)
	} else {
		log.Printf("Warning: Failed to load server URL from UP CLI profile: %v", err)
	}

	// Try to load authentication token from UP CLI config
	if token, err := authManager.GetCurrentToken(); err == nil {
		client.SetToken(token.AccessToken)
		log.Printf("Loaded authentication token from UP CLI profile")
	} else {
		log.Printf("Warning: Failed to load authentication from UP CLI: %v", err)
	}

	s := &Server{
		client:      client,
		authManager: authManager,
	}

	// Create MCP server with server info
	mcpServer := server.NewMCPServer(
		"marketplace-mcp-server",
		"1.0.0",
	)

	s.mcpServer = mcpServer

	// Register tools
	s.registerTools()

	return s
}

// Start starts the MCP server using stdio transport.
func (s *Server) Start(_ context.Context) error {
	return server.ServeStdio(s.mcpServer)
}

// GetMCPServer returns the underlying MCP server for use with other transports.
func (s *Server) GetMCPServer() *server.MCPServer {
	return s.mcpServer
}

// registerTools registers all available tools.
func (s *Server) registerTools() {
	// Search packages tool
	s.mcpServer.AddTool(mcp.Tool{
		Name:        "search_packages",
		Description: "Search for packages in the Upbound Marketplace",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "Search query for packages",
				},
				"family": map[string]any{
					"type":        "string",
					"description": "Family repository key to filter by",
				},
				"package_type": map[string]any{
					"type":        "string",
					"description": "Type of package (provider, configuration, function)",
				},
				"account_name": map[string]any{
					"type":        "string",
					"description": "Account/organization name to filter by",
				},
				"tier": map[string]any{
					"type":        "string",
					"description": "Package tier (official, community, etc.)",
				},
				"public": map[string]any{
					"type":        "boolean",
					"description": "Filter by public/private packages",
				},
				"size": map[string]any{
					"type":        "integer",
					"description": "Number of results to return (max 500)",
					"default":     20,
				},
				"page": map[string]any{
					"type":        "integer",
					"description": "Page number (0-indexed)",
					"default":     0,
				},
				"use_v1": map[string]any{
					"type":        "boolean",
					"description": "Use v1 API instead of v2",
					"default":     false,
				},
			},
		},
	}, s.handleSearchPackages)

	// Get package metadata tool
	s.mcpServer.AddTool(mcp.Tool{
		Name:        "get_package_metadata",
		Description: "Get detailed metadata for a specific package",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"account": map[string]any{
					"type":        "string",
					"description": "Account/organization name",
				},
				"repository": map[string]any{
					"type":        "string",
					"description": "Repository name",
				},
				"version": map[string]any{
					"type":        "string",
					"description": "Package version (optional, gets latest if not specified)",
				},
				"use_v1": map[string]any{
					"type":        "boolean",
					"description": "Use v1 API instead of v2",
					"default":     false,
				},
			},
			Required: []string{"account", "repository"},
		},
	}, s.handleGetPackageMetadata)

	// Get package assets tool
	s.mcpServer.AddTool(mcp.Tool{
		Name:        "get_package_assets",
		Description: "Get assets (documentation, icons, release notes, etc.) for a specific package version",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"account": map[string]any{
					"type":        "string",
					"description": "Account/organization name",
				},
				"repository": map[string]any{
					"type":        "string",
					"description": "Repository name",
				},
				"version": map[string]any{
					"type":        "string",
					"description": "Package version or 'latest'",
				},
				"asset_type": map[string]any{
					"type":        "string",
					"description": "Type of asset to retrieve",
					"enum":        []string{"docs", "icon", "readme", "releaseNotes", "sbom"},
				},
			},
			Required: []string{"account", "repository", "version", "asset_type"},
		},
	}, s.handleGetPackageAssets)

	// Get repositories tool
	s.mcpServer.AddTool(mcp.Tool{
		Name:        "get_repositories",
		Description: "Get repositories for an account",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"account": map[string]any{
					"type":        "string",
					"description": "Account/organization name",
				},
				"filter": map[string]any{
					"type":        "string",
					"description": "AIP-160 formatted filter (v2 only)",
				},
				"size": map[string]any{
					"type":        "integer",
					"description": "Number of results to return (max 100)",
					"default":     20,
				},
				"page": map[string]any{
					"type":        "integer",
					"description": "Page number (0-indexed)",
					"default":     0,
				},
				"use_v1": map[string]any{
					"type":        "boolean",
					"description": "Use v1 API instead of v2",
					"default":     false,
				},
			},
			Required: []string{"account"},
		},
	}, s.handleGetRepositories)

	// Get Package Version Resources tool
	s.mcpServer.AddTool(mcp.Tool{
		Name:        "get_package_version_resources",
		Description: "Get package version resources for a supplied repository name.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"account": map[string]any{
					"type":        "string",
					"description": "Account/organization name. For example upbound.",
				},
				"repository_name": map[string]any{
					"type":        "string",
					"description": "The name of the repository. For example provider-aws-s3.",
				},
				"version": map[string]any{
					"type":        "string",
					"description": "The version of the package. For example v1.23.1.",
				},
			},
			Required: []string{"account", "repository_name", "version"},
		},
	}, s.handleGetPackagesAccountRepositoryVersionResources)

	// Get Package Version Compositions Resources for Group & Kind tool
	s.mcpServer.AddTool(mcp.Tool{
		Name:        "get_package_version_composition_resources",
		Description: "Get package version composition resources for a supplied group, kind and version and composition.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"account": map[string]any{
					"type":        "string",
					"description": "Account/organization name. For example upbound.",
				},
				"repository_name": map[string]any{
					"type":        "string",
					"description": "The name of the repository. For example provider-aws-s3.",
				},
				"version": map[string]any{
					"type":        "string",
					"description": "The version of the package. For example v1.23.1.",
				},
				"resource_group": map[string]any{
					"type":        "string",
					"description": "The group of the resource. For example s3.aws.upbound.io.",
				},
				"resource_kind": map[string]any{
					"type":        "string",
					"description": "The kind of the resource. For example Bucket.",
				},
				"composition_name": map[string]any{
					"type":        "string",
					"description": "The name of the composition.",
				},
			},
			Required: []string{"account", "repository_name", "version", "resource_group", "resource_kind", "composition_name"},
		},
	}, s.handleGetPackagesAccountRepositoryVersionResourcesGroupKindComposition)

	// Get Package Version Resources for Group & Kind tool
	s.mcpServer.AddTool(mcp.Tool{
		Name:        "get_package_version_groupkind_resources",
		Description: "Get package version resources for a supplied group, kind and version.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"account": map[string]any{
					"type":        "string",
					"description": "Account/organization name. For example upbound.",
				},
				"repository_name": map[string]any{
					"type":        "string",
					"description": "The name of the repository. For example provider-aws-s3.",
				},
				"version": map[string]any{
					"type":        "string",
					"description": "The version of the package. For example v1.23.1.",
				},
				"resource_group": map[string]any{
					"type":        "string",
					"description": "The group of the resource. For example s3.aws.upbound.io.",
				},
				"resource_kind": map[string]any{
					"type":        "string",
					"description": "The kind of the resource. For example Bucket.",
				},
			},
			Required: []string{"account", "repository_name", "version", "resource_group", "resource_kind"},
		},
	}, s.handleGetPackagesAccountRepositoryVersionResourcesGroupKind)

	// Get specific package examples for account / repo / version / group and kind.
	s.mcpServer.AddTool(mcp.Tool{
		Name:        "get_package_version_examples",
		Description: "Get package version examples for a supplied group, kind and version.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"account": map[string]any{
					"type":        "string",
					"description": "Account/organization name. For example upbound.",
				},
				"repository_name": map[string]any{
					"type":        "string",
					"description": "The name of the repository. For example provider-aws-s3.",
				},
				"version": map[string]any{
					"type":        "string",
					"description": "The version of the package. For example v1.23.1.",
				},
				"resource_group": map[string]any{
					"type":        "string",
					"description": "The group of the resource. For example s3.aws.upbound.io.",
				},
				"resource_kind": map[string]any{
					"type":        "string",
					"description": "The kind of the resource. For example Bucket.",
				},
			},
			Required: []string{"account", "repository_name", "version", "resource_group", "resource_kind"},
		},
	}, s.handleGetPackagesAccountRepositoryVersionResourcesGroupKindExamples)

	// Reload auth tool
	s.mcpServer.AddTool(mcp.Tool{
		Name:        "reload_auth",
		Description: "Reload authentication and server configuration from UP CLI configuration (useful if you switched profiles)",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"random_string": map[string]any{
					"type":        "string",
					"description": "Dummy parameter for no-parameter tools",
				},
			},
			Required: []string{"random_string"},
		},
	}, s.handleReloadAuth)
}
