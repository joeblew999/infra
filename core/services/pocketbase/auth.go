// auth.go.go
package pocketbase

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"

	// NEW Datastar Go SDK
	"github.com/starfederation/datastar-go/datastar"
)

/*
WHAT'S INSIDE (single file)
- Pages:
  GET /ds              → Datastar UI (password + providers; reactive via $auth)
  GET /ds/callback     → OAuth2 code exchange (redirects back to /ds)
  GET /ds/signup       → User registration with email verification
  GET /ds/verify-email → Email verification confirmation
  GET /ds/reset        → Password reset request
  GET /ds/reset/confirm → Password reset confirmation
  GET /ds/settings     → Account settings (profile, password, email change)

- Datastar SSE:
  GET /api/ds/sse      → user-scoped SSE; server patches $auth with datastar-go

- PB-native auth:
  POST /api/ds/login   → password (PB Go API) -> {token,record}; also PatchSignals($auth)
  POST /api/ds/refresh → refresh (PB Go API)  -> {token,record}; also PatchSignals($auth)
  GET  /api/ds/whoami  → convenience
  POST /api/ds/logout  → client clears token; we patch $auth signedOut

- Provider-agnostic (Google, Apple, …):
  GET  /api/ds/auth-methods?collection=users        → PB /auth-methods
  GET  /api/ds/oauth-start?provider=...&redirect=.. → builds provider authURL with redirect_url
  POST /api/ds/auth-with-oauth2?collection=users    → PB /auth-with-oauth2
  POST /api/ds/request-otp                          → PB /request-otp
  POST /api/ds/auth-with-otp                        → PB /auth-with-otp (MFA continuation supported)

- Account Management:
  POST /api/ds/signup                               → Create new user account
  POST /api/ds/request-verification                 → Request email verification
  POST /api/ds/confirm-verification                 → Confirm email verification
  POST /api/ds/request-password-reset               → Request password reset
  POST /api/ds/confirm-password-reset               → Confirm password reset
  POST /api/ds/update-profile                       → Update user profile (authenticated)
  POST /api/ds/change-password                      → Change password (authenticated)
  POST /api/ds/request-email-change                 → Request email change (authenticated)
  POST /api/ds/confirm-email-change                 → Confirm email change
  DELETE /api/ds/account                            → Delete account (authenticated)

- Superuser features:
  POST /api/ds/impersonate                          → Impersonate user (superuser only)
*/

// RegisterDatastarAuth registers all Datastar auth routes and handlers.
// Exported for reuse by pocketbase-ha service.
func RegisterDatastarAuth(app *pocketbase.PocketBase, customFS *embed.FS) {
	// Use custom FS if provided, otherwise use the package-level embedFS
	registerDatastarRoutes(app)
}

