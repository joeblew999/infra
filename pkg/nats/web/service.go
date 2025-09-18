package web

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	datastarlib "github.com/starfederation/datastar-go/datastar"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
	natspkg "github.com/joeblew999/infra/pkg/nats"
	"github.com/joeblew999/infra/pkg/webapp/templates"
)

//go:embed templates/nats-page.html
var natsPageTemplate string

const streamInterval = 5 * time.Second

func init() {
	templates.RegisterNavItem(templates.NavItem{
		Href:  "/nats",
		Text:  "NATS",
		Icon:  "🛰️",
		Color: "sky",
		Order: 55,
	})
}

// WebService serves the NATS cluster dashboard.
type WebService struct{}

// NewWebService constructs the service.
func NewWebService() *WebService { return &WebService{} }

// RegisterRoutes mounts routes for the NATS dashboard.
func (s *WebService) RegisterRoutes(r chi.Router) {
	r.Get("/", HandleClusterPage)
	r.Get("/api/stream", HandleClusterStream)
}

// HandleClusterPage renders the main dashboard page.
func HandleClusterPage(w http.ResponseWriter, r *http.Request) {
	pageHTML, err := renderPage()
	if err != nil {
		log.Error("error rendering NATS cluster page", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	fullHTML, err := templates.RenderBasePage("NATS", pageHTML, "/nats")
	if err != nil {
		log.Error("error rendering base page", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(fullHTML))
}

func renderPage() (string, error) {
	tmpl, err := template.New("nats-page").Parse(natsPageTemplate)
	if err != nil {
		return "", fmt.Errorf("parse nats page template: %w", err)
	}

	data := struct{ BasePath string }{BasePath: "/nats"}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute nats page template: %w", err)
	}
	return buf.String(), nil
}

// HandleClusterStream streams cluster updates over SSE.
func HandleClusterStream(w http.ResponseWriter, r *http.Request) {
	sse := datastarlib.NewSSE(w, r)

	if err := sendClusterSnapshot(sse); err != nil {
		log.Error("error sending NATS cluster snapshot", "error", err)
		return
	}

	ticker := time.NewTicker(streamInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			if err := sendClusterSnapshot(sse); err != nil {
				log.Error("error streaming NATS cluster snapshot", "error", err)
				return
			}
		}
	}
}

func sendClusterSnapshot(sse *datastarlib.ServerSentEventGenerator) error {
	cluster, err := natspkg.GetClusterStatus(config.IsDevelopment())
	if err != nil {
		return err
	}

	html, err := RenderClusterCards(buildClusterTemplateData(cluster, config.IsDevelopment()))
	if err != nil {
		return err
	}

	return sse.PatchElements(html)
}
