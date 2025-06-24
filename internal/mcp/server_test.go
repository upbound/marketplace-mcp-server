package mcp

import (
	"testing"

	"github.com/upbound/marketplace-mcp-server/internal/marketplace"
)

func TestNewServer(t *testing.T) {
	client := marketplace.NewClient()
	server := NewServer(client)

	if server == nil {
		t.Fatal("NewServer() returned nil")
	}

	if server.client != client {
		t.Error("Server client should match the provided client")
	}

	if server.authManager == nil {
		t.Error("AuthManager should not be nil")
	}

	if server.mcpServer == nil {
		t.Error("MCP server should not be nil")
	}
}
