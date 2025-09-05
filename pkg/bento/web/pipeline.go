package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/starfederation/datastar-go/datastar"
	"gopkg.in/yaml.v3"

	"github.com/joeblew999/infra/pkg/log"
)

// PipelineComponent represents a component in the Bento pipeline
type PipelineComponent struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`        // input, processor, output
	Kind        string                 `json:"kind"`        // specific component like "generate", "stdout", etc
	Config      map[string]interface{} `json:"config"`
	Position    Position               `json:"position"`
	Connections []string               `json:"connections"` // connected component IDs
}

// Position represents the visual position of a component
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Pipeline represents a complete Bento pipeline configuration
type Pipeline struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Components  []*PipelineComponent `json:"components"`
	Created     time.Time            `json:"created"`
	Modified    time.Time            `json:"modified"`
	Status      string               `json:"status"` // draft, active, paused
}

// PipelineTemplate represents available component templates
type PipelineTemplate struct {
	Kind        string                 `json:"kind"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Icon        string                 `json:"icon"`
	Config      map[string]interface{} `json:"config"`
	Category    string                 `json:"category"`
}

// GetAvailableTemplates returns the available Bento component templates
func GetAvailableTemplates() []PipelineTemplate {
	return []PipelineTemplate{
		// Input templates
		{
			Kind:        "generate",
			Type:        "input",
			Name:        "Data Generator",
			Description: "Generate test data at intervals",
			Icon:        "⚡",
			Category:    "Source",
			Config: map[string]interface{}{
				"mapping":  `root = { "message": "hello world", "timestamp": now() }`,
				"interval": "5s",
			},
		},
		{
			Kind:        "stdin",
			Type:        "input",
			Name:        "Standard Input",
			Description: "Read from standard input",
			Icon:        "📥",
			Category:    "Source",
			Config:      map[string]interface{}{},
		},
		{
			Kind:        "http_server",
			Type:        "input",
			Name:        "HTTP Server",
			Description: "Receive data via HTTP requests",
			Icon:        "🌐",
			Category:    "Source",
			Config: map[string]interface{}{
				"address":      "0.0.0.0:8080",
				"path":         "/post",
				"allowed_verbs": []string{"POST"},
			},
		},
		{
			Kind:        "nats",
			Type:        "input",
			Name:        "NATS Subscribe",
			Description: "Subscribe to NATS messages",
			Icon:        "📡",
			Category:    "Source",
			Config: map[string]interface{}{
				"urls":    []string{"nats://localhost:4222"},
				"subject": "events.>",
			},
		},
		
		// Processor templates
		{
			Kind:        "mapping",
			Type:        "processor",
			Name:        "Data Mapping",
			Description: "Transform message content using Bloblang",
			Icon:        "🔄",
			Category:    "Transform",
			Config: map[string]interface{}{
				"mapping": `root.processed_at = now()`,
			},
		},
		{
			Kind:        "log",
			Type:        "processor",
			Name:        "Logger",
			Description: "Log messages for debugging",
			Icon:        "📝",
			Category:    "Debug",
			Config: map[string]interface{}{
				"level":   "INFO",
				"message": "Processing: ${! content() }",
			},
		},
		{
			Kind:        "filter",
			Type:        "processor",
			Name:        "Message Filter",
			Description: "Filter messages based on conditions",
			Icon:        "🔍",
			Category:    "Transform",
			Config: map[string]interface{}{
				"condition": `content().length() > 0`,
			},
		},
		{
			Kind:        "rate_limit",
			Type:        "processor",
			Name:        "Rate Limiter",
			Description: "Control message throughput",
			Icon:        "⏱️",
			Category:    "Control",
			Config: map[string]interface{}{
				"resource": "global_limit",
			},
		},
		
		// Output templates
		{
			Kind:        "stdout",
			Type:        "output",
			Name:        "Standard Output",
			Description: "Print messages to console",
			Icon:        "📤",
			Category:    "Sink",
			Config:      map[string]interface{}{},
		},
		{
			Kind:        "http_client",
			Type:        "output",
			Name:        "HTTP Client",
			Description: "Send data via HTTP requests",
			Icon:        "🌐",
			Category:    "Sink",
			Config: map[string]interface{}{
				"url":    "http://localhost:8080/webhook",
				"verb":   "POST",
				"headers": map[string]string{
					"Content-Type": "application/json",
				},
			},
		},
		{
			Kind:        "nats",
			Type:        "output",
			Name:        "NATS Publish",
			Description: "Publish messages to NATS",
			Icon:        "📡",
			Category:    "Sink",
			Config: map[string]interface{}{
				"urls":    []string{"nats://localhost:4222"},
				"subject": "processed.messages",
			},
		},
		{
			Kind:        "file",
			Type:        "output",
			Name:        "File Writer",
			Description: "Write messages to files",
			Icon:        "💾",
			Category:    "Sink",
			Config: map[string]interface{}{
				"path": "./output/${! timestamp_unix() }.json",
			},
		},
	}
}

