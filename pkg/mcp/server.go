package mcp

import (
	"fmt"
	"log"
	"net/http"
)

// StartServer starts a simple HTTP server for MCP interactions.
func StartServer() error {
	http.Handle("/mcp/", http.StripPrefix("/mcp", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from infra MCP server! Path: %s", r.URL.Path)
	})))

	log.Println("MCP server starting on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Printf("MCP server failed to start: %v", err)
	}
	return err
}
