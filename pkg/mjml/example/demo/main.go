package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joeblew999/infra/pkg/mjml"
)

func main() {
	// Create renderer with caching enabled
	renderer := mjml.NewRenderer(
		mjml.WithCache(true),
		mjml.WithDebug(false),
		mjml.WithTemplateDir("../../templates"),
	)

	// Load templates from the templates directory
	err := renderer.LoadTemplatesFromDir("../../templates")
	if err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}

	fmt.Printf("Loaded templates: %v\n\n", renderer.ListTemplates())

	// Example 1: Simple email
	fmt.Println("=== Simple Email Example ===")
	simpleData := mjml.EmailData{
		Name:    "John Doe",
		Email:   "john@example.com",
		Subject: "Simple Test Email",
		Title:   "Test Email",
		Message: "This is a simple test email generated using MJML templates.",
		ButtonText: "Visit Website",
		ButtonURL:  "https://example.com",
	}

	html, err := renderer.RenderTemplate("simple", simpleData)
	if err != nil {
		log.Fatalf("Failed to render simple template: %v", err)
	}
	fmt.Printf("Generated HTML length: %d characters\n", len(html))
	
	// Save to file for inspection
	err = os.WriteFile("simple_email.html", []byte(html), 0644)
	if err != nil {
		log.Printf("Failed to save simple email: %v", err)
	} else {
		fmt.Println("✓ Saved to simple_email.html")
	}

	// Example 2: Welcome email
	fmt.Println("\n=== Welcome Email Example ===")
	welcomeData := mjml.WelcomeEmailData{
		EmailData: mjml.EmailData{
			Name:        "Jane Smith",
			Email:       "jane@example.com",
			Subject:     "Welcome to Our Platform!",
			CompanyName: "Tech Innovations Inc",
			CompanyLogo: "https://via.placeholder.com/200x80/3498db/ffffff?text=LOGO",
			CompanyURL:  "https://techinnovations.com",
			Message:     "Welcome to our platform! We're excited to have you on board. Click the button below to activate your account and get started.",
			Timestamp:   time.Now(),
		},
		ActivationURL: "https://techinnovations.com/activate?token=abc123",
		LoginURL:      "https://techinnovations.com/login",
	}

	html, err = renderer.RenderTemplate("welcome", welcomeData)
	if err != nil {
		log.Fatalf("Failed to render welcome template: %v", err)
	}
	fmt.Printf("Generated HTML length: %d characters\n", len(html))
	
	err = os.WriteFile("welcome_email.html", []byte(html), 0644)
	if err != nil {
		log.Printf("Failed to save welcome email: %v", err)
	} else {
		fmt.Println("✓ Saved to welcome_email.html")
	}

	// Example 3: Password reset email
	fmt.Println("\n=== Password Reset Email Example ===")
	resetData := mjml.ResetPasswordData{
		EmailData: mjml.EmailData{
			Name:        "Bob Johnson",
			Email:       "bob@example.com",
			Subject:     "Password Reset Request",
			CompanyName: "Secure Systems Corp",
			Timestamp:   time.Now(),
		},
		ResetURL:    "https://securesystems.com/reset?token=xyz789",
		ExpiresIn:   24 * time.Hour,
		RequestIP:   "192.168.1.100",
		RequestTime: time.Now(),
	}

	html, err = renderer.RenderTemplate("reset_password", resetData)
	if err != nil {
		log.Fatalf("Failed to render reset password template: %v", err)
	}
	fmt.Printf("Generated HTML length: %d characters\n", len(html))
	
	err = os.WriteFile("reset_password_email.html", []byte(html), 0644)
	if err != nil {
		log.Printf("Failed to save reset password email: %v", err)
	} else {
		fmt.Println("✓ Saved to reset_password_email.html")
	}

	// Example 4: Notification email
	fmt.Println("\n=== Notification Email Example ===")
	notificationData := mjml.NotificationData{
		EmailData: mjml.EmailData{
			Name:        "Alice Williams",
			Email:       "alice@example.com",
			Subject:     "System Alert: High CPU Usage",
			CompanyName: "Monitoring Solutions",
			Title:       "System Alert",
			Message:     "We've detected high CPU usage on your server. This requires immediate attention to prevent service disruption.",
			ButtonText:  "View Dashboard",
			ButtonURL:   "https://monitoring.com/dashboard",
			Timestamp:   time.Now(),
		},
		NotificationType: "System Alert",
		Priority:         "high",
		ActionRequired:   true,
		Details: map[string]interface{}{
			"server":     "web-01",
			"cpu_usage":  "95%",
			"memory":     "78%",
			"disk_space": "45%",
		},
	}

	html, err = renderer.RenderTemplate("notification", notificationData)
	if err != nil {
		log.Fatalf("Failed to render notification template: %v", err)
	}
	fmt.Printf("Generated HTML length: %d characters\n", len(html))
	
	err = os.WriteFile("notification_email.html", []byte(html), 0644)
	if err != nil {
		log.Printf("Failed to save notification email: %v", err)
	} else {
		fmt.Println("✓ Saved to notification_email.html")
	}

	// Example 5: Business announcement email
	fmt.Println("\n=== Business Announcement Email Example ===")
	businessData := map[string]interface{}{
		"subject":                "Grand Opening in Austin!",
		"preview":                "Join us for our grand opening event",
		"company_name":           "Croft's Accountants",
		"company_logo":           "https://via.placeholder.com/150x60/040B4F/ffffff?text=CROFTS",
		"name":                   "Sarah Davis",
		"location":               "Austin, TX",
		"venue":                  "Austin Convention Center",
		"address":                "123 Main Street, 78701",
		"title":                  "Croft's in Austin is opening December 20th",
		"message":                "We're excited to announce our new location opening in Austin! Join us for our grand opening event with special offers and networking opportunities.",
		"description":            "This new location will provide comprehensive accounting services to the Austin business community.",
		"features": []map[string]string{
			{"title": "Tax Services", "description": "Complete tax preparation and planning"},
			{"title": "Business Consulting", "description": "Strategic advice for growing businesses"},
			{"title": "Bookkeeping", "description": "Professional bookkeeping and financial management"},
		},
		"call_to_action_text":    "Don't miss this opportunity to connect with our team and learn about our services.",
		"primary_button_text":    "RSVP Today",
		"primary_button_url":     "https://crofts.com/rsvp",
		"secondary_button_text":  "Book an Appointment",
		"secondary_button_url":   "https://crofts.com/appointment",
		"additional_info":        "Refreshments will be provided and door prizes will be given away throughout the event.",
		"closing_message":        "We look forward to meeting you and serving the Austin business community.",
		"visit_title":            "Come see us!",
		"visit_message":          "We're looking forward to meeting you.",
		"hours": []map[string]string{
			{"day": "Monday, December 20th", "time": "8:00AM - 5:00PM"},
			{"day": "Tuesday, December 21st", "time": "8:00AM - 5:00PM"},
		},
		"map_image":              "https://via.placeholder.com/570x200/cccccc/666666?text=MAP",
		"social_links": []map[string]string{
			{"platform": "facebook", "url": "https://facebook.com/crofts"},
			{"platform": "twitter", "url": "https://twitter.com/crofts"},
			{"platform": "linkedin", "url": "https://linkedin.com/company/crofts"},
		},
		"disclaimer":             "You are receiving this email because you registered with Croft's Accountants and agreed to receive emails from us regarding new features, events and special offers.",
		"privacy_url":            "https://crofts.com/privacy",
		"unsubscribe_url":        "https://crofts.com/unsubscribe",
	}

	html, err = renderer.RenderTemplate("business_announcement", businessData)
	if err != nil {
		log.Fatalf("Failed to render business announcement template: %v", err)
	}
	fmt.Printf("Generated HTML length: %d characters\n", len(html))
	
	err = os.WriteFile("business_announcement_email.html", []byte(html), 0644)
	if err != nil {
		log.Printf("Failed to save business announcement email: %v", err)
	} else {
		fmt.Println("✓ Saved to business_announcement_email.html")
	}

	// Show cache statistics
	fmt.Printf("\nCache statistics: %d items cached\n", renderer.GetCacheSize())

	fmt.Println("\n✅ All examples completed successfully!")
	fmt.Println("📧 Generated email files can be opened in a web browser to preview")
	fmt.Println("🔧 Templates can be found in ../../templates/ directory")
	fmt.Println("📝 This demonstrates how AI/MCP can generate emails using MJML templates")
}