package mjml

import (
	"fmt"
	"html/template"
	"strings"
	"time"
)

// EmailData represents common email template data
type EmailData struct {
	// Recipient information
	Name  string `json:"name"`
	Email string `json:"email"`
	
	// Email metadata
	Subject   string    `json:"subject"`
	Timestamp time.Time `json:"timestamp"`
	
	// Brand/company information
	CompanyName string `json:"company_name"`
	CompanyLogo string `json:"company_logo"`
	CompanyURL  string `json:"company_url"`
	
	// Content
	Title   string `json:"title"`
	Message string `json:"message"`
	
	// Action items
	ButtonText string `json:"button_text"`
	ButtonURL  string `json:"button_url"`
	
	// Additional data for template flexibility
	Data map[string]interface{} `json:"data"`
}

// WelcomeEmailData represents data for welcome emails
type WelcomeEmailData struct {
	EmailData
	ActivationURL string `json:"activation_url"`
	LoginURL      string `json:"login_url"`
}

// ResetPasswordData represents data for password reset emails
type ResetPasswordData struct {
	EmailData
	ResetURL   string        `json:"reset_url"`
	ExpiresIn  time.Duration `json:"expires_in"`
	RequestIP  string        `json:"request_ip"`
	RequestTime time.Time    `json:"request_time"`
}

// NotificationData represents data for notification emails
type NotificationData struct {
	EmailData
	NotificationType string                 `json:"notification_type"`
	Priority         string                 `json:"priority"`
	Details          map[string]interface{} `json:"details"`
	ActionRequired   bool                   `json:"action_required"`
}

// DefaultTemplateNames returns the names of default templates
func DefaultTemplateNames() []string {
	return []string{"welcome", "reset_password", "notification", "simple"}
}

// DefaultTemplates returns common email templates as MJML content
func DefaultTemplates() map[string]string {
	return map[string]string{
		"welcome":        welcomeTemplate,
		"reset_password": resetPasswordTemplate,
		"notification":   notificationTemplate,
		"simple":         simpleTemplate,
	}
}

// LoadDefaultTemplates loads common email templates into the renderer atomically
func (r *Renderer) LoadDefaultTemplates() error {
	templates := DefaultTemplates()
	
	// Parse all templates first to catch errors before modifying renderer state
	parsedTemplates := make(map[string]*template.Template)
	for name, content := range templates {
		tmpl, err := template.New(name).Parse(content)
		if err != nil {
			return fmt.Errorf("failed to parse default template %s: %w", name, err)
		}
		parsedTemplates[name] = tmpl
	}
	
	// Only modify renderer state after all templates parse successfully
	r.mu.Lock()
	defer r.mu.Unlock()
	
	for name, tmpl := range parsedTemplates {
		r.templates[name] = tmpl
		// Clear cache for this template if caching is enabled
		if r.options.EnableCache {
			for key := range r.cache {
				if strings.HasPrefix(key, name+"_") {
					delete(r.cache, key)
				}
			}
		}
	}
	
	return nil
}

// TestDataFactory provides common test data patterns to reduce duplication
type TestDataFactory struct{}

// NewTestDataFactory creates a new test data factory
func NewTestDataFactory() *TestDataFactory {
	return &TestDataFactory{}
}

// SimpleEmailData creates standard simple email test data
func (f *TestDataFactory) SimpleEmailData() EmailData {
	return EmailData{
		Name:       "Test User",
		Email:      "test@example.com",
		Subject:    "Test Email",
		Title:      "Test Title",
		Message:    "This is a test message",
		ButtonText: "Click Here",
		ButtonURL:  "https://example.com",
		Timestamp:  time.Now(),
		CompanyName: "Test Company",
	}
}

// WelcomeEmailData creates standard welcome email test data
func (f *TestDataFactory) WelcomeEmailData() WelcomeEmailData {
	return WelcomeEmailData{
		EmailData: EmailData{
			Name:        "New User",
			Email:       "newuser@example.com", 
			Subject:     "Welcome!",
			CompanyName: "Test Company",
			CompanyLogo: "https://via.placeholder.com/200x80/3498db/ffffff?text=LOGO",
			CompanyURL:  "https://testcompany.com",
			Message:     "Welcome to our platform!",
			Timestamp:   time.Now(),
		},
		ActivationURL: "https://testcompany.com/activate?token=test123",
		LoginURL:      "https://testcompany.com/login",
	}
}