// HandleGetTemplates returns available pipeline component templates
func HandleGetTemplates(w http.ResponseWriter, r *http.Request) {
	templates := GetAvailableTemplates()
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(templates); err != nil {
		log.Error("Error encoding templates", "error", err)
		http.Error(w, "Failed to encode templates", http.StatusInternalServerError)
	}
}

// HandlePipelineBuilder streams the pipeline builder interface
func HandlePipelineBuilder(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)
	
	// Get available templates
	templates := GetAvailableTemplates()
	
	// Group templates by category
	categories := make(map[string][]PipelineTemplate)
	for _, tmpl := range templates {
		categories[tmpl.Category] = append(categories[tmpl.Category], tmpl)
	}
	
	// Render the component palette
	paletteHTML := renderComponentPalette(categories)
	if err := sse.PatchElements(fmt.Sprintf(`<div id="component-palette">%s</div>`, paletteHTML)); err != nil {
		log.Error("Error sending component palette", "error", err)
		return
	}
	
	// Send initial empty canvas
	canvasHTML := `
		<div class="flex-1 bg-gray-100 dark:bg-gray-800 rounded-lg border-2 border-dashed border-gray-300 dark:border-gray-600 min-h-96 relative" 
		     id="pipeline-canvas" 
		     data-pipeline-canvas="true">
			<div class="absolute inset-0 flex items-center justify-center text-gray-500 dark:text-gray-400">
				<div class="text-center">
					<div class="text-4xl mb-2">🎯</div>
					<div class="text-lg font-medium">Drop components here</div>
					<div class="text-sm">Drag from the component palette to build your pipeline</div>
				</div>
			</div>
		</div>
	`
	
	if err := sse.PatchElements(fmt.Sprintf(`<div id="pipeline-canvas-container">%s</div>`, canvasHTML)); err != nil {
		log.Error("Error sending pipeline canvas", "error", err)
		return
	}
	
	// Keep connection alive for real-time updates
	<-r.Context().Done()
}

// renderComponentPalette generates HTML for the component palette
func renderComponentPalette(categories map[string][]PipelineTemplate) string {
	var html strings.Builder
	
	html.WriteString(`<div class="space-y-4">`)
	html.WriteString(`<h3 class="text-lg font-semibold text-gray-900 dark:text-white">Component Palette</h3>`)
	
	categoryOrder := []string{"Source", "Transform", "Control", "Debug", "Sink"}
	
	for _, category := range categoryOrder {
		templates, exists := categories[category]
		if !exists {
			continue
		}
		
		html.WriteString(fmt.Sprintf(`
			<div class="mb-4">
				<h4 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">%s</h4>
				<div class="space-y-2">
		`, category))
		
		for _, tmpl := range templates {
			html.WriteString(fmt.Sprintf(`
				<div class="p-3 bg-white dark:bg-gray-700 rounded-lg border border-gray-200 dark:border-gray-600 cursor-move hover:shadow-md transition-shadow"
				     draggable="true"
				     data-component-kind="%s"
				     data-component-type="%s"
				     data-component-name="%s">
					<div class="flex items-center space-x-2">
						<span class="text-lg">%s</span>
						<div>
							<div class="text-sm font-medium text-gray-900 dark:text-white">%s</div>
							<div class="text-xs text-gray-600 dark:text-gray-400">%s</div>
						</div>
					</div>
				</div>
			`, tmpl.Kind, tmpl.Type, tmpl.Name, tmpl.Icon, tmpl.Name, tmpl.Description))
		}
		
		html.WriteString(`</div></div>`)
	}
	
	html.WriteString(`</div>`)
	return html.String()
}

