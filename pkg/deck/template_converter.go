package deck

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	
	"github.com/joeblew999/infra/pkg/config"
)

// TemplateConverter handles template-based markdown to deck conversion
type TemplateConverter struct {
	TemplateDir string
	Manager     *Manager
}

// NewTemplateConverter creates a new template converter
func NewTemplateConverter() *TemplateConverter {
	return &TemplateConverter{
		TemplateDir: filepath.Join("pkg", "deck", "templates"),
		Manager:     NewManager(),
	}
}

// ConvertMarkdownWithTemplate converts markdown using a deck template
func (tc *TemplateConverter) ConvertMarkdownWithTemplate(markdownFile, templateName, outputSVG string) error {
	// Read template
	templatePath := filepath.Join(tc.TemplateDir, templateName+".dsh")
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %w", templatePath, err)
	}

	// Extract title from markdown
	title, err := tc.extractTitle(markdownFile)
	if err != nil {
		return fmt.Errorf("failed to extract title: %w", err)
	}

	// Replace template placeholders
	dshContent := string(templateContent)
	dshContent = strings.ReplaceAll(dshContent, "{{TITLE}}", title)
	dshContent = strings.ReplaceAll(dshContent, "{{MARKDOWN_FILE}}", markdownFile)
	dshContent = strings.ReplaceAll(dshContent, "{{BUILD_HASH}}", config.GetShortHash())

	// Create temporary .dsh file
	tempDSH := strings.TrimSuffix(outputSVG, ".svg") + "_template.dsh"
	if err := os.WriteFile(tempDSH, []byte(dshContent), 0644); err != nil {
		return fmt.Errorf("failed to write template DSH: %w", err)
	}
	defer os.Remove(tempDSH) // Cleanup

	// Validate DSH template using decklint
	if err := tc.validateDSH(tempDSH); err != nil {
		return fmt.Errorf("DSH template validation failed: %w", err)
	}

	// Generate XML from DSH using decksh
	tempXML := strings.TrimSuffix(outputSVG, ".svg") + "_template.xml"
	if err := tc.generateXMLFromDSH(tempDSH, tempXML); err != nil {
		return fmt.Errorf("failed to generate XML from DSH: %w", err)
	}
	defer os.Remove(tempXML) // Cleanup
	
	// Use deck pipeline: XML -> SVG
	if err := tc.Manager.ConvertToPDF(tempXML, outputSVG); err != nil {
		return fmt.Errorf("failed to convert with template: %w", err)
	}

	return nil
}

// extractTitle extracts the first # heading from markdown file
func (tc *TemplateConverter) extractTitle(markdownFile string) (string, error) {
	content, err := os.ReadFile(markdownFile)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(line[2:]), nil
		}
	}

	// Fallback to filename
	base := filepath.Base(markdownFile)
	return strings.TrimSuffix(base, filepath.Ext(base)), nil
}

// generateXMLFromDSH runs decksh to convert DSH to XML
func (tc *TemplateConverter) generateXMLFromDSH(dshPath, xmlPath string) error {
	deckshPath := filepath.Join("pkg", "deck", ".build", "bin", "decksh")
	
	// Check if decksh exists
	if _, err := os.Stat(deckshPath); os.IsNotExist(err) {
		return fmt.Errorf("decksh not found: %s", deckshPath)
	}
	
	// Run decksh
	cmd := exec.Command(deckshPath, dshPath)
	cmd.Env = append(os.Environ(), "DECKFONTS="+config.GetFontPath())
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("decksh failed: %w, output: %s", err, string(output))
	}
	
	// Write XML output
	return os.WriteFile(xmlPath, output, 0644)
}

// validateDSH runs decklint to validate DSH syntax
func (tc *TemplateConverter) validateDSH(dshPath string) error {
	decklintPath := filepath.Join("pkg", "deck", ".build", "bin", "decklint")
	
	// Check if decklint exists
	if _, err := os.Stat(decklintPath); os.IsNotExist(err) {
		return fmt.Errorf("decklint not found: %s", decklintPath)
	}
	
	// Run decklint
	cmd := exec.Command(decklintPath, dshPath)
	cmd.Env = append(os.Environ(), "DECKFONTS="+config.GetFontPath())
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("decklint validation failed: %w, output: %s", err, string(output))
	}
	
	// If there's output from decklint, it usually means warnings or errors
	if len(output) > 0 {
		return fmt.Errorf("DSH validation warnings: %s", string(output))
	}
	
	return nil
}