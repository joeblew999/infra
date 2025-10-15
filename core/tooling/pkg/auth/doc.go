// Package auth provides authentication workflows for external services.
//
// This package handles the complete authentication lifecycle for Fly.io and Cloudflare,
// including token acquisition, verification, permission checking, and secure storage.
//
// # Cloudflare Authentication
//
// Cloudflare supports TWO authentication methods, both producing identical results
// but differing in how the initial token is obtained:
//
// 1. Manual Token Method (Recommended):
//   - User creates a scoped API token in the Cloudflare dashboard
//   - Most secure: user controls exact permissions granted
//   - Implemented in: RunCloudflareAuth()
//   - Token stored with kind: TokenKindManual
//
// 2. Bootstrap Method (Convenience):
//   - Uses Global API Key ONCE to programmatically create a scoped token
//   - Global API Key is NOT stored (only the generated scoped token)
//   - Implemented in: RunCloudflareBootstrap()
//   - Token stored with kind: TokenKindBootstrap
//
// Both methods result in a scoped API token stored at the same location.
// The "kind" tracking (manual/bootstrap) is metadata only - both tokens
// function identically after creation.
//
// # Token Storage Strategy
//
// To prevent conflicts between authentication methods, tokens are stored with:
//   - Main location: .data/core/secrets/cloudflare/api_token (active token)
//   - Method tracking: .data/core/cloudflare/tokens/{manual|bootstrap}
//   - Active pointer: .data/core/cloudflare/active_token
//
// Last authentication wins - if you auth with bootstrap then manual, the
// manual token becomes active. This is expected behavior.
//
// # Permission Verification
//
// After token verification, the package tests actual API access:
//   - Zone:Zone:Read - REQUIRED (auth fails if missing)
//   - Zone:DNS:Edit - Optional (warns if missing)
//   - Account:R2:Edit - Optional (warns if missing)
//
// This ensures tokens are not just valid, but actually usable for the
// intended operations. Users can choose to continue with limited permissions.
//
// # User Experience Design
//
// Authentication flows are designed to:
//   1. Check existing credentials first (idempotency)
//   2. Show clear instructions BEFORE taking action
//   3. Wait for user confirmation before opening browsers
//   4. Provide detailed feedback at each step
//   5. Distinguish between failures (✗) and warnings (⚠)
//   6. Allow users to make informed decisions
//
// # Fly.io Authentication
//
// Fly.io uses a simpler single-method approach:
//   - Interactive browser-based OAuth flow
//   - Or direct token input via --token flag
//   - Token verification against Fly API
//   - Optional app connectivity test
//
// # Error Handling Philosophy
//
// Errors are categorized as:
//   - Fatal: Missing required permissions, invalid credentials
//   - Warnings: Missing optional features, unreachable resources
//   - User choice: Continue with warnings or cancel
//
// This allows the CLI to work in partial-feature scenarios while still
// ensuring critical functionality is available.
package auth
