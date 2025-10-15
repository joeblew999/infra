package cloudflare

import (
	"context"
	"testing"

	cf "github.com/cloudflare/cloudflare-go"
)

func TestSelectPermissionGroups(t *testing.T) {
	groups := []cf.APITokenPermissionGroups{
		{ID: "1", Name: "Zone Read"},
		{ID: "2", Name: "DNS Edit"},
		{ID: "3", Name: "R2 Storage Edit"},
	}

	selection, err := SelectPermissionGroups(groups, true)
	if err != nil {
		t.Fatalf("SelectPermissionGroups returned error: %v", err)
	}
	if selection.ZoneRead.ID != "1" {
		t.Fatalf("expected zone read ID 1, got %q", selection.ZoneRead.ID)
	}
	if selection.DNSEdit.ID != "2" {
		t.Fatalf("expected dns edit ID 2, got %q", selection.DNSEdit.ID)
	}
	if selection.R2Edit == nil || selection.R2Edit.ID != "3" {
		t.Fatalf("expected r2 edit ID 3, got %#v", selection.R2Edit)
	}

	selectionNoR2, err := SelectPermissionGroups(groups, false)
	if err != nil {
		t.Fatalf("SelectPermissionGroups without R2 returned error: %v", err)
	}
	if selectionNoR2.R2Edit != nil {
		t.Fatalf("expected no R2 group when includeR2=false")
	}
}

func TestCreateScopedTokenMissingParams(t *testing.T) {
	_, err := CreateScopedToken(context.Background(), nil, CreateScopedTokenParams{})
	if err == nil {
		t.Fatal("expected error when API client is nil")
	}

	client, _ := cf.New("dummy", "user@example.com")
	_, err = CreateScopedToken(context.Background(), client, CreateScopedTokenParams{})
	if err == nil {
		t.Fatal("expected error when account/zone IDs missing")
	}
}