// ResetPasswordData creates standard password reset test data  
func (f *TestDataFactory) ResetPasswordData() ResetPasswordData {
	return ResetPasswordData{
		EmailData: EmailData{
			Name:        "User",
			Email:       "user@example.com",
			Subject:     "Password Reset",
			CompanyName: "Test Company", 
			Timestamp:   time.Now(),
		},
		ResetURL:    "https://testcompany.com/reset?token=test456",
		ExpiresIn:   24 * time.Hour,
		RequestIP:   "127.0.0.1", 
		RequestTime: time.Now(),
	}
}

// NotificationData creates standard notification test data
func (f *TestDataFactory) NotificationData() NotificationData {
	return NotificationData{
		EmailData: EmailData{
			Name:        "Alert User", 
			Email:       "alerts@example.com",
			Subject:     "System Alert",
			Title:       "Test Alert",
			Message:     "This is a test notification",
			ButtonText:  "View Details",
			ButtonURL:   "https://monitoring.com/dashboard",
			Timestamp:   time.Now(),
			CompanyName: "Monitoring Solutions",
		},
		NotificationType: "system",
		Priority:         "high",
		ActionRequired:   true,
		Details: map[string]interface{}{
			"server": "test-server",
			"metric": "CPU usage",
		},
	}
}

const welcomeTemplate = `<mjml>
  <mj-head>
    <mj-title>{{.Subject}}</mj-title>
    <mj-preview>Welcome to {{.CompanyName}}!</mj-preview>
    <mj-attributes>
      <mj-all font-family="Arial, sans-serif" />
      <mj-text color="#333333" font-size="16px" line-height="1.6" />
    </mj-attributes>
  </mj-head>
  <mj-body background-color="#f4f4f4">
    <!-- Header -->
    <mj-section background-color="#ffffff" padding="20px">
      <mj-column>
        {{if .CompanyLogo}}
        <mj-image src="{{.CompanyLogo}}" alt="{{.CompanyName}}" width="200px" />
        {{end}}
        <mj-text align="center" font-size="24px" font-weight="bold" color="#2c3e50">
          Welcome to {{.CompanyName}}!
        </mj-text>
      </mj-column>
    </mj-section>
    
    <!-- Content -->
    <mj-section background-color="#ffffff" padding="20px">
      <mj-column>
        <mj-text font-size="18px" color="#2c3e50">
          Hi {{.Name}},
        </mj-text>
        <mj-text>
          {{.Message}}
        </mj-text>
        {{if .ActivationURL}}
        <mj-button background-color="#3498db" color="#ffffff" href="{{.ActivationURL}}">
          {{if .ButtonText}}{{.ButtonText}}{{else}}Activate Account{{end}}
        </mj-button>
        {{end}}
        {{if .LoginURL}}
        <mj-text align="center">
          <a href="{{.LoginURL}}" style="color: #3498db;">Or log in here</a>
        </mj-text>
        {{end}}
      </mj-column>
    </mj-section>
    
    <!-- Footer -->
    <mj-section background-color="#ecf0f1" padding="20px">
      <mj-column>
        <mj-text align="center" font-size="12px" color="#7f8c8d">
          © {{.Timestamp.Year}} {{.CompanyName}}. All rights reserved.
        </mj-text>
        {{if .CompanyURL}}
        <mj-text align="center" font-size="12px">
          <a href="{{.CompanyURL}}" style="color: #3498db;">Visit our website</a>
        </mj-text>
        {{end}}
      </mj-column>
    </mj-section>
  </mj-body>
</mjml>
`

