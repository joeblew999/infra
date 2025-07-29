package mcp

import (
	"fmt"
	"net/http"

	"github.com/joeblew999/infra/pkg/log"
)

// StartServer starts a simple HTTP server for MCP interactions.
func StartServer() error {
	http.Handle("/mcp/", http.StripPrefix("/mcp", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from infra MCP server! Path: %s", r.URL.Path)
	})))

	log.Info("MCP server starting", "port", 8080)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Error("MCP server failed to start", "error", err)
	}
	return err
}
