package datastar

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"sync"
)

// StatusTemplateData contains all values required to render the live status dashboard.
type StatusTemplateData struct {
	LastUpdatedDisplay string
	LastUpdatedISO     string
	SummaryIcon        string
	SummaryHeadline    string
	SummaryBody        string
	SummaryBorder      string
	SummaryGradient    string
	Uptime             string
	LoadAverage        string
	Runtime            StatusRuntime
	Services           []StatusService
}

// StatusRuntime captures runtime-oriented statistics for display.
type StatusRuntime struct {
	Goroutines     int
	NumCPU         int
	HeapAlloc      string
	HeapSys        string
	StackInuse     string
	NextGC         string
	LastGCPause    string
	TotalGC        uint32
	GoVersion      string
	GOOS           string
	GOARCH         string
	MemoryPercent  float64
	MemoryBarClass string
}

// StatusService represents a service badge in the dashboard.
type StatusService struct {
	Name   string
	Status string
	Detail string
	Icon   string
	Border string
	Pill   string
	Port   int
}

var (
	statusCardsTplOnce sync.Once
	statusCardsTpl     *template.Template
	statusCardsTplErr  error
)

//go:embed templates/status_cards.html
var statusCardsTemplate string

// RenderStatusCards renders the live status dashboard partial for DataStar SSE updates.
func RenderStatusCards(data StatusTemplateData) (string, error) {
	statusCardsTplOnce.Do(func() {
		statusCardsTpl, statusCardsTplErr = template.New("datastar-status-cards").Parse(statusCardsTemplate)
	})

	if statusCardsTplErr != nil {
		return "", fmt.Errorf("parse status cards template: %w", statusCardsTplErr)
	}

	var buf bytes.Buffer
	if err := statusCardsTpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute status cards template: %w", err)
	}

	return buf.String(), nil
}
