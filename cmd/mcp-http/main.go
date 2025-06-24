package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mark3labs/mcp-go/server"
	"github.com/upbound/marketplace-mcp-server/internal/marketplace"
	"github.com/upbound/marketplace-mcp-server/internal/mcp"
)

func main() {
	// Create marketplace client
	client := marketplace.NewClient()

	// Create MCP server
	mcpServer := mcp.NewServer(client)

	// Create HTTP server using mcp-go framework with stateless mode
	httpServer := server.NewStreamableHTTPServer(
		mcpServer.GetMCPServer(),
		server.WithStateLess(true), // Enable stateless mode for simple HTTP API usage
	)

	// Setup signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal")
		cancel()

		// Shutdown HTTP server
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down HTTP server: %v", err)
		}
	}()

	// Start the HTTP server
	addr := ":8765"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	log.Printf("Starting Upbound Marketplace MCP HTTP Server on %s...", addr)
	log.Printf("MCP endpoint will be available at http://localhost%s/mcp", addr)

	if err := httpServer.Start(addr); err != nil {
		log.Fatalf("HTTP server failed to start: %v", err)
	}

	log.Println("Server stopped")
}
