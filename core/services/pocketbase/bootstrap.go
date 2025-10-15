package pocketbase

import (
	"fmt"
	"os"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// BootstrapAuth configures PocketBase authentication settings on app startup.
// This includes creating default admin user and ensuring proper collection setup.
// Exported for reuse by pocketbase-ha service.
func BootstrapAuth(app *pocketbase.PocketBase) error {
	// Use OnServe hook - this runs after database is initialized
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		// Ensure default admin user exists (after DB is ready)
		if err := ensureAdminUser(app); err != nil {
			return fmt.Errorf("ensure admin user: %w", err)
		}

		// Configure SMTP and OAuth2 hints
		if err := configureSettings(app); err != nil {
			return fmt.Errorf("configure settings: %w", err)
		}

		return e.Next()
	})

	return nil
}

// configureSettings updates app settings from environment variables
func configureSettings(app *pocketbase.PocketBase) error {
	// In PocketBase v0.29+, settings are managed differently
	// We'll use the admin UI or direct database updates for SMTP/OAuth2
	// For now, just log that configuration should be done via admin UI

	appURL := os.Getenv("CORE_POCKETBASE_APP_URL")
	smtpHost := os.Getenv("CORE_POCKETBASE_SMTP_HOST")

	if appURL != "" || smtpHost != "" {
		fmt.Println("Note: SMTP and OAuth2 configuration should be done via PocketBase Admin UI")
		fmt.Println("Admin UI is available at: http://localhost:8090/_/")
		if appURL != "" {
			fmt.Printf("App URL: %s\n", appURL)
		}
		if smtpHost != "" {
			fmt.Printf("SMTP Host: %s\n", smtpHost)
		}
	}

	return nil
}

// getEnvOrDefault returns environment variable value or default if empty
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ensureAdminUser creates a default superuser with sensible defaults
func ensureAdminUser(app *pocketbase.PocketBase) error {
	// Use defaults for local development, can be overridden via env
	email := getEnvOrDefault("CORE_POCKETBASE_ADMIN_EMAIL", "admin@localhost")
	password := getEnvOrDefault("CORE_POCKETBASE_ADMIN_PASSWORD", "changeme123")

	// Check if any superuser already exists
	superusers, err := app.FindAllRecords(core.CollectionNameSuperusers)
	if err == nil && len(superusers) > 0 {
		// Superusers already exist
		return nil
	}

	// Create new superuser
	collection, err := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	if err != nil {
		return fmt.Errorf("find superusers collection: %w", err)
	}

	superuser := core.NewRecord(collection)

	superuser.Set("email", email)
	superuser.SetPassword(password)

	if err := app.Save(superuser); err != nil {
		return fmt.Errorf("save superuser: %w", err)
	}

	fmt.Printf("Created superuser: %s\n", email)
	return nil
}
