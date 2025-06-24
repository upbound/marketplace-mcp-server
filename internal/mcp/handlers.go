package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/upbound/marketplace-mcp-server/internal/marketplace"
)

// handleSearchPackages handles the search_packages tool call
func (s *Server) handleSearchPackages(ctx context.Context, req *MCPRequest, args map[string]interface{}) error {
	searchParams := marketplace.SearchParams{}

	if query, ok := args["query"].(string); ok {
		searchParams.Query = query
	}
	if family, ok := args["family"].(string); ok {
		searchParams.Family = family
	}
	if packageType, ok := args["package_type"].(string); ok {
		searchParams.PackageType = packageType
	}
	if accountName, ok := args["account_name"].(string); ok {
		searchParams.AccountName = accountName
	}
	if tier, ok := args["tier"].(string); ok {
		searchParams.Tier = tier
	}
	if public, ok := args["public"].(bool); ok {
		searchParams.Public = &public
	}
	if size, ok := args["size"].(float64); ok {
		searchParams.Size = int(size)
	}
	if page, ok := args["page"].(float64); ok {
		searchParams.Page = int(page)
	}
	if useV1, ok := args["use_v1"].(bool); ok {
		searchParams.UseV1 = useV1
	}

	result, err := s.client.SearchPackages(ctx, searchParams)
	if err != nil {
		return s.sendError("search_failed", err.Error(), req.ID)
	}

	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: ToolCallResult{
			Content: []Content{
				{
					Type: "text",
					Text: s.formatSearchResults(result),
				},
			},
		},
	}

	return s.sendResponse(response)
}

// handleGetPackageMetadata handles the get_package_metadata tool call
func (s *Server) handleGetPackageMetadata(ctx context.Context, req *MCPRequest, args map[string]interface{}) error {
	account, ok := args["account"].(string)
	if !ok {
		return s.sendError("invalid_params", "Missing required parameter: account", req.ID)
	}

	repository, ok := args["repository"].(string)
	if !ok {
		return s.sendError("invalid_params", "Missing required parameter: repository", req.ID)
	}

	version := ""
	if v, ok := args["version"].(string); ok {
		version = v
	}

	useV1 := false
	if v, ok := args["use_v1"].(bool); ok {
		useV1 = v
	}

	result, err := s.client.GetPackageMetadata(ctx, account, repository, version, useV1)
	if err != nil {
		return s.sendError("metadata_failed", err.Error(), req.ID)
	}

	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: ToolCallResult{
			Content: []Content{
				{
					Type: "text",
					Text: s.formatPackageMetadata(result),
				},
			},
		},
	}

	return s.sendResponse(response)
}

// handleGetPackageAssets handles the get_package_assets tool call
func (s *Server) handleGetPackageAssets(ctx context.Context, req *MCPRequest, args map[string]interface{}) error {
	account, ok := args["account"].(string)
	if !ok {
		return s.sendError("invalid_params", "Missing required parameter: account", req.ID)
	}

	repository, ok := args["repository"].(string)
	if !ok {
		return s.sendError("invalid_params", "Missing required parameter: repository", req.ID)
	}

	version, ok := args["version"].(string)
	if !ok {
		return s.sendError("invalid_params", "Missing required parameter: version", req.ID)
	}

	assetType, ok := args["asset_type"].(string)
	if !ok {
		return s.sendError("invalid_params", "Missing required parameter: asset_type", req.ID)
	}

	result, err := s.client.GetPackageAssets(ctx, account, repository, version, assetType)
	if err != nil {
		return s.sendError("assets_failed", err.Error(), req.ID)
	}

	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: ToolCallResult{
			Content: []Content{
				{
					Type: "text",
					Text: s.formatAssetResponse(result, assetType),
				},
			},
		},
	}

	return s.sendResponse(response)
}

