package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/upbound/marketplace-mcp-server/internal/auth"
	"github.com/upbound/marketplace-mcp-server/internal/marketplace"
)

// Server represents the MCP server
type Server struct {
	client      *marketplace.Client
	authManager *auth.Manager
	reader      *bufio.Reader
	writer      io.Writer
}

// NewServer creates a new MCP server
func NewServer(client *marketplace.Client) *Server {
	// Initialize auth manager
	authManager := auth.NewManager()

	return &Server{
		client:      client,
		authManager: authManager,
		reader:      bufio.NewReader(os.Stdin),
		writer:      os.Stdout,
	}
}

// Start starts the MCP server
func (s *Server) Start(ctx context.Context) error {
	log.Println("MCP Server started, waiting for requests...")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := s.handleRequest(ctx); err != nil {
				if err == io.EOF {
					return nil
				}
				log.Printf("Error handling request: %v", err)
				if err == io.EOF {
					log.Println("EOF received, stopping")	    
		}
	}
}

// handleRequest handles a single MCP request
func (s *Server) handleRequest(ctx context.Context) error {
	line, err := s.reader.ReadString('\n')
	if err != nil {
		return err
	}

	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	var request MCPRequest
	if err := json.Unmarshal([]byte(line), &request); err != nil {
		return s.sendError("parse_error", "Invalid JSON", nil)
	}

	return s.processRequest(ctx, &request)
}

// processRequest processes an MCP request
func (s *Server) processRequest(ctx context.Context, req *MCPRequest) error {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(ctx, req)
	case "resources/list":
		return s.handleResourcesList(req)
	case "resources/read":
		return s.handleResourcesRead(ctx, req)
	default:
		return s.sendError("method_not_found", fmt.Sprintf("Unknown method: %s", req.Method), req.ID)
	}
}

// handleInitialize handles the initialize request
func (s *Server) handleInitialize(req *MCPRequest) error {
	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: InitializeResult{
			ProtocolVersion: "2024-11-05",
			Capabilities: ServerCapabilities{
				Tools: &ToolsCapability{
					ListChanged: false,
				},
				Resources: &ResourcesCapability{
					Subscribe:   false,
					ListChanged: false,
				},
			},
			ServerInfo: ServerInfo{
				Name:    "marketplace-mcp-server",
				Version: "1.0.0",
			},
		},
	}

	return s.sendResponse(response)
}

// handleToolsList handles the tools/list request
func (s *Server) handleToolsList(req *MCPRequest) error {
	tools := []Tool{
		{
			Name:        "search_packages",
			Description: "Search for packages in the Upbound Marketplace",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query for packages",
					},
					"family": map[string]interface{}{
						"type":        "string",
						"description": "Family repository key to filter by",
					},
					"package_type": map[string]interface{}{
						"type":        "string",
						"description": "Type of package (provider, configuration, function)",
					},
					"account_name": map[string]interface{}{
						"type":        "string",
						"description": "Account/organization name to filter by",
					},
					"tier": map[string]interface{}{
						"type":        "string",
						"description": "Package tier (official, community, etc.)",
					},
					"public": map[string]interface{}{
						"type":        "boolean",
						"description": "Filter by public/private packages",
					},
					"size": map[string]interface{}{
						"type":        "integer",
						"description": "Number of results to return (max 500)",
						"default":     20,
					},
					"page": map[string]interface{}{
						"type":        "integer",
						"description": "Page number (0-indexed)",
						"default":     0,
					},
					"use_v1": map[string]interface{}{
						"type":        "boolean",
						"description": "Use v1 API instead of v2",
						"default":     false,
					},
				},
			},
		},
		{
			Name:        "get_package_metadata",
			Description: "Get detailed metadata for a specific package",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"account": map[string]interface{}{
						"type":        "string",
						"description": "Account/organization name",
					},
					"repository": map[string]interface{}{
						"type":        "string",
						"description": "Repository name",
					},
					"version": map[string]interface{}{
						"type":        "string",
						"description": "Package version (optional, gets latest if not specified)",
					},
					"use_v1": map[string]interface{}{
						"type":        "boolean",
						"description": "Use v1 API instead of v2",
						"default":     false,
					},
				},
				"required": []string{"account", "repository"},
			},
		},
		{
			Name:        "get_package_assets",
			Description: "Get assets (documentation, icons, release notes, etc.) for a specific package version",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"account": map[string]interface{}{
						"type":        "string",
						"description": "Account/organization name",
					},
					"repository": map[string]interface{}{
						"type":        "string",
						"description": "Repository name",
					},
					"version": map[string]interface{}{
						"type":        "string",
						"description": "Package version or 'latest'",
					},
					"asset_type": map[string]interface{}{
						"type":        "string",
						"description": "Type of asset to retrieve",
						"enum":        []string{"docs", "icon", "readme", "releaseNotes", "sbom"},
					},
				},
				"required": []string{"account", "repository", "version", "asset_type"},
			},
		},
		{
			Name:        "get_repositories",
			Description: "Get repositories for an account",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"account": map[string]interface{}{
						"type":        "string",
						"description": "Account/organization name",
					},
					"filter": map[string]interface{}{
						"type":        "string",
						"description": "AIP-160 formatted filter (v2 only)",
					},
					"size": map[string]interface{}{
						"type":        "integer",
						"description": "Number of results to return (max 100)",
						"default":     20,
					},
					"page": map[string]interface{}{
						"type":        "integer",
						"description": "Page number (0-indexed)",
						"default":     0,
					},
					"use_v1": map[string]interface{}{
						"type":        "boolean",
						"description": "Use v1 API instead of v2",
						"default":     false,
					},
				},
				"required": []string{"account"},
			},
		},
		{
			Name:        "authenticate",
			Description: "Authenticate with Upbound using callback-based TOTP flow to access private resources",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}

	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: ToolsListResult{
			Tools: tools,
		},
	}

	return s.sendResponse(response)
}

// handleToolsCall handles the tools/call request
func (s *Server) handleToolsCall(ctx context.Context, req *MCPRequest) error {
	var params ToolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return s.sendError("invalid_params", "Invalid tool call parameters", req.ID)
	}

	switch params.Name {
	case "search_packages":
		return s.handleSearchPackages(ctx, req, params.Arguments)
	case "get_package_metadata":
		return s.handleGetPackageMetadata(ctx, req, params.Arguments)
	case "get_package_assets":
		return s.handleGetPackageAssets(ctx, req, params.Arguments)
	case "get_repositories":
		return s.handleGetRepositories(ctx, req, params.Arguments)
	case "authenticate":
		return s.handleAuthenticate(ctx, req, params.Arguments)
	default:
		return s.sendError("unknown_tool", fmt.Sprintf("Unknown tool: %s", params.Name), req.ID)
	}
}

// sendResponse sends an MCP response
func (s *Server) sendResponse(response MCPResponse) error {
	data, err := json.Marshal(response)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(s.writer, "%s\n", data)
	return err
}

// sendError sends an MCP error response
func (s *Server) sendError(code, message string, id interface{}) error {
	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
		},
	}

	return s.sendResponse(response)
}
