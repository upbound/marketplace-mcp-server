/*
Package main is the main entrypoint to the general server.
*/
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/upbound/marketplace-mcp-server/internal/marketplace"
	"github.com/upbound/marketplace-mcp-server/internal/mcp"
)

func main() {
	// Create marketplace client
	client := marketplace.NewClient()

	// Create MCP server
	server := mcp.NewServer(client)

	// Setup signal handling
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal")
		cancel()
	}()

	// Start the MCP server
	log.Println("Starting Upbound Marketplace MCP Server...")
	if err := server.Start(ctx); err != nil {
		cancel()
		log.Fatalf("Server failed to start: %v", err)
	}

	cancel()
	log.Println("Server stopped")
}
