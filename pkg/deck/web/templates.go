package web

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Deck Examples Viewer</title>
    <script src="https://cdn.jsdelivr.net/npm/@sudodevnull/datastar@latest/dist/datastar.js"></script>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
            background: #f5f5f5;
            height: 100vh;
            display: flex;
        }
        
        .sidebar {
            width: 300px;
            background: white;
            border-right: 1px solid #e0e0e0;
            display: flex;
            flex-direction: column;
        }
        
        .sidebar-header {
            padding: 1rem;
            border-bottom: 1px solid #e0e0e0;
            background: #fafafa;
        }
        
        .examples-list {
            flex: 1;
            overflow-y: auto;
        }
        
        .example-item {
            padding: 0.75rem 1rem;
            border-bottom: 1px solid #f0f0f0;
            cursor: pointer;
            transition: background-color 0.2s;
        }
        
        .example-item:hover {
            background: #f8f9fa;
        }
        
        .example-item.active {
            background: #e3f2fd;
            border-left: 3px solid #2196f3;
        }
        
        .main-content {
            flex: 1;
            display: flex;
            flex-direction: column;
            background: white;
        }
        
        .content-header {
            padding: 1rem;
            border-bottom: 1px solid #e0e0e0;
            background: #fafafa;
        }
        
        .tabs {
            display: flex;
            border-bottom: 1px solid #e0e0e0;
        }
        
        .tab {
            padding: 0.75rem 1.5rem;
            background: #f8f9fa;
            border: none;
            cursor: pointer;
            border-bottom: 2px solid transparent;
            font-size: 0.9rem;
        }
        
        .tab:hover {
            background: #e9ecef;
        }
        
        .tab.active {
            background: white;
            border-bottom-color: #2196f3;
        }
        
        .content-area {
            flex: 1;
            padding: 1rem;
            overflow: auto;
        }
        
        .loading {
            display: flex;
            justify-content: center;
            align-items: center;
            height: 200px;
            color: #666;
        }
        
        .error {
            color: #d32f2f;
            padding: 1rem;
            background: #ffebee;
            border-radius: 4px;
            margin: 1rem 0;
        }
        
        .output-frame {
            width: 100%;
            height: 600px;
            border: 1px solid #e0e0e0;
            border-radius: 4px;
        }
        
        .empty-state {
            text-align: center;
            color: #666;
            padding: 2rem;
        }
    </style>
</head>
<body data-load="/api/examples" data-target="#examples-container">
    
    <div class="sidebar">
        <div class="sidebar-header">
            <h2>Deck Examples</h2>
            <p>Click an example to generate outputs</p>
        </div>
        
        <div class="examples-list" id="examples-container">
            <div class="loading">Loading examples...</div>
        </div>
    </div>
    
    <div class="main-content">
        <div class="content-header">
            <h1 id="current-example">Select an Example</h1>
            <p>Choose an example from the sidebar to see SVG, PNG, and PDF outputs</p>
        </div>
        
        <div class="tabs">
            <button class="tab active" data-on-click="showTab('svg')">SVG</button>
            <button class="tab" data-on-click="showTab('png')">PNG</button>
            <button class="tab" data-on-click="showTab('pdf')">PDF</button>
        </div>
        
        <div class="content-area">
            <div id="output-container">
                <div class="empty-state">
                    <h3>Welcome to Deck Examples Viewer</h3>
                    <p>Select an example from the sidebar to generate and view SVG, PNG, and PDF outputs.</p>
                </div>
            </div>
        </div>
    </div>

    <script>
        let currentExample = null;
        let currentTab = 'svg';
        let generatedOutputs = {};
        
        // Handle examples list loading
        window.addEventListener('datastar:load', (e) => {
            if (e.detail.url.includes('/api/examples')) {
                const examples = e.detail.data;
                const container = document.getElementById('examples-container');
                
                if (examples && examples.length > 0) {
                    container.innerHTML = examples.map(example => 
                        '<div class="example-item" data-on-click="selectExample(\'' + example.name + '\')">' +
                        '<strong>' + example.name + '</strong>' +
                        '</div>'
                    ).join('');
                } else {
                    container.innerHTML = '<div class="empty-state">No examples found</div>';
                }
            }
        });
        
        function selectExample(exampleName) {
            // Update active state
            document.querySelectorAll('.example-item').forEach(item => {
                item.classList.remove('active');
            });
            event.target.classList.add('active');
            
            currentExample = exampleName;
            document.getElementById('current-example').textContent = exampleName;
            
            // Show loading
            document.getElementById('output-container').innerHTML = 
                '<div class="loading">Generating outputs...</div>';
            
            // Generate outputs
            fetch('/api/generate/' + exampleName, { method: 'POST' })
                .then(response => response.json())
                .then(result => {
                    if (result.success) {
                        generatedOutputs[exampleName] = result;
                        showCurrentTab();
                    } else {
                        document.getElementById('output-container').innerHTML = 
                            '<div class="error">Error: ' + (result.error || 'Generation failed') + '</div>';
                    }
                })
                .catch(err => {
                    document.getElementById('output-container').innerHTML = 
                        '<div class="error">Network error: ' + err.message + '</div>';
                });
        }
        
        function showTab(tabName) {
            // Update tab active state
            document.querySelectorAll('.tab').forEach(tab => {
                tab.classList.remove('active');
            });
            event.target.classList.add('active');
            
            currentTab = tabName;
            showCurrentTab();
        }
        
        function showCurrentTab() {
            if (!currentExample || !generatedOutputs[currentExample]) {
                return;
            }
            
            const outputs = generatedOutputs[currentExample];
            const container = document.getElementById('output-container');
            
            let content = '';
            
            switch(currentTab) {
                case 'svg':
                    if (outputs.svgUrl) {
                        content = '<embed src="' + outputs.svgUrl + '" type="image/svg+xml" class="output-frame">';
                    }
                    break;
                case 'png':
                    if (outputs.pngUrl) {
                        content = '<img src="' + outputs.pngUrl + '" class="output-frame" style="object-fit: contain;">';
                    }
                    break;
                case 'pdf':
                    if (outputs.pdfUrl) {
                        content = '<embed src="' + outputs.pdfUrl + '" type="application/pdf" class="output-frame">';
                    }
                    break;
            }
            
            if (content) {
                container.innerHTML = content;
            } else {
                container.innerHTML = '<div class="error">No ' + currentTab.toUpperCase() + ' output available</div>';
            }
        }
    </script>
</body>
</html>`