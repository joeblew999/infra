package auth

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nkeys"
	"github.com/starfederation/datastar-go/datastar"
	pkgweb "github.com/joeblew999/infra/pkg/web"
	"github.com/joeblew999/infra/pkg/log"
)

// Store holds Datastar signals for auth flows
type Store struct {
	Username string `json:"username"`
}

// PageData holds the template data for rendering pages
type PageData struct {
	Navigation template.HTML
	Footer     template.HTML
}

// DatastarHandlers provides Datastar SSE fragment handlers for WebAuthn
type DatastarHandlers struct {
	authService *WebAuthnService
	templates   *template.Template
	webDir      string
}

// NewDatastarHandlers creates new Datastar handlers
func NewDatastarHandlers(authService *WebAuthnService, webDir string) *DatastarHandlers {
	// Load HTML templates from pkg/auth/web/fragments
	templates := template.New("")
	templates.Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
	})
	
	// Load fragment templates if webDir is provided
	if webDir != "" {
		globPattern := webDir + "/fragments/*.html"
		fmt.Printf("DEBUG: Loading templates from: %s\n", globPattern)
		
		// Use ParseGlob but don't panic if no files found (for container deployments)
		if _, err := templates.ParseGlob(globPattern); err != nil {
			fmt.Printf("DEBUG: No fragment templates found at %s (continuing without auth fragments)\n", globPattern)
		}
		
		// Debug: List loaded templates
		for _, tmpl := range templates.Templates() {
			fmt.Printf("DEBUG: Loaded template: %s\n", tmpl.Name())
		}
	}
	
	return &DatastarHandlers{
		authService: authService,
		templates:   templates,
		webDir:      webDir,
	}
}

// RegisterRoutes sets up Chi routes for auth handlers
func (h *DatastarHandlers) RegisterRoutes(r chi.Router) {
	r.Post("/register/start", h.RegisterStart)
	r.Post("/register/finish", h.RegisterFinish)
	r.Post("/login/start", h.LoginStart)
	r.Post("/login/finish", h.LoginFinish)
	
	// Credential management routes
	r.Get("/credentials", h.ShowCredentials)
	r.Get("/credentials/list", h.ListCredentials)
	r.Delete("/credentials/{index}", h.DeleteCredential)
	
	// Session status routes
	r.Get("/session/status", h.CheckSessionStatus)
	
	// Serve static files if webDir is configured
	if h.webDir != "" {
		r.Get("/", h.ServeIndex)
		r.Get("/auth.css", h.ServeCSS)
		r.Get("/auth.js", h.ServeJS)
	}
}

// ServeIndex serves the main index.html page
func (h *DatastarHandlers) ServeIndex(w http.ResponseWriter, r *http.Request) {
	h.renderTemplate(w, "/auth", "index")
}

// renderTemplate is a helper for rendering templates with nav/footer
func (h *DatastarHandlers) renderTemplate(w http.ResponseWriter, currentPath, templateName string) {
	// Get centralized navigation and footer
	navHTML, err := pkgweb.RenderNav(currentPath)
	if err != nil {
		log.Error("Error rendering navigation", "error", err)
		http.Error(w, "Failed to render navigation", http.StatusInternalServerError)
		return
	}
	
	footerHTML, err := pkgweb.RenderFooter()
	if err != nil {
		log.Error("Error rendering footer", "error", err)
		footerHTML = ""
	}
	
	// Load template from filesystem
	templatePath := h.webDir + "/" + templateName + ".html"
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		log.Error("Error loading template", "template", templateName, "error", err)
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}
	
	data := PageData{
		Navigation: template.HTML(navHTML),
		Footer:     template.HTML(footerHTML),
	}
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Error("Error executing template", "template", templateName, "error", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

// ServeCSS serves the auth.css file
func (h *DatastarHandlers) ServeCSS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")
	http.ServeFile(w, r, h.webDir+"/auth.css")
}

// ServeJS serves the auth.js file
func (h *DatastarHandlers) ServeJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	http.ServeFile(w, r, h.webDir+"/auth.js")
}

// RegisterStart handles /register/start with Datastar SSE
func (h *DatastarHandlers) RegisterStart(w http.ResponseWriter, r *http.Request) {
	// Read signals from request
	store := &Store{}
	if err := datastar.ReadSignals(r, store); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if store.Username == "" {
		sse := datastar.NewSSE(w, r)
		sse.PatchElements(`<div id="log">Error: Please enter username</div>`)
		return
	}

	// Detect browser for optimal WebAuthn settings
	userAgent := r.Header.Get("User-Agent")
	usePlatform := strings.Contains(userAgent, "Safari") && !strings.Contains(userAgent, "Chrome")
	
	options, token, err := h.authService.BeginRegistrationWithPlatform(store.Username, usePlatform)
	if err != nil {
		sse := datastar.NewSSE(w, r)
		sse.PatchElements(fmt.Sprintf(`<div id="log">Registration error: %s</div>`, err.Error()))
		return
	}

	// Create SSE connection
	sse := datastar.NewSSE(w, r)
	
	// Update log with progress
	sse.PatchElements(`<div id="log">Starting registration...</div>`)
	
	// Convert options to JSON and execute WebAuthn
	optionsJSON, _ := json.Marshal(options)
	script := fmt.Sprintf(`
		(async () => {
			try {
				const options = %s;
				window.currentToken = '%s';
				const response = await startRegistration(options);
				// Send both token and WebAuthn response to finish endpoint
				await fetch('/register/finish', {
					method: 'POST',
					headers: {'Content-Type': 'application/json'},
					body: JSON.stringify({token: '%s', response: response})
				});
				document.getElementById('log').textContent = 'Registration completed successfully!';
			} catch (e) {
				document.getElementById('log').textContent = 'Registration failed: ' + e.message;
			}
		})();
	`, string(optionsJSON), token, token)
	
	sse.ExecuteScript(script)
}

