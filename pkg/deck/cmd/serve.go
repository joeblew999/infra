package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/joeblew999/infra/pkg/deck"
)

// ServeCmd starts a web server for real-time deck rendering
func ServeCmd(args []string) error {
	port := 8080
	if len(args) > 0 {
		if p, err := strconv.Atoi(args[0]); err == nil {
			port = p
		}
	}

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/render", handleRender)
	http.HandleFunc("/api/examples", handleExamples)
	http.HandleFunc("/api/examples/content", handleExampleContent)
	
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Starting deck server on http://localhost%s\n", addr)
	fmt.Printf("🎯 Features:\n")
	fmt.Printf("  • Live decksh editor with real-time SVG rendering\n")
	fmt.Printf("  • 411 example templates from DuBois & DeckViz collections\n")
	fmt.Printf("  • Browse examples by category\n")
	fmt.Printf("  • Auto-rendering with 1-second debounce\n")
	
	return http.ListenAndServe(addr, nil)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <title>🎯 Deck Live Editor with Examples</title>
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; 
            margin: 0; padding: 20px; background: #f8f9fa;
        }
        .header { margin-bottom: 20px; }
        .container { display: flex; height: 85vh; gap: 20px; }
        .sidebar { 
            width: 280px; background: white; border-radius: 8px; 
            box-shadow: 0 2px 8px rgba(0,0,0,0.1); overflow: hidden;
        }
        .main { flex: 1; display: flex; gap: 20px; }
        .editor, .output { 
            flex: 1; background: white; border-radius: 8px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1); display: flex; flex-direction: column;
        }
        .panel-header { 
            padding: 15px; background: #f8f9fa; border-bottom: 1px solid #e9ecef; 
            font-weight: 600; margin: 0;
        }
        textarea { 
            flex: 1; font-family: 'Monaco', 'Menlo', monospace; font-size: 14px; 
            padding: 15px; border: none; resize: none; outline: none;
        }
        .svg-container { 
            flex: 1; padding: 15px; overflow: auto; 
            border: 1px solid #e9ecef; margin: 15px; border-radius: 4px;
        }
        .controls { 
            padding: 15px; border-bottom: 1px solid #e9ecef; display: flex; gap: 10px; 
        }
        button { 
            padding: 8px 16px; border: 1px solid #007bff; background: #007bff; 
            color: white; border-radius: 4px; cursor: pointer; font-size: 14px;
        }
        button:hover { background: #0056b3; }
        button.secondary { background: white; color: #007bff; }
        button.secondary:hover { background: #f8f9fa; }
        
        .examples-list { max-height: calc(100vh - 200px); overflow-y: auto; }
        .category { border-bottom: 1px solid #e9ecef; }
        .category-header { 
            padding: 12px 15px; background: #f8f9fa; cursor: pointer; 
            display: flex; justify-content: space-between; align-items: center;
        }
        .category-header:hover { background: #e9ecef; }
        .category-content { display: none; }
        .category.open .category-content { display: block; }
        .example-item { 
            padding: 10px 15px; cursor: pointer; border-bottom: 1px solid #f8f9fa;
        }
        .example-item:hover { background: #f8f9fa; }
        .example-name { font-weight: 500; color: #007bff; }
        .example-desc { font-size: 12px; color: #6c757d; margin-top: 2px; }
        .loading { padding: 20px; text-align: center; color: #6c757d; }
        .error { padding: 20px; color: #dc3545; background: #f8d7da; margin: 15px; border-radius: 4px; }
    </style>
</head>
<body>
    <div class="header">
        <h1 style="margin: 0; color: #343a40;">🎯 Deck Live Editor</h1>
        <p style="margin: 5px 0 0 0; color: #6c757d;">Real-time decksh editor with 411 example templates</p>
    </div>
    
    <div class="container">
        <div class="sidebar">
            <div class="panel-header">📚 Examples</div>
            <div id="examples-list" class="examples-list">
                <div class="loading">Loading examples...</div>
            </div>
        </div>
        
        <div class="main">
            <div class="editor">
                <div class="panel-header">✏️ Decksh Code</div>
                <div class="controls">
                    <button onclick="render()">🎯 Render</button>
                    <button onclick="loadDefaultExample()" class="secondary">📄 Default</button>
                    <button onclick="clearEditor()" class="secondary">🗑️ Clear</button>
                </div>
                <textarea id="decksh" placeholder="Enter your decksh code here...">deck
slide "lightblue" "darkblue"
text "Welcome to Deck!" 50 30 4 "sans" "white"
circle 25 60 10 "yellow"
rect 75 60 20 15 "red" 0.8
text "Choose an example from the sidebar!" 50 80 2 "sans" "lightgray"
eslide
edeck</textarea>
            </div>
            <div class="output">
                <div class="panel-header">🎨 SVG Output</div>
                <div id="svg-output" class="svg-container"></div>
            </div>
        </div>
    </div>

    <script>
        let categories = [];
        
        // Load examples on startup
        async function loadExamples() {
            try {
                const response = await fetch('/api/examples');
                categories = await response.json();
                renderExamplesList();
            } catch (error) {
                document.getElementById('examples-list').innerHTML = 
                    '<div class="error">Failed to load examples. Make sure to run "go run fetch.go" first.</div>';
            }
        }
        
        function renderExamplesList() {
            const container = document.getElementById('examples-list');
            
            if (categories.length === 0) {
                container.innerHTML = '<div class="error">No examples found</div>';
                return;
            }
            
            container.innerHTML = categories.map(category => ` + "`" + `
                <div class="category" id="cat-${category.name.replace(/\s+/g, '-').toLowerCase()}">
                    <div class="category-header" onclick="toggleCategory('${category.name.replace(/\s+/g, '-').toLowerCase()}')">
                        <div>
                            <strong>${category.name}</strong><br>
                            <small>${category.count} examples</small>
                        </div>
                        <span class="toggle">▶</span>
                    </div>
                    <div class="category-content">
                        ${category.examples.slice(0, 8).map(example => ` + "`" + `
                            <div class="example-item" onclick="loadExample('${example.path}')">
                                <div class="example-name">${example.name}</div>
                                <div class="example-desc">${example.description}</div>
                            </div>
                        ` + "`" + `).join('')}
                        ${category.examples.length > 8 ? ` + "`" + `<div class="example-item" style="font-style: italic; color: #6c757d;">... and ${category.examples.length - 8} more</div>` + "`" + ` : ''}
                    </div>
                </div>
            ` + "`" + `).join('');
        }
        
        function toggleCategory(categoryId) {
            const category = document.getElementById('cat-' + categoryId);
            const toggle = category.querySelector('.toggle');
            
            category.classList.toggle('open');
            toggle.textContent = category.classList.contains('open') ? '▼' : '▶';
        }
        
        async function loadExample(path) {
            try {
                const response = await fetch('/api/examples/content?path=' + encodeURIComponent(path));
                const content = await response.text();
                document.getElementById('decksh').value = content;
                render();
            } catch (error) {
                alert('Failed to load example: ' + error);
            }
        }
        
        function render() {
            const decksh = document.getElementById('decksh').value;
            const output = document.getElementById('svg-output');
            
            output.innerHTML = '<div style="padding: 20px; color: #6c757d;">Rendering...</div>';
            
            fetch('/render', {
                method: 'POST',
                headers: { 'Content-Type': 'text/plain' },
                body: decksh
            })
            .then(response => response.text())
            .then(svg => {
                output.innerHTML = svg;
            })
            .catch(error => {
                output.innerHTML = '<div class="error">Render Error: ' + error + '</div>';
            });
        }
        
        function loadDefaultExample() {
            document.getElementById('decksh').value = ` + "`" + `deck
slide "white" "black"
text "Welcome to Deck!" 50 20 5 "sans" "darkblue"
text "Real-time decksh rendering" 50 35 3 "sans" "gray"

// Basic shapes
circle 20 60 8 "red" 0.7
rect 50 60 15 12 "blue" 0.7
circle 80 60 8 "green" 0.7

// Try editing this code!
text "Edit the code and see live updates!" 50 85 2 "sans" "darkred"
eslide
edeck` + "`" + `;
            render();
        }
        
        function clearEditor() {
            document.getElementById('decksh').value = '';
            document.getElementById('svg-output').innerHTML = '';
        }
        
        // Auto-render on startup and typing
        window.onload = function() {
            loadExamples();
            render();
        };
        
        let timeout;
        document.getElementById('decksh').addEventListener('input', function() {
            clearTimeout(timeout);
            timeout = setTimeout(render, 1000);
        });
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(tmpl))
}

func handleRender(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the decksh input
	body := make([]byte, r.ContentLength)
	_, err := r.Body.Read(body)
	if err != nil && err.Error() != "EOF" {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	deckshInput := string(body)
	
	// Create renderer
	renderer := deck.NewDefaultRenderer()
	opts := deck.DefaultRenderOptions()
	opts.Title = "Live Editor"

	// Convert to SVG
	svg, err := renderer.DeckshToSVG(deckshInput, opts)
	if err != nil {
		log.Printf("Render error: %v", err)
		http.Error(w, fmt.Sprintf("Render error: %v", err), http.StatusBadRequest)
		return
	}

	// Return SVG
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Write([]byte(svg))
}

func handleExamples(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	categories, err := GetExampleCategories()
	if err != nil {
		log.Printf("Failed to get examples: %v", err)
		http.Error(w, "Failed to load examples", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

func handleExampleContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Missing path parameter", http.StatusBadRequest)
		return
	}

	content, err := GetExampleContent(path)
	if err != nil {
		log.Printf("Failed to get example content: %v", err)
		http.Error(w, "Failed to load example", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(content))
}