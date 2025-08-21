package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joeblew999/infra/pkg/mjml"
)

func main() {
	var (
		testEmail = flag.String("email", "", "Email address to send test emails to (required)")
		single    = flag.String("single", "", "Send only one template (simple, welcome, reset, notification, business)")
		validate  = flag.Bool("validate", false, "Only validate HTML compatibility, don't send")
	)
	flag.Parse()

	if *testEmail == "" && !*validate {
		fmt.Println("Usage:")
		fmt.Println("  go run . -email your@email.com                    # Send all test emails")
		fmt.Println("  go run . -email your@email.com -single welcome    # Send only welcome email")
		fmt.Println("  go run . -validate                                # Validate HTML only")
		fmt.Println("")
		fmt.Println("Environment variables needed for sending:")
		fmt.Println("  GMAIL_USERNAME=your-email@gmail.com")
		fmt.Println("  GMAIL_APP_PASSWORD=your-16-char-app-password")
		fmt.Println("")
		fmt.Println("To get Gmail app password:")
		fmt.Println("  1. Enable 2FA on your Google account")
		fmt.Println("  2. Go to https://myaccount.google.com/apppasswords")
		fmt.Println("  3. Generate an app password for 'Mail'")
		os.Exit(1)
	}

	// Validate mode
	if *validate {
		fmt.Println("🔍 Validating email HTML compatibility...")
		validateAllEmails()
		return
	}

	// Configure Gmail SMTP
	config := mjml.GetGmailConfig()
	if config.Username == "" || config.Password == "" {
		log.Fatal("Missing GMAIL_USERNAME or GMAIL_APP_PASSWORD environment variables")
	}

	// Send specific template or all templates
	if *single != "" {
		err := sendSingleTemplate(config, *testEmail, *single)
		if err != nil {
			log.Fatalf("Failed to send %s template: %v", *single, err)
		}
		fmt.Printf("✅ Successfully sent %s template to %s\n", *single, *testEmail)
	} else {
		fmt.Printf("📧 Sending all test emails to %s...\n", *testEmail)
		err := mjml.SendAllTestEmails(config, *testEmail)
		if err != nil {
			log.Fatalf("Failed to send test emails: %v", err)
		}
		fmt.Println("✅ All test emails sent successfully!")
		fmt.Println("📱 Check your inbox in Gmail, Outlook, Apple Mail, etc.")
	}
}

func validateAllEmails() {
	emailFiles := []string{
		"../demo/simple_email.html",
		"../demo/welcome_email.html", 
		"../demo/reset_password_email.html",
		"../demo/notification_email.html",
		"../demo/business_announcement_email.html",
	}

	for _, filename := range emailFiles {
		content, err := os.ReadFile(filename)
		if err != nil {
			fmt.Printf("❌ Error reading %s: %v\n", filename, err)
			continue
		}

		issues := mjml.ValidateEmailHTML(string(content))
		if len(issues) == 0 {
			fmt.Printf("✅ %s - No issues found\n", filename)
		} else {
			fmt.Printf("⚠️  %s - Issues found:\n", filename)
			for _, issue := range issues {
				fmt.Printf("   • %s\n", issue)
			}
		}
	}
}

func sendSingleTemplate(config mjml.EmailTestConfig, email, template string) error {
	templateMap := map[string]string{
		"simple":       "../demo/simple_email.html",
		"welcome":      "../demo/welcome_email.html",
		"reset":        "../demo/reset_password_email.html",
		"notification": "../demo/notification_email.html",
		"business":     "../demo/business_announcement_email.html",
	}

	filename, exists := templateMap[template]
	if !exists {
		return fmt.Errorf("unknown template: %s. Available: simple, welcome, reset, notification, business", template)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading template: %v", err)
	}

	subject := fmt.Sprintf("MJML Test - %s Template", template)
	return mjml.SendTestEmail(config, email, subject, string(content))
}