// HandlePipelineValidate validates a pipeline configuration
func HandlePipelineValidate(w http.ResponseWriter, r *http.Request) {
	var pipeline Pipeline
	if err := json.NewDecoder(r.Body).Decode(&pipeline); err != nil {
		http.Error(w, "Invalid pipeline JSON", http.StatusBadRequest)
		return
	}
	
	// Perform basic validation
	errors := validatePipeline(&pipeline)
	
	response := map[string]interface{}{
		"valid":  len(errors) == 0,
		"errors": errors,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandlePipelineExport exports a pipeline as YAML
func HandlePipelineExport(w http.ResponseWriter, r *http.Request) {
	var pipeline Pipeline
	if err := json.NewDecoder(r.Body).Decode(&pipeline); err != nil {
		http.Error(w, "Invalid pipeline JSON", http.StatusBadRequest)
		return
	}
	
	// Convert to Bento YAML format
	bentoConfig := pipelineToYAML(&pipeline)
	
	yamlBytes, err := yaml.Marshal(bentoConfig)
	if err != nil {
		log.Error("Error marshaling pipeline to YAML", "error", err)
		http.Error(w, "Failed to export pipeline", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/yaml")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.yaml\"", pipeline.Name))
	w.Write(yamlBytes)
}

// validatePipeline performs basic pipeline validation
func validatePipeline(pipeline *Pipeline) []string {
	var errors []string
	
	if pipeline.Name == "" {
		errors = append(errors, "Pipeline name is required")
	}
	
	hasInput := false
	hasOutput := false
	
	for _, component := range pipeline.Components {
		if component.Type == "input" {
			hasInput = true
		}
		if component.Type == "output" {
			hasOutput = true
		}
	}
	
	if !hasInput {
		errors = append(errors, "Pipeline must have at least one input component")
	}
	
	if !hasOutput {
		errors = append(errors, "Pipeline must have at least one output component")
	}
	
	return errors
}

// pipelineToYAML converts a Pipeline struct to Bento YAML configuration
func pipelineToYAML(pipeline *Pipeline) map[string]interface{} {
	config := make(map[string]interface{})
	
	// Find input, processors, and output components
	var input *PipelineComponent
	var processors []*PipelineComponent
	var output *PipelineComponent
	
	for _, component := range pipeline.Components {
		switch component.Type {
		case "input":
			if input == nil { // Use the first input found
				input = component
			}
		case "processor":
			processors = append(processors, component)
		case "output":
			if output == nil { // Use the first output found
				output = component
			}
		}
	}
	
	// Set input configuration
	if input != nil {
		config["input"] = map[string]interface{}{
			input.Kind: input.Config,
		}
	}
	
	// Set processor configurations
	if len(processors) > 0 {
		if len(processors) == 1 {
			config["pipeline"] = map[string]interface{}{
				"processors": []map[string]interface{}{
					{processors[0].Kind: processors[0].Config},
				},
			}
		} else {
			processorConfigs := make([]map[string]interface{}, len(processors))
			for i, proc := range processors {
				processorConfigs[i] = map[string]interface{}{
					proc.Kind: proc.Config,
				}
			}
			config["pipeline"] = map[string]interface{}{
				"processors": processorConfigs,
			}
		}
	}
	
	// Set output configuration
	if output != nil {
		config["output"] = map[string]interface{}{
			output.Kind: output.Config,
		}
	}
	
	// Add HTTP server configuration
	config["http"] = map[string]interface{}{
		"address": "0.0.0.0:4195",
		"enabled": true,
	}
	
	return config
}