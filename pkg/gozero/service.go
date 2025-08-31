package gozero

import (
	"context"
	"fmt"
	"strings"

	"github.com/joeblew999/infra/pkg/log"
)

// Service provides high-level go-zero operations following infra patterns
type Service struct {
	runner *GoZeroRunner
}

// NewService creates a new go-zero service wrapper
func NewService(debug bool) *Service {
	return &Service{
		runner: NewGoZeroRunner(debug),
	}
}

// CreateMCPAPI creates a go-zero API service for MCP integration
func (s *Service) CreateMCPAPI(ctx context.Context, packageName, description string, outputDir string) error {
	log.Info("Creating MCP API service", "package", packageName, "output", outputDir)
	
	// Generate API content for MCP integration
	apiContent := s.generateMCPAPIContent(packageName, description)
	
	// Use our wrapper to generate the service
	s.runner.SetWorkDir(outputDir)
	return s.runner.GenerateInfraAPI(ctx, packageName, apiContent, outputDir)
}

// CreateStandardAPI creates a basic go-zero API service
func (s *Service) CreateStandardAPI(ctx context.Context, packageName string, endpoints []Endpoint, outputDir string) error {
	log.Info("Creating standard API service", "package", packageName, "endpoints", len(endpoints))
	
	apiContent := s.generateStandardAPIContent(packageName, endpoints)
	
	s.runner.SetWorkDir(outputDir)
	return s.runner.GenerateInfraAPI(ctx, packageName, apiContent, outputDir)
}

// Endpoint represents an API endpoint
type Endpoint struct {
	Name        string
	Method      string // GET, POST, PUT, DELETE
	Path        string
	RequestType string
	ResponseType string
	Handler     string
}

// generateMCPAPIContent creates a standard MCP-compatible API definition
func (s *Service) generateMCPAPIContent(packageName, description string) string {
	// Convert hyphens to underscores for go-zero compatibility
	serviceName := strings.ReplaceAll(packageName, "-", "_")
	return fmt.Sprintf(`syntax = "v1"

info(
    title: "%s API"
    desc: "%s"
    author: "infra"
    version: "v1.0"
)

type McpRequest {
    Jsonrpc string ` + "`json:\"jsonrpc\"`" + `
    Method  string ` + "`json:\"method\"`" + `
    Id      int64  ` + "`json:\"id\"`" + `
    Params  string ` + "`json:\"params,optional\"`" + `
}

type McpResponse {
    Jsonrpc string ` + "`json:\"jsonrpc\"`" + `
    Id      int64  ` + "`json:\"id\"`" + `
    Result  string ` + "`json:\"result,optional\"`" + `
    Error   string ` + "`json:\"error,optional\"`" + `
}

type HealthResponse {
    Status  string ` + "`json:\"status\"`" + `
    Version string ` + "`json:\"version\"`" + `
    Ready   bool   ` + "`json:\"ready\"`" + `
}

service %s_api {
    @handler McpHandler
    post /mcp (McpRequest) returns (McpResponse)
    
    @handler HealthHandler
    get /health returns (HealthResponse)
}
`, packageName, description, serviceName)
}

// generateStandardAPIContent creates API content from endpoints
func (s *Service) generateStandardAPIContent(packageName string, endpoints []Endpoint) string {
	// Convert hyphens to underscores for go-zero compatibility
	serviceName := strings.ReplaceAll(packageName, "-", "_")
	api := fmt.Sprintf(`syntax = "v1"

info(
    title: "%s API"
    desc: "Generated API service"
    author: "infra"
    version: "v1.0"
)

`, packageName)

	// Add request/response types
	typeMap := make(map[string]bool)
	for _, endpoint := range endpoints {
		if endpoint.RequestType != "" && !typeMap[endpoint.RequestType] {
			api += fmt.Sprintf(`type %s {
    // TODO: Add request fields
}

`, endpoint.RequestType)
			typeMap[endpoint.RequestType] = true
		}
		
		if endpoint.ResponseType != "" && !typeMap[endpoint.ResponseType] {
			api += fmt.Sprintf(`type %s {
    // TODO: Add response fields  
}

`, endpoint.ResponseType)
			typeMap[endpoint.ResponseType] = true
		}
	}

	// Add service definition
	api += fmt.Sprintf(`service %s_api {
`, serviceName)

	for _, endpoint := range endpoints {
		method := endpoint.Method
		if method == "" {
			method = "get"
		}
		
		handler := endpoint.Handler
		if handler == "" {
			handler = endpoint.Name + "Handler"
		}
		
		api += fmt.Sprintf(`    @handler %s
    %s %s`, handler, method, endpoint.Path)
		
		if endpoint.RequestType != "" {
			api += fmt.Sprintf(" (%s)", endpoint.RequestType)
		}
		
		if endpoint.ResponseType != "" {
			api += fmt.Sprintf(" returns (%s)", endpoint.ResponseType)
		}
		
		api += "\n"
	}

	api += "}\n"
	return api
}

// GetProjectStructure returns the expected project structure after generation
func (s *Service) GetProjectStructure(packageName string) []string {
	return []string{
		packageName + ".api",
		packageName + ".go",
		packageName + ".json", // Swagger
		"go.mod",
		"go.sum",
		"etc/" + packageName + "_api.yaml",
		"internal/config/config.go",
		"internal/handler/",
		"internal/logic/", 
		"internal/svc/servicecontext.go",
		"internal/types/types.go",
	}
}