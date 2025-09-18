package web

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"sync"
)

// ProcessTemplateData contains the fields required to render the Goreman process dashboard.
type ProcessTemplateData struct {
	LastUpdatedDisplay string
	LastUpdatedISO     string
	Summary            ProcessSummary
	Processes          []ProcessCard
	HasProcesses       bool
}

// ProcessSummary surfaces aggregate counts for supervised processes.
type ProcessSummary struct {
	Total   int
	Running int
	Stopped int
}

// ProcessCard represents a single supervised process in the UI.
type ProcessCard struct {
	Name             string
	StatusLabel      string
	Indicator        string
	Uptime           string
	ShowStart        bool
	ShowStop         bool
	ShowRestart      bool
	StartAction      string
	StopAction       string
	RestartAction    string
	CardBorderClass  string
	StatusBadgeClass string
}

var (
	processCardsTplOnce sync.Once
	processCardsTpl     *template.Template
	processCardsTplErr  error
)

//go:embed templates/process_cards.html
var processCardsTemplate string

// RenderProcessCards renders the live process dashboard partial for DataStar SSE updates.
func RenderProcessCards(data ProcessTemplateData) (string, error) {
	processCardsTplOnce.Do(func() {
		processCardsTpl, processCardsTplErr = template.New("goreman-process-cards").Parse(processCardsTemplate)
	})

	if processCardsTplErr != nil {
		return "", fmt.Errorf("parse process cards template: %w", processCardsTplErr)
	}

	var buf bytes.Buffer
	if err := processCardsTpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute process cards template: %w", err)
	}

	return buf.String(), nil
}
