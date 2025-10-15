# PocketBase Service

https://github.com/pocketbase/pocketbase

PocketBase gives us a lightweight database + admin UI we can use to store
cluster configuration and emit change events that propagate through NATS. This
service will eventually become the authoritative config store for the
orchestrator.

## ✨ Features

**Complete Authentication System:**
- ✅ Password authentication (email/username + password)
- ✅ OAuth2 providers (Google, GitHub, Microsoft, Apple)
- ✅ Email verification
- ✅ Password reset
- ✅ One-Time Passwords (OTP)
- ✅ Multi-Factor Authentication (MFA) ready
- ✅ Account management (profile, password change, email change, deletion)
- ✅ User impersonation (superuser only)
- ✅ Real-time auth state via SSE

**Datastar Integration:**
- ✅ Reactive UI components
- ✅ Server-sent events for live auth updates
- ✅ Token-based authentication with localStorage
- ✅ Comprehensive auth pages (login, signup, reset, settings)

## 📁 Files

- `service.json` — Service manifest with environment variable configuration
- `service.go` — Embedded PocketBase runner
- `auth.go` — Datastar auth routes and API endpoints
- `bootstrap.go` — Auto-configuration (admin user)
- `auth_*.html` — Datastar UI pages (embedded)

## 🚀 Quick Start

### 1. Configure Environment Variables

See [`.env.example`](../../.env.example) in the repository root. Minimal configuration:

```bash
# Application URL
CORE_POCKETBASE_APP_URL=http://localhost:8090

# Default admin (auto-created on first run)
CORE_POCKETBASE_ADMIN_EMAIL=admin@localhost
CORE_POCKETBASE_ADMIN_PASSWORD=changeme123
```

### 2. Run PocketBase

```bash
# Via process compose (recommended - runs full stack)
go run ./cmd/core stack process

# Or directly
go run ./cmd/pocketbase serve
```

### 3. Access Services

- **Datastar Auth UI**: http://localhost:8090/ds
- **PocketBase Admin**: http://localhost:8090/_/
- **API Docs**: http://localhost:8090/api/

## 🔐 Authentication Pages

All pages use Datastar for reactive UI:

| Page | URL | Description |
|------|-----|-------------|
| Main Login | `/ds` | Password login + OAuth2 providers |
| Sign Up | `/ds/signup` | User registration |
| Email Verification | `/ds/verify-email` | Email confirmation (from email link) |
| Password Reset Request | `/ds/reset` | Request password reset link |
| Password Reset Confirm | `/ds/reset/confirm` | Set new password (from email link) |
| Account Settings | `/ds/settings` | Profile, password, email management |
| OAuth2 Callback | `/ds/callback` | OAuth2 redirect handler |

## 📡 API Endpoints

### Authentication
- `POST /api/ds/login` - Password login
- `POST /api/ds/refresh` - Refresh token
- `POST /api/ds/logout` - Logout
- `GET /api/ds/whoami` - Get current user
- `GET /api/ds/sse` - Server-sent events (auth state)

### OAuth2
- `GET /api/ds/auth-methods` - List available providers
- `GET /api/ds/oauth-start` - Start OAuth2 flow
- `POST /api/ds/auth-with-oauth2` - Complete OAuth2 login

### OTP/MFA
- `POST /api/ds/request-otp` - Request one-time password
- `POST /api/ds/auth-with-otp` - Authenticate with OTP

### Account Management
- `POST /api/ds/signup` - Create account
- `POST /api/ds/request-verification` - Request email verification
- `POST /api/ds/confirm-verification` - Confirm email
- `POST /api/ds/request-password-reset` - Request password reset
- `POST /api/ds/confirm-password-reset` - Confirm password reset
- `PATCH /api/ds/update-profile` - Update profile
- `POST /api/ds/change-password` - Change password
- `POST /api/ds/request-email-change` - Request email change
- `POST /api/ds/confirm-email-change` - Confirm email change
- `DELETE /api/ds/account` - Delete account

### Superuser
- `POST /api/ds/impersonate` - Impersonate user (admin only)

## ⚙️ Configuration

### SMTP Setup (Optional)

Required for email verification, password reset, and email change.

**Note:** In PocketBase v0.29+, SMTP is configured via the Admin UI at `http://localhost:8090/_/settings/mail`

Alternatively, you can set environment variables (values are logged on startup for reference):

```bash
CORE_POCKETBASE_SMTP_HOST=smtp.gmail.com
CORE_POCKETBASE_SMTP_PORT=587
CORE_POCKETBASE_SMTP_USERNAME=your-email@gmail.com
CORE_POCKETBASE_SMTP_PASSWORD=your-app-password
CORE_POCKETBASE_SMTP_FROM=noreply@yourapp.com
CORE_POCKETBASE_SMTP_TLS=true
```

**Gmail Setup:**
1. Enable 2-factor authentication
2. Generate app password: https://support.google.com/accounts/answer/185833
3. Use app password (not your regular password)

### OAuth2 Providers (Optional)

**Note:** OAuth2 providers are configured via the Admin UI at `http://localhost:8090/_/settings/auth-providers`

Environment variables can be set for reference:

```bash
# Google OAuth2
CORE_POCKETBASE_GOOGLE_CLIENT_ID=your-client-id
CORE_POCKETBASE_GOOGLE_CLIENT_SECRET=your-client-secret

# GitHub OAuth2
CORE_POCKETBASE_GITHUB_CLIENT_ID=your-client-id
CORE_POCKETBASE_GITHUB_CLIENT_SECRET=your-client-secret
```

**Google OAuth2:**
1. Create project: https://console.cloud.google.com/
2. Enable Google+ API
3. Create OAuth2 credentials
4. Add authorized redirect: `http://localhost:8090/api/oauth2-redirect`

**GitHub OAuth2:**
1. Register app: https://github.com/settings/developers
2. Set callback URL: `http://localhost:8090/api/oauth2-redirect`

## 🏗️ Architecture

```go
// Bootstrap: Auto-creates superuser from environment variables
app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
    return ensureAdminUser(app)
})

// Auth Routes: Registered via registerDatastarPocketBaseAuth()
app.OnServe().BindFunc(func(se *core.ServeEvent) error {
    registerDatastarPocketBaseAuth(app)
    return se.Next()
})
```

**Key Components:**
- **Bootstrap**: Auto-creates superuser from `CORE_POCKETBASE_ADMIN_*` env vars
- **Auth Routes**: Registered via `registerDatastarPocketBaseAuth()` in service.go
- **SSE Hub**: Manages real-time auth state updates per user
- **Token Storage**: Client-side in localStorage with Authorization header
- **Security**: Collection-level rules enforce user can only access own data

## 🛠️ CLI Usage

```bash
# Run with process compose (full stack)
go run ./cmd/core stack process

# Run standalone
go run ./cmd/pocketbase serve

# Ensure tooling binaries
core pocketbase ensure

# Inspect manifest metadata
core pocketbase spec
```

## 📝 Notes

- **PocketBase v0.29+**: Settings (SMTP, OAuth2) are configured via Admin UI, not programmatically
- **Superuser**: Auto-created on first run from `CORE_POCKETBASE_ADMIN_*` environment variables
- **SMTP Required**: For email verification and password reset features to work
- **OAuth2 Optional**: Works without OAuth2 - just password authentication
- **Development**: Default credentials are `admin@localhost` / `changeme123`
- **Production**: Change admin credentials immediately via Admin UI
