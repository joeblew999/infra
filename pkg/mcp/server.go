package mcp

import (
	"fmt"
	"log"
	"net/http"
)

// StartServer starts a simple HTTP server for MCP interactions.
func StartServer() error {
	http.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from infra MCP server!")
	})

	log.Println("MCP server starting on :8080")
	return http.ListenAndServe(":8080", nil)
}