// LoginStart handles /login/start with Datastar SSE
func (h *DatastarHandlers) LoginStart(w http.ResponseWriter, r *http.Request) {
	// Read signals from request
	store := &Store{}
	if err := datastar.ReadSignals(r, store); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if store.Username == "" {
		sse := datastar.NewSSE(w, r)
		sse.PatchElements(`<div id="log">Error: Please enter username</div>`)
		return
	}

	options, token, err := h.authService.BeginLogin(store.Username)
	if err != nil {
		sse := datastar.NewSSE(w, r)
		sse.PatchElements(fmt.Sprintf(`<div id="log">Login error: %s</div>`, err.Error()))
		return
	}

	// Create SSE connection
	sse := datastar.NewSSE(w, r)
	
	// Update log with progress
	sse.PatchElements(`<div id="log">Starting login...</div>`)
	
	// Convert options to JSON and execute WebAuthn
	optionsJSON, _ := json.Marshal(options)
	script := fmt.Sprintf(`
		(async () => {
			try {
				const options = %s;
				window.currentToken = '%s';
				const response = await startAuthentication(options);
				// Send both token and WebAuthn response to finish endpoint
				await fetch('/login/finish', {
					method: 'POST',
					headers: {'Content-Type': 'application/json'},
					body: JSON.stringify({token: '%s', response: response})
				});
				document.getElementById('log').textContent = 'Login completed successfully!';
			} catch (e) {
				document.getElementById('log').textContent = 'Login failed: ' + e.message;
			}
		})();
	`, string(optionsJSON), token, token)
	
	sse.ExecuteScript(script)
}

// RegisterFinish handles /register/finish (called from WebAuthn JS)
func (h *DatastarHandlers) RegisterFinish(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token    string      `json:"token"`
		Response interface{} `json:"response"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sse := datastar.NewSSE(w, r)
		sse.PatchElements(`<div id="log">Invalid request</div>`)
		return
	}

	// Convert response to proper WebAuthn format and finish registration
	user, err := h.authService.FinishRegistrationFromJSON(req.Token, req.Response)
	if err != nil {
		sse := datastar.NewSSE(w, r)
		sse.PatchElements(fmt.Sprintf(`<div id="log">Registration failed: %s</div>`, err.Error()))
		return
	}

	// Generate NATS credentials
	kp, _ := nkeys.CreateUser()
	pub, _ := kp.PublicKey()
	seed, _ := kp.Seed()

	sse := datastar.NewSSE(w, r)
	sse.PatchElements(fmt.Sprintf(`
		<div id="log">Registration complete for %s!<br>
		NATS Public: %s<br>
		NATS Seed: %s</div>
	`, user.WebAuthnName(), string(pub), string(seed)))
}

// LoginFinish handles /login/finish (called from WebAuthn JS)
func (h *DatastarHandlers) LoginFinish(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token    string      `json:"token"`
		Response interface{} `json:"response"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sse := datastar.NewSSE(w, r)
		sse.PatchElements(`<div id="log">Invalid request</div>`)
		return
	}

	// Convert response to proper WebAuthn format and finish login
	user, sessionID, err := h.authService.FinishLoginFromJSON(req.Token, req.Response)
	if err != nil {
		sse := datastar.NewSSE(w, r)
		sse.PatchElements(fmt.Sprintf(`<div id="log">Login failed: %s</div>`, err.Error()))
		return
	}

	// Generate fresh NATS credentials
	kp, _ := nkeys.CreateUser()
	seed, _ := kp.Seed()

	sse := datastar.NewSSE(w, r)
	sse.PatchElements(fmt.Sprintf(`
		<div id="log">Login successful for %s!<br>
		Session: %s<br>
		NATS Seed: %s<br>
		<button data-on-click="@get('/dashboard')">Go to Dashboard</button></div>
	`, user.WebAuthnName(), sessionID, string(seed)))
}

