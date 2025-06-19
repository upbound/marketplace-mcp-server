package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/upbound/marketplace-mcp-server/internal/mcp"
)

func main() {
	proxy, err := mcp.NewProxy("./bin/marketplace-mcp-server")
	if err != nil {
		log.Fatalf("failed to start MCP subprocess: %v", err)
	}

	r := chi.NewRouter()

	r.Post("/initialize", func(w http.ResponseWriter, r *http.Request) {
		result, err := proxy.Forward("initialize", map[string]interface{}{})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		var resp mcp.MCPResponse
		if err := json.Unmarshal(result, &resp); err != nil {
			log.Fatalf("failed to unmarshal mcp response %v", err)
		}

		if resp.Error != nil {
			log.Fatalf("error from server: %s", resp.Error.Message)
		}

		var initResult mcp.InitializeResult
		if err := json.Unmarshal(result, &initResult); err != nil {
			log.Fatalf("failed to parse initialize response: %v", err)
		}

		writeJSON(w, result)
	})

	srv := &http.Server{
		Addr:         ":8765",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("🚀 MCP proxy listening on http://localhost:8765")
	log.Fatal(srv.ListenAndServe())
}

func writeJSON(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write(data)
	if err != nil {
		log.Fatalf("error writing HTTP response data: %v", err)
	}
}