const resetPasswordTemplate = `<mjml>
  <mj-head>
    <mj-title>{{.Subject}}</mj-title>
    <mj-preview>Reset your password</mj-preview>
    <mj-attributes>
      <mj-all font-family="Arial, sans-serif" />
      <mj-text color="#333333" font-size="16px" line-height="1.6" />
    </mj-attributes>
  </mj-head>
  <mj-body background-color="#f4f4f4">
    <!-- Header -->
    <mj-section background-color="#ffffff" padding="20px">
      <mj-column>
        {{if .CompanyLogo}}
        <mj-image src="{{.CompanyLogo}}" alt="{{.CompanyName}}" width="200px" />
        {{end}}
        <mj-text align="center" font-size="24px" font-weight="bold" color="#e74c3c">
          Password Reset Request
        </mj-text>
      </mj-column>
    </mj-section>
    
    <!-- Content -->
    <mj-section background-color="#ffffff" padding="20px">
      <mj-column>
        <mj-text font-size="18px" color="#2c3e50">
          Hi {{.Name}},
        </mj-text>
        <mj-text>
          We received a request to reset your password. Click the button below to reset it:
        </mj-text>
        <mj-button background-color="#e74c3c" color="#ffffff" href="{{.ResetURL}}">
          {{if .ButtonText}}{{.ButtonText}}{{else}}Reset Password{{end}}
        </mj-button>
        <mj-text>
          This link will expire in {{.ExpiresIn.Hours}} hours.
        </mj-text>
        <mj-text font-size="14px" color="#7f8c8d">
          If you didn't request this, you can safely ignore this email.
          <br/>Request made from IP: {{.RequestIP}}
          <br/>Time: {{.RequestTime.Format "2006-01-02 15:04:05 UTC"}}
        </mj-text>
      </mj-column>
    </mj-section>
    
    <!-- Footer -->
    <mj-section background-color="#ecf0f1" padding="20px">
      <mj-column>
        <mj-text align="center" font-size="12px" color="#7f8c8d">
          © {{.Timestamp.Year}} {{.CompanyName}}. All rights reserved.
        </mj-text>
      </mj-column>
    </mj-section>
  </mj-body>
</mjml>
`

const notificationTemplate = `<mjml>
  <mj-head>
    <mj-title>{{.Subject}}</mj-title>
    <mj-preview>{{.NotificationType}} notification</mj-preview>
    <mj-attributes>
      <mj-all font-family="Arial, sans-serif" />
      <mj-text color="#333333" font-size="16px" line-height="1.6" />
    </mj-attributes>
  </mj-head>
  <mj-body background-color="#f4f4f4">
    <!-- Header -->
    <mj-section background-color="#ffffff" padding="20px">
      <mj-column>
        {{if .CompanyLogo}}
        <mj-image src="{{.CompanyLogo}}" alt="{{.CompanyName}}" width="200px" />
        {{end}}
        <mj-text align="center" font-size="24px" font-weight="bold" 
                 color="{{if eq .priority "high"}}#e74c3c{{else if eq .priority "medium"}}#f39c12{{else}}#3498db{{end}}">
          {{.Title}}
        </mj-text>
      </mj-column>
    </mj-section>
    
    <!-- Content -->
    <mj-section background-color="#ffffff" padding="20px">
      <mj-column>
        <mj-text font-size="18px" color="#2c3e50">
          Hi {{.Name}},
        </mj-text>
        <mj-text>
          {{.Message}}
        </mj-text>
        {{if .ActionRequired}}
        <mj-button background-color="{{if eq .priority "high"}}#e74c3c{{else if eq .priority "medium"}}#f39c12{{else}}#3498db{{end}}" 
                   color="#ffffff" href="{{.ButtonURL}}">
          {{if .ButtonText}}{{.ButtonText}}{{else}}Take Action{{end}}
        </mj-button>
        {{end}}
      </mj-column>
    </mj-section>
    
    <!-- Footer -->
    <mj-section background-color="#ecf0f1" padding="20px">
      <mj-column>
        <mj-text align="center" font-size="12px" color="#7f8c8d">
          © {{.Timestamp.Year}} {{.CompanyName}}. All rights reserved.
        </mj-text>
      </mj-column>
    </mj-section>
  </mj-body>
</mjml>
`

const simpleTemplate = `<mjml>
  <mj-head>
    <mj-title>{{.Subject}}</mj-title>
    <mj-attributes>
      <mj-all font-family="Arial, sans-serif" />
      <mj-text color="#333333" font-size="16px" line-height="1.6" />
    </mj-attributes>
  </mj-head>
  <mj-body background-color="#f4f4f4">
    <mj-section background-color="#ffffff" padding="40px">
      <mj-column>
        {{if .Title}}
        <mj-text align="center" font-size="24px" font-weight="bold" color="#2c3e50">
          {{.Title}}
        </mj-text>
        {{end}}
        {{if .Name}}
        <mj-text font-size="18px" color="#2c3e50">
          Hi {{.Name}},
        </mj-text>
        {{end}}
        <mj-text>
          {{.Message}}
        </mj-text>
        {{if .ButtonURL}}
        <mj-button background-color="#3498db" color="#ffffff" href="{{.ButtonURL}}">
          {{if .ButtonText}}{{.ButtonText}}{{else}}Click Here{{end}}
        </mj-button>
        {{end}}
      </mj-column>
    </mj-section>
  </mj-body>
</mjml>
`