// ShowCredentials displays the credentials management section
func (h *DatastarHandlers) ShowCredentials(w http.ResponseWriter, r *http.Request) {
	store := &Store{}
	if err := datastar.ReadSignals(r, store); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if store.Username == "" {
		sse := datastar.NewSSE(w, r)
		sse.PatchElements(`<div id="log">Error: Please enter username first</div>`)
		return
	}

	sse := datastar.NewSSE(w, r)
	
	// Show credentials section with JavaScript to make it visible
	credentialsHTML := fmt.Sprintf(`
		<div id="credentials-section" style="margin-top: 1rem; padding: 1rem; background: #fff3cd; border: 1px solid #ffeaa7; border-radius: 4px;">
			<div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem;">
				<h3 style="margin: 0;">Registered Passkeys for %s</h3>
				<button onclick="document.getElementById('credentials-section').style.display='none'"
						style="background: #6c757d; color: white; border: none; padding: 0.25rem 0.5rem; border-radius: 4px; cursor: pointer;">
					Hide
				</button>
			</div>
			
			<div data-on-load="@get('/credentials/list')" data-signals-username="'%s'">
				<div style="text-align: center; padding: 1rem;">Loading credentials...</div>
			</div>
		</div>
	`, store.Username, store.Username)
	
	sse.PatchElements(credentialsHTML)
}

// ListCredentials returns the credentials list fragment
func (h *DatastarHandlers) ListCredentials(w http.ResponseWriter, r *http.Request) {
	store := &Store{}
	if err := datastar.ReadSignals(r, store); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if store.Username == "" {
		http.Error(w, "Username required", http.StatusBadRequest)
		return
	}

	user, err := h.authService.GetUserCredentials(store.Username)
	if err != nil {
		sse := datastar.NewSSE(w, r)
		sse.PatchElements(`<div style="text-align: center; padding: 1rem; color: #dc3545;">No credentials found for this user</div>`)
		return
	}

	sse := datastar.NewSSE(w, r)
	
	if len(user.Credentials) == 0 {
		sse.PatchElements(`
			<div id="credentials-list">
				<div style="text-align: center; padding: 1rem; color: #6c757d;">
					No passkeys registered yet. Click "Register New Passkey" above to create one.
				</div>
			</div>
		`)
		return
	}
	
	// Simple credentials list HTML without templates
	credentialsHTML := `<div id="credentials-list">`
	for i, cred := range user.Credentials {
		credentialsHTML += fmt.Sprintf(`
			<div class="credential-item">
				<div class="credential-info">
					<strong>Passkey %d</strong><br>
					<div class="credential-id">ID: %.20s...</div>
					<div class="credential-meta">Uses: %d</div>
				</div>
				<button data-on-click="@delete('/credentials/%d')"
						data-signals-username="'%s'"
						class="btn btn-danger" style="font-size: 0.75rem; padding: 0.25rem 0.5rem;">
					Delete
				</button>
			</div>
		`, i+1, string(cred.ID), cred.Authenticator.SignCount, i, store.Username)
	}
	credentialsHTML += `</div>`
	
	sse.PatchElements(credentialsHTML)
}

// DeleteCredential removes a credential by index
func (h *DatastarHandlers) DeleteCredential(w http.ResponseWriter, r *http.Request) {
	store := &Store{}
	if err := datastar.ReadSignals(r, store); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	indexStr := chi.URLParam(r, "index")
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		sse := datastar.NewSSE(w, r)
		sse.PatchElements(`<div id="log">Invalid credential index</div>`)
		return
	}

	if store.Username == "" {
		sse := datastar.NewSSE(w, r)
		sse.PatchElements(`<div id="log">Error: Username required</div>`)
		return
	}

	err = h.authService.DeleteUserCredential(store.Username, index)
	if err != nil {
		sse := datastar.NewSSE(w, r)
		sse.PatchElements(fmt.Sprintf(`<div id="log">Failed to delete passkey: %s</div>`, err.Error()))
		return
	}

	sse := datastar.NewSSE(w, r)
	sse.PatchElements(`<div id="log">Passkey deleted successfully!</div>`)
	
	// Refresh the credentials list
	h.ListCredentials(w, r)
}

// CheckSessionStatus checks if user is logged in and shows appropriate UI
func (h *DatastarHandlers) CheckSessionStatus(w http.ResponseWriter, r *http.Request) {
	sessionCookie, err := r.Cookie("session")
	if err != nil {
		// No session cookie, user not logged in
		return
	}

	userID, err := h.authService.GetUserSession(sessionCookie.Value)
	if err != nil {
		// Invalid session
		return
	}

	// User is logged in, show logged-in fragments
	sse := datastar.NewSSE(w, r)
	
	// Load and render the logged-in header fragment
	if tmpl := h.templates.Lookup("logged-in-header.html"); tmpl != nil {
		var headerHTML strings.Builder
		tmpl.Execute(&headerHTML, map[string]string{"UserID": userID})
		sse.PatchElements(fmt.Sprintf(`<div id="session-status">%s</div>`, headerHTML.String()))
	}
	
	// Load and render the logged-in actions fragment
	if tmpl := h.templates.Lookup("logged-in-actions.html"); tmpl != nil {
		var actionsHTML strings.Builder
		tmpl.Execute(&actionsHTML, map[string]string{"UserID": userID})
		sse.PatchElements(fmt.Sprintf(`<div id="user-actions">%s</div>`, actionsHTML.String()))
	}
	
	// Hide the auth section
	sse.ExecuteScript(`document.getElementById('auth-section').style.display = 'none';`)
}