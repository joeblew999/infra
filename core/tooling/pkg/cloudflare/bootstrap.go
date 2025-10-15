package cloudflare

import (
	"context"
	"fmt"
	"strings"

	cf "github.com/cloudflare/cloudflare-go"
)

// PermissionSelection contains the permission groups required to operate the
// tooling workflow. R2 permissions are optional and populated when requested.
type PermissionSelection struct {
	ZoneRead cf.APITokenPermissionGroups
	DNSEdit  cf.APITokenPermissionGroups
	R2Edit   *cf.APITokenPermissionGroups
}

// SelectPermissionGroups discovers the required permission groups from the set
// returned by Cloudflare. It prefers exact name matches but falls back to
// substring checks to tolerate minor naming changes.
func SelectPermissionGroups(groups []cf.APITokenPermissionGroups, includeR2 bool) (PermissionSelection, error) {
	var selection PermissionSelection

	match := func(target string, candidates ...string) bool {
		target = strings.ToLower(strings.TrimSpace(target))
		for _, cand := range candidates {
			cand = strings.ToLower(strings.TrimSpace(cand))
			if cand == target {
				return true
			}
		}
		return false
	}

	for _, pg := range groups {
		name := strings.ToLower(pg.Name)
		switch {
		case match(name, "zone read", "zone:read") || (strings.Contains(name, "zone") && strings.Contains(name, "read")):
			selection.ZoneRead = pg
		case match(name, "dns edit", "dns:edit", "zone dns settings write", "zone write") || (strings.Contains(name, "dns") && strings.Contains(name, "write")):
			selection.DNSEdit = pg
		case includeR2 && (match(name, "r2 storage edit", "r2 storage write", "workers r2 storage write", "r2 write") || (strings.Contains(name, "r2") && strings.Contains(name, "write"))):
			pgCopy := pg
			selection.R2Edit = &pgCopy
		}
	}

	if selection.ZoneRead.ID == "" {
		return PermissionSelection{}, fmt.Errorf("cloudflare bootstrap: unable to locate 'Zone Read' permission group")
	}
	if selection.DNSEdit.ID == "" {
		return PermissionSelection{}, fmt.Errorf("cloudflare bootstrap: unable to locate 'DNS Edit' permission group")
	}
	if includeR2 && (selection.R2Edit == nil || selection.R2Edit.ID == "") {
		return PermissionSelection{}, fmt.Errorf("cloudflare bootstrap: unable to locate 'R2 Edit' permission group")
	}
	return selection, nil
}

// CreateScopedTokenParams describes the scoped token we wish to create.
type CreateScopedTokenParams struct {
	Name          string
	AccountID     string
	ZoneID        string
	ZoneName      string
	IncludeR2     bool
	PermissionSet PermissionSelection
}

// CreateScopedToken uses the provided (elevated) Cloudflare API client to
// create a scoped token with the minimal permissions required by the tooling
// workflow. The newly generated token is returned.
func CreateScopedToken(ctx context.Context, api *cf.API, params CreateScopedTokenParams) (cf.APIToken, error) {
	if api == nil {
		return cf.APIToken{}, fmt.Errorf("cloudflare bootstrap: API client is nil")
	}
	if strings.TrimSpace(params.AccountID) == "" || strings.TrimSpace(params.ZoneID) == "" {
		return cf.APIToken{}, fmt.Errorf("cloudflare bootstrap: account or zone not provided")
	}

	resources := map[string]interface{}{
		"com.cloudflare.api.account.zone.*": "*",
	}
	if params.IncludeR2 {
		resources["com.cloudflare.api.account.*"] = "*"
	}

	permissionGroups := []cf.APITokenPermissionGroups{
		params.PermissionSet.ZoneRead,
		params.PermissionSet.DNSEdit,
	}
	if params.IncludeR2 && params.PermissionSet.R2Edit != nil {
		permissionGroups = append(permissionGroups, *params.PermissionSet.R2Edit)
	}

	token := cf.APIToken{
		Name: params.Name,
		Policies: []cf.APITokenPolicies{
			{
				Effect:           "allow",
				Resources:        resources,
				PermissionGroups: permissionGroups,
			},
		},
	}

	created, err := api.CreateAPIToken(ctx, token)
	if err != nil {
		return cf.APIToken{}, fmt.Errorf("cloudflare bootstrap: create token: %w", err)
	}
	if strings.TrimSpace(created.Value) == "" {
		return cf.APIToken{}, fmt.Errorf("cloudflare bootstrap: API did not return token value; ensure the global key has token-create permissions")
	}
	return created, nil
}