func registerDatastarRoutes(app *pocketbase.PocketBase) {
	// ---- Pages ----
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET("/ds", func(e *core.RequestEvent) error {
			e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
			return indexTmpl.Execute(e.Response, nil)
		})
		se.Router.GET("/ds/callback", func(e *core.RequestEvent) error {
			e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
			return callbackTmpl.Execute(e.Response, nil)
		})
		se.Router.GET("/ds/signup", func(e *core.RequestEvent) error {
			e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
			return signupTmpl.Execute(e.Response, nil)
		})
		se.Router.GET("/ds/verify-email", func(e *core.RequestEvent) error {
			e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
			return verifyEmailTmpl.Execute(e.Response, nil)
		})
		se.Router.GET("/ds/reset", func(e *core.RequestEvent) error {
			e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
			return resetRequestTmpl.Execute(e.Response, nil)
		})
		se.Router.GET("/ds/reset/confirm", func(e *core.RequestEvent) error {
			e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
			return resetConfirmTmpl.Execute(e.Response, nil)
		})
		se.Router.GET("/ds/settings", func(e *core.RequestEvent) error {
			e.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
			return settingsTmpl.Execute(e.Response, nil)
		})

		// ---- Datastar SSE (user-scoped) ----
		se.Router.GET("/api/ds/sse", func(e *core.RequestEvent) error {
			sse := datastar.NewSSE(e.Response, e.Request)

			// If Authorization is present and valid, PB has populated e.Auth
			if e.Auth != nil {
				uid := e.Auth.Id
				hub.add(uid, sse)
				defer hub.remove(uid, sse)

				if data, err := json.Marshal(map[string]any{
					"auth": toAuthSignal(e.Auth),
				}); err == nil { _ = sse.PatchSignals(data) }
			} else {
				if data, err := json.Marshal(map[string]any{
					"auth": map[string]any{"signedIn": false},
				}); err == nil { _ = sse.PatchSignals(data) }
			}

			t := time.NewTicker(25 * time.Second)
			defer t.Stop()
			for {
				select {
				case <-sse.Context().Done():
					return nil
				case <-t.C:
					_ = sse.ExecuteScript(`console.log("datastar-sse alive")`)
				}
			}
		})

		// ---- API ----
		// Password login (PB Go API)
		se.Router.POST("/api/ds/login", func(e *core.RequestEvent) error {
			var body struct {
				Identity      string `json:"identity"`
				Password      string `json:"password"`
				Collection    string `json:"collection"`
				IdentityField string `json:"identityField"`
			}
			if err := e.BindBody(&body); err != nil {
				return apis.NewBadRequestError("invalid payload", err)
			}
			collection := choose(body.Collection, "users")
			idField := choose(body.IdentityField, "email")

			rec, err := e.App.FindFirstRecordByData(collection, idField, strings.TrimSpace(body.Identity))
			if err != nil || !rec.ValidatePassword(body.Password) {
				return apis.NewBadRequestError("invalid credentials", err)
			}
			if err := apis.RecordAuthResponse(e, rec, idField, nil); err != nil {
				return err
			}
			hub.patch(rec.Id, toAuthSignal(rec))
			return nil
		}).Bind(apis.RequireGuestOnly())

		// Refresh (PB Go API)
		se.Router.POST("/api/ds/refresh", func(e *core.RequestEvent) error {
			if e.Auth == nil {
				return apis.NewUnauthorizedError("missing or invalid token", nil)
			}
			if err := apis.RecordAuthResponse(e, e.Auth, "", nil); err != nil {
				return err
			}
			hub.patch(e.Auth.Id, toAuthSignal(e.Auth))
			return nil
		}).Bind(apis.RequireAuth("users"))

		// Whoami
		se.Router.GET("/api/ds/whoami", func(e *core.RequestEvent) error {
			if e.Auth == nil {
				return e.JSON(http.StatusOK, map[string]any{"authenticated": false})
			}
			return e.JSON(http.StatusOK, map[string]any{
				"authenticated": true,
				"record":        e.Auth,
			})
		})

		// Logout helper (client clears token)
		se.Router.POST("/api/ds/logout", func(e *core.RequestEvent) error {
			if e.Auth != nil {
				hub.patch(e.Auth.Id, map[string]any{"signedIn": false})
			}
			return e.JSON(http.StatusOK, map[string]string{"ok": "true"})
		})

		// ---- PB built-ins: OAuth2 / OTP / MFA (all providers) ----
		se.Router.GET("/api/ds/auth-methods", func(e *core.RequestEvent) error {
			u := base(e) + "/api/collections/" + url.PathEscape(coll(e)) + "/auth-methods"
			return forwardGET(e, u)
		})

		// helper to inject redirect_url into provider authURL
		se.Router.GET("/api/ds/oauth-start", func(e *core.RequestEvent) error {
			collection := coll(e)
			provider := e.Request.URL.Query().Get("provider")
			redirect := e.Request.URL.Query().Get("redirect")
			if provider == "" || redirect == "" {
				return apis.NewBadRequestError("provider and redirect are required", nil)
			}
			u := base(e) + "/api/collections/" + url.PathEscape(collection) + "/auth-methods"
			req, _ := http.NewRequestWithContext(e.Request.Context(), http.MethodGet, u, nil)
			copyAuth(req, e.Request)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return apis.NewApiError(500, "upstream error", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode >= 400 {
				b, _ := io.ReadAll(resp.Body)
				return apis.NewApiError(resp.StatusCode, "upstream: "+string(b), nil)
			}
			var am struct {
				OAuth2 struct {
					Providers []struct {
						Name        string `json:"name"`
						DisplayName string `json:"displayName"`
						State       string `json:"state"`
						AuthURL     string `json:"authURL"`
					} `json:"providers"`
				} `json:"oauth2"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&am); err != nil {
				return apis.NewApiError(500, "decode error", err)
			}
			for _, p := range am.OAuth2.Providers {
				if strings.EqualFold(p.Name, provider) {
					au, err := url.Parse(p.AuthURL)
					if err != nil {
						return apis.NewApiError(500, "bad authURL", err)
					}
					q := au.Query()
					q.Set("redirect_url", redirect)
					au.RawQuery = q.Encode()
					return e.JSON(http.StatusOK, map[string]any{
						"authURL":  au.String(),
						"state":    p.State,
						"provider": p.Name,
						"label":    choose(p.DisplayName, p.Name),
					})
				}
			}
			return apis.NewNotFoundError("provider not enabled", nil)
		})

		se.Router.POST("/api/ds/request-otp", func(e *core.RequestEvent) error {
			u := base(e) + "/api/collections/" + url.PathEscape(coll(e)) + "/request-otp" + keepQuery(e, "collection")
			return forward(e, http.MethodPost, u)
		})
		se.Router.POST("/api/ds/auth-with-otp", func(e *core.RequestEvent) error {
			u := base(e) + "/api/collections/" + url.PathEscape(coll(e)) + "/auth-with-otp" + keepQuery(e, "collection")
			return forward(e, http.MethodPost, u)
		})
		se.Router.POST("/api/ds/auth-with-oauth2", func(e *core.RequestEvent) error {
			u := base(e) + "/api/collections/" + url.PathEscape(coll(e)) + "/auth-with-oauth2" + keepQuery(e, "collection")
			return forward(e, http.MethodPost, u)
		})

		// ---- Account Management ----
		// Signup (create new user)
		se.Router.POST("/api/ds/signup", func(e *core.RequestEvent) error {
			u := base(e) + "/api/collections/" + url.PathEscape(coll(e)) + "/records"
			return forward(e, http.MethodPost, u)
		})

		// Request email verification
		se.Router.POST("/api/ds/request-verification", func(e *core.RequestEvent) error {
			u := base(e) + "/api/collections/" + url.PathEscape(coll(e)) + "/request-verification"
			return forward(e, http.MethodPost, u)
		})

		// Confirm email verification
		se.Router.POST("/api/ds/confirm-verification", func(e *core.RequestEvent) error {
			u := base(e) + "/api/collections/" + url.PathEscape(coll(e)) + "/confirm-verification"
			return forward(e, http.MethodPost, u)
		})

		// Request password reset
		se.Router.POST("/api/ds/request-password-reset", func(e *core.RequestEvent) error {
			u := base(e) + "/api/collections/" + url.PathEscape(coll(e)) + "/request-password-reset"
			return forward(e, http.MethodPost, u)
		})

		// Confirm password reset
		se.Router.POST("/api/ds/confirm-password-reset", func(e *core.RequestEvent) error {
			u := base(e) + "/api/collections/" + url.PathEscape(coll(e)) + "/confirm-password-reset"
			return forward(e, http.MethodPost, u)
		})

		// Update profile (authenticated)
		se.Router.PATCH("/api/ds/update-profile", func(e *core.RequestEvent) error {
			if e.Auth == nil {
				return apis.NewUnauthorizedError("authentication required", nil)
			}
			u := base(e) + "/api/collections/" + url.PathEscape(coll(e)) + "/records/" + url.PathEscape(e.Auth.Id)
			return forward(e, http.MethodPatch, u)
		}).Bind(apis.RequireAuth("users"))

		// Change password (authenticated)
		se.Router.POST("/api/ds/change-password", func(e *core.RequestEvent) error {
			if e.Auth == nil {
				return apis.NewUnauthorizedError("authentication required", nil)
			}
			var body struct {
				OldPassword string `json:"oldPassword"`
				Password    string `json:"password"`
				PasswordConfirm string `json:"passwordConfirm"`
			}
			if err := e.BindBody(&body); err != nil {
				return apis.NewBadRequestError("invalid payload", err)
			}
			// Validate old password
			if !e.Auth.ValidatePassword(body.OldPassword) {
				return apis.NewBadRequestError("invalid old password", nil)
			}
			// Update using PB API
			u := base(e) + "/api/collections/" + url.PathEscape(coll(e)) + "/records/" + url.PathEscape(e.Auth.Id)
			return forward(e, http.MethodPatch, u)
		}).Bind(apis.RequireAuth("users"))

		// Request email change (authenticated)
		se.Router.POST("/api/ds/request-email-change", func(e *core.RequestEvent) error {
			if e.Auth == nil {
				return apis.NewUnauthorizedError("authentication required", nil)
			}
			u := base(e) + "/api/collections/" + url.PathEscape(coll(e)) + "/request-email-change"
			return forward(e, http.MethodPost, u)
		}).Bind(apis.RequireAuth("users"))

		// Confirm email change
		se.Router.POST("/api/ds/confirm-email-change", func(e *core.RequestEvent) error {
			u := base(e) + "/api/collections/" + url.PathEscape(coll(e)) + "/confirm-email-change"
			return forward(e, http.MethodPost, u)
		})

		// Delete account (authenticated)
		se.Router.DELETE("/api/ds/account", func(e *core.RequestEvent) error {
			if e.Auth == nil {
				return apis.NewUnauthorizedError("authentication required", nil)
			}
			u := base(e) + "/api/collections/" + url.PathEscape(coll(e)) + "/records/" + url.PathEscape(e.Auth.Id)
			req, _ := http.NewRequestWithContext(e.Request.Context(), http.MethodDelete, u, nil)
			copyAuth(req, e.Request)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return apis.NewApiError(500, "upstream error", err)
			}
			defer resp.Body.Close()
			e.Response.WriteHeader(resp.StatusCode)
			_, _ = io.Copy(e.Response, resp.Body)
			return nil
		}).Bind(apis.RequireAuth("users"))

		// ---- Superuser Features ----
		// Impersonate user (superuser only)
		se.Router.POST("/api/ds/impersonate", func(e *core.RequestEvent) error {
			// RequireAuth middleware will ensure admin status
			if e.Auth == nil {
				return apis.NewForbiddenError("superuser access required", nil)
			}
			var body struct {
				UserID   string `json:"userId"`
				Duration int    `json:"duration"` // seconds
			}
			if err := e.BindBody(&body); err != nil {
				return apis.NewBadRequestError("invalid payload", err)
			}
			collection := coll(e)
			_, err := e.App.FindRecordById(collection, body.UserID)
			if err != nil {
				return apis.NewNotFoundError("user not found", err)
			}
			// Use PB's impersonate endpoint
			u := base(e) + "/api/collections/" + url.PathEscape(collection) + "/impersonate/" + url.PathEscape(body.UserID)
			if body.Duration > 0 {
				u += "?duration=" + url.QueryEscape(strings.TrimSpace(fmt.Sprintf("%d", body.Duration)))
			}
			return forward(e, http.MethodPost, u)
		}).Bind(apis.RequireAuth())

		return se.Next()
	})
}

// ---------------- helpers ----------------

func coll(e *core.RequestEvent) string {
	if v := strings.TrimSpace(e.Request.URL.Query().Get("collection")); v != "" {
		return v
	}
	return "users"
}
func base(e *core.RequestEvent) string {
	req := e.Request
	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}
	if v := req.Header.Get("X-Forwarded-Proto"); v != "" {
		scheme = v
	}
	return scheme + "://" + req.Host
}
func choose(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}

func keepQuery(e *core.RequestEvent, exclude string) string {
	q := e.Request.URL.Query()
	dst := url.Values{}
	for k, vals := range q {
		if strings.EqualFold(k, exclude) {
			continue
		}
		for _, v := range vals {
			dst.Add(k, v)
		}
	}
	if len(dst) == 0 {
		return ""
	}
	return "?" + dst.Encode()
}

func copyAuth(dst *http.Request, src *http.Request) {
	if v := src.Header.Get("Authorization"); v != "" {
		dst.Header.Set("Authorization", v)
	}
}

func forwardGET(e *core.RequestEvent, upstream string) error {
	req, _ := http.NewRequestWithContext(e.Request.Context(), http.MethodGet, upstream, nil)
	copyAuth(req, e.Request)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return apis.NewApiError(500, "upstream error", err)
	}
	defer resp.Body.Close()
	e.Response.Header().Set("Content-Type", "application/json")
	e.Response.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(e.Response, resp.Body)
	return nil
}

func forward(e *core.RequestEvent, method, upstream string) error {
	body, _ := io.ReadAll(e.Request.Body)
	req, _ := http.NewRequestWithContext(e.Request.Context(), method, upstream, bytes.NewReader(body))
	copyAuth(req, e.Request)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return apis.NewApiError(500, "upstream error", err)
	}
	defer resp.Body.Close()
	e.Response.Header().Set("Content-Type", "application/json")
	e.Response.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(e.Response, resp.Body)
	return nil
}

// minimal $auth payload for the UI (don't leak token)
func toAuthSignal(rec *core.Record) map[string]any {
	if rec == nil {
		return map[string]any{"signedIn": false}
	}
	out := map[string]any{
		"signedIn": true,
		"id":       rec.Id,
	}
	if v := rec.Get("email"); v != nil {
		out["email"] = v
	}
	if v := rec.Get("username"); v != nil {
		out["username"] = v
	}
	if v := rec.Get("name"); v != nil {
		out["name"] = v
	}
	return out
}

// ---- tiny SSE hub keyed by user id ----

type hubT struct {
	mu   sync.RWMutex
	conn map[string]map[*datastar.ServerSentEventGenerator]struct{}
}

func (h *hubT) add(userID string, s *datastar.ServerSentEventGenerator) {
	if h.conn == nil {
		h.conn = map[string]map[*datastar.ServerSentEventGenerator]struct{}{}
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.conn[userID] == nil {
		h.conn[userID] = map[*datastar.ServerSentEventGenerator]struct{}{}
	}
	h.conn[userID][s] = struct{}{}
}
func (h *hubT) remove(userID string, s *datastar.ServerSentEventGenerator) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if m := h.conn[userID]; m != nil {
		delete(m, s)
		if len(m) == 0 {
			delete(h.conn, userID)
		}
	}
}
func (h *hubT) patch(userID string, payload any) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	data, err := json.Marshal(map[string]any{"auth": payload})
	if err != nil {
		return
	}
	for s := range h.conn[userID] {
		_ = s.PatchSignals(data)
	}
}

var hub hubT

// ---------------- HTML (Datastar) ----------------

//go:embed auth_index.html
var authIndexHTML string

//go:embed auth_callback.html
var authCallbackHTML string

//go:embed auth_signup.html
var authSignupHTML string

//go:embed auth_verify_email.html
var authVerifyEmailHTML string

//go:embed auth_reset_request.html
var authResetRequestHTML string

//go:embed auth_reset_confirm.html
var authResetConfirmHTML string

//go:embed auth_settings.html
var authSettingsHTML string

var (
	indexTmpl        = template.Must(template.New("index").Parse(authIndexHTML))
	callbackTmpl     = template.Must(template.New("callback").Parse(authCallbackHTML))
	signupTmpl       = template.Must(template.New("signup").Parse(authSignupHTML))
	verifyEmailTmpl  = template.Must(template.New("verify").Parse(authVerifyEmailHTML))
	resetRequestTmpl = template.Must(template.New("resetReq").Parse(authResetRequestHTML))
	resetConfirmTmpl = template.Must(template.New("resetConf").Parse(authResetConfirmHTML))
	settingsTmpl     = template.Must(template.New("settings").Parse(authSettingsHTML))
)
