package mjml

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

// EmailTestConfig holds configuration for sending test emails
type EmailTestConfig struct {
	SMTPHost     string
	SMTPPort     string
	Username     string
	Password     string
	FromEmail    string
	FromName     string
}

// SendTestEmail sends an HTML email for testing in real email clients
func SendTestEmail(config EmailTestConfig, toEmail, subject, htmlBody string) error {
	// Create message
	message := fmt.Sprintf(
		"From: %s <%s>\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"\r\n"+
			"%s",
		config.FromName, config.FromEmail,
		toEmail,
		subject,
		htmlBody,
	)

	// Setup authentication
	auth := smtp.PlainAuth("", config.Username, config.Password, config.SMTPHost)
	
	// Send email
	err := smtp.SendMail(
		config.SMTPHost+":"+config.SMTPPort,
		auth,
		config.FromEmail,
		[]string{toEmail},
		[]byte(message),
	)
	
	return err
}

// SendAllTestEmails sends all generated email templates for testing
func SendAllTestEmails(config EmailTestConfig, testEmail string) error {
	emailFiles := map[string]string{
		"Simple Test Email":           "simple_email.html",
		"Welcome Email":              "welcome_email.html", 
		"Password Reset":             "reset_password_email.html",
		"System Alert Notification": "notification_email.html",
		"Business Announcement":      "business_announcement_email.html",
	}

	for subject, filename := range emailFiles {
		htmlContent, err := os.ReadFile(fmt.Sprintf("pkg/mjml/example/demo/%s", filename))
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", filename, err)
			continue
		}

		err = SendTestEmail(config, testEmail, subject+" - MJML Test", string(htmlContent))
		if err != nil {
			fmt.Printf("Error sending %s: %v\n", subject, err)
		} else {
			fmt.Printf("✓ Sent %s to %s\n", subject, testEmail)
		}
	}

	return nil
}

// GetGmailConfig returns a pre-configured EmailTestConfig for Gmail SMTP
func GetGmailConfig() EmailTestConfig {
	return EmailTestConfig{
		SMTPHost: "smtp.gmail.com",
		SMTPPort: "587",
		Username: os.Getenv("GMAIL_USERNAME"), // your-email@gmail.com
		Password: os.Getenv("GMAIL_APP_PASSWORD"), // Gmail app password, not regular password
		FromEmail: os.Getenv("GMAIL_USERNAME"),
		FromName:  "MJML Test Sender",
	}
}

// ValidateEmailHTML performs basic HTML validation for email compatibility
func ValidateEmailHTML(htmlContent string) []string {
	var issues []string
	
	// Check for common email client compatibility issues
	if !strings.Contains(strings.ToLower(htmlContent), "doctype html") {
		issues = append(issues, "Missing DOCTYPE declaration")
	}
	
	if !strings.Contains(htmlContent, "xmlns:v=\"urn:schemas-microsoft-com:vml\"") {
		issues = append(issues, "Missing VML namespace for Outlook compatibility")
	}
	
	if !strings.Contains(htmlContent, "<!--[if mso") {
		issues = append(issues, "Missing Outlook conditional comments")
	}
	
	if !strings.Contains(htmlContent, "border-collapse: collapse") {
		issues = append(issues, "Missing border-collapse for table compatibility")
	}
	
	if strings.Contains(htmlContent, "display: flex") {
		issues = append(issues, "WARNING: CSS flexbox not supported in many email clients")
	}
	
	if strings.Contains(htmlContent, "background-image") && !strings.Contains(htmlContent, "mso-hide") {
		issues = append(issues, "WARNING: Background images not supported in Outlook")
	}
	
	return issues
}