// handleGetRepositories handles the get_repositories tool call
func (s *Server) handleGetRepositories(ctx context.Context, req *MCPRequest, args map[string]interface{}) error {
	account, ok := args["account"].(string)
	if !ok {
		return s.sendError("invalid_params", "Missing required parameter: account", req.ID)
	}

	params := marketplace.RepositoryParams{}
	if filter, ok := args["filter"].(string); ok {
		params.Filter = filter
	}
	if size, ok := args["size"].(float64); ok {
		params.Size = int(size)
	}
	if page, ok := args["page"].(float64); ok {
		params.Page = int(page)
	}
	if useV1, ok := args["use_v1"].(bool); ok {
		params.UseV1 = useV1
	}

	result, err := s.client.GetRepositories(ctx, account, params)
	if err != nil {
		if strings.Contains(err.Error(), "authentication required") {
			return s.sendError("auth_required", "Authentication required for this endpoint. Please run 'up login' to authenticate with UP CLI.", req.ID)
		}
		return s.sendError("repositories_failed", err.Error(), req.ID)
	}

	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: ToolCallResult{
			Content: []Content{
				{
					Type: "text",
					Text: s.formatRepositoryResponse(result),
				},
			},
		},
	}

	return s.sendResponse(response)
}

// handleReloadAuth handles reloading authentication from UP CLI config
func (s *Server) handleReloadAuth(ctx context.Context, req *MCPRequest, args map[string]interface{}) error {
	// Reload server URL from UP CLI profile
	serverURL, err := s.authManager.GetCurrentServerURL()
	if err != nil {
		return s.sendError("auth_failed", fmt.Sprintf("Failed to load server URL from UP CLI profile: %v", err), req.ID)
	}
	s.client.SetBaseURL(serverURL)

	// Reload authentication from UP CLI config
	token, err := s.authManager.GetCurrentToken()
	if err != nil {
		return s.sendError("auth_failed", fmt.Sprintf("Failed to load authentication from UP CLI: %v", err), req.ID)
	}

	// Set token in marketplace client
	s.client.SetToken(token.AccessToken)

	// Get current profile info
	profile, err := s.authManager.GetCurrentProfile()
	if err != nil {
		return s.sendError("auth_failed", fmt.Sprintf("Failed to get current profile: %v", err), req.ID)
	}

	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: ToolCallResult{
			Content: []Content{
				{
					Type: "text",
					Text: fmt.Sprintf("Authentication and server configuration reloaded successfully!\nProfile: '%s' (%s)\nServer: %s",
						profile.ID, profile.Organization, serverURL),
				},
			},
		},
	}

	return s.sendResponse(response)
}

// handleResourcesList handles the resources/list request
func (s *Server) handleResourcesList(req *MCPRequest) error {
	resources := []Resource{
		{
			URI:         "marketplace://packages",
			Name:        "Marketplace Packages",
			Description: "Search and browse packages in the Upbound Marketplace",
			MimeType:    "application/json",
		},
		{
			URI:         "marketplace://repositories",
			Name:        "Marketplace Repositories",
			Description: "Browse repositories in the Upbound Marketplace",
			MimeType:    "application/json",
		},
	}

	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: ResourcesListResult{
			Resources: resources,
		},
	}

	return s.sendResponse(response)
}

// handleResourcesRead handles the resources/read request
func (s *Server) handleResourcesRead(ctx context.Context, req *MCPRequest) error {
	var params ResourceReadParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return s.sendError("invalid_params", "Invalid resource read parameters", req.ID)
	}

	switch params.URI {
	case "marketplace://packages":
		return s.handleReadPackages(ctx, req)
	case "marketplace://repositories":
		return s.handleReadRepositories(ctx, req)
	default:
		return s.sendError("unknown_resource", fmt.Sprintf("Unknown resource: %s", params.URI), req.ID)
	}
}

// handleReadPackages handles reading the packages resource
func (s *Server) handleReadPackages(ctx context.Context, req *MCPRequest) error {
	// Default search for popular packages
	searchParams := marketplace.SearchParams{
		Size:   20,
		Public: func() *bool { b := true; return &b }(),
		Tier:   "official",
	}

	result, err := s.client.SearchPackages(ctx, searchParams)
	if err != nil {
		return s.sendError("search_failed", err.Error(), req.ID)
	}

	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: ResourceReadResult{
			Contents: []Content{
				{
					Type: "text",
					Text: s.formatSearchResults(result),
				},
			},
		},
	}

	return s.sendResponse(response)
}

// handleReadRepositories handles reading the repositories resource
func (s *Server) handleReadRepositories(ctx context.Context, req *MCPRequest) error {
	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: ResourceReadResult{
			Contents: []Content{
				{
					Type: "text",
					Text: "To read repositories, use the get_repositories tool with a specific account name.",
				},
			},
		},
	}

	return s.sendResponse(response)
}
