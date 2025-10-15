package config

// PocketBase SMTP Configuration
const (
	// EnvVarPocketBaseSMTPHost is the SMTP server hostname
	EnvVarPocketBaseSMTPHost = "CORE_POCKETBASE_SMTP_HOST"

	// EnvVarPocketBaseSMTPPort is the SMTP server port (typically 587 for TLS, 465 for SSL)
	EnvVarPocketBaseSMTPPort = "CORE_POCKETBASE_SMTP_PORT"

	// EnvVarPocketBaseSMTPUsername is the SMTP authentication username
	EnvVarPocketBaseSMTPUsername = "CORE_POCKETBASE_SMTP_USERNAME"

	// EnvVarPocketBaseSMTPPassword is the SMTP authentication password
	EnvVarPocketBaseSMTPPassword = "CORE_POCKETBASE_SMTP_PASSWORD"

	// EnvVarPocketBaseSMTPFrom is the sender email address for outgoing emails
	EnvVarPocketBaseSMTPFrom = "CORE_POCKETBASE_SMTP_FROM"

	// EnvVarPocketBaseSMTPTLS enables TLS for SMTP connections
	EnvVarPocketBaseSMTPTLS = "CORE_POCKETBASE_SMTP_TLS"
)

// PocketBase Application Configuration
const (
	// EnvVarPocketBaseAppURL is the public URL of the application (used in email links)
	EnvVarPocketBaseAppURL = "CORE_POCKETBASE_APP_URL"

	// EnvVarPocketBaseAdminEmail is the email for the default admin user
	EnvVarPocketBaseAdminEmail = "CORE_POCKETBASE_ADMIN_EMAIL"

	// EnvVarPocketBaseAdminPassword is the password for the default admin user
	EnvVarPocketBaseAdminPassword = "CORE_POCKETBASE_ADMIN_PASSWORD"
)

// PocketBase OAuth2 Provider Configuration
const (
	// Google OAuth2
	EnvVarPocketBaseGoogleClientID     = "CORE_POCKETBASE_GOOGLE_CLIENT_ID"
	EnvVarPocketBaseGoogleClientSecret = "CORE_POCKETBASE_GOOGLE_CLIENT_SECRET"

	// GitHub OAuth2
	EnvVarPocketBaseGitHubClientID     = "CORE_POCKETBASE_GITHUB_CLIENT_ID"
	EnvVarPocketBaseGitHubClientSecret = "CORE_POCKETBASE_GITHUB_CLIENT_SECRET"

	// Microsoft OAuth2
	EnvVarPocketBaseMicrosoftClientID     = "CORE_POCKETBASE_MICROSOFT_CLIENT_ID"
	EnvVarPocketBaseMicrosoftClientSecret = "CORE_POCKETBASE_MICROSOFT_CLIENT_SECRET"

	// Apple OAuth2
	EnvVarPocketBaseAppleClientID     = "CORE_POCKETBASE_APPLE_CLIENT_ID"
	EnvVarPocketBaseAppleClientSecret = "CORE_POCKETBASE_APPLE_CLIENT_SECRET"
)
