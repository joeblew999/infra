package cloudflare

import (
	"context"
	"fmt"
	"log"
	"strings"

	cf "github.com/cloudflare/cloudflare-go"

	"github.com/joeblew999/infra/core/controller/pkg/reconcile"
	controllerspec "github.com/joeblew999/infra/core/controller/pkg/spec"
)

// Provider reconciles DNS state in Cloudflare for services configured with the
// cloudflare routing provider.
type Provider struct {
	api *cf.API
}

// New constructs a Cloudflare routing provider using the supplied API token.
// The token must have DNS edit permissions for the zones being managed.
func New(token string) (*Provider, error) {
	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("cloudflare: token is required")
	}
	api, err := cf.NewWithAPIToken(token)
	if err != nil {
		return nil, fmt.Errorf("cloudflare: %w", err)
	}
	return &Provider{api: api}, nil
}

// EnsureRouting satisfies reconcile.RoutingProvider.
func (p *Provider) EnsureRouting(ctx context.Context, svc controllerspec.Service, runtime reconcile.ServiceRuntimeState) error {
	if !strings.EqualFold(svc.Routing.Provider, "cloudflare") {
		return nil
	}
	if svc.Routing.Zone == "" {
		return fmt.Errorf("service %s missing routing zone", svc.ID)
	}

	zoneID, err := p.api.ZoneIDByName(svc.Routing.Zone)
	if err != nil {
		return fmt.Errorf("cloudflare: resolve zone %s: %w", svc.Routing.Zone, err)
	}
	zone := cf.ZoneIdentifier(zoneID)

	for _, record := range svc.Routing.DNSRecords {
		if record.Name == "" || record.Type == "" {
			log.Printf("[cloudflare] service=%s skipping incomplete record: %+v", svc.ID, record)
			continue
		}
		fqdn := record.Name
		if !strings.HasSuffix(strings.ToLower(fqdn), strings.ToLower(svc.Routing.Zone)) {
			fqdn = fmt.Sprintf("%s.%s", strings.TrimSuffix(record.Name, "."), svc.Routing.Zone)
		}
		desiredContent := strings.TrimSpace(record.Content)
		if desiredContent == "" {
			log.Printf("[cloudflare] service=%s record=%s missing content; skipping", svc.ID, fqdn)
			continue
		}

		params := cf.ListDNSRecordsParams{Name: fqdn, Type: record.Type}
		matches, _, err := p.api.ListDNSRecords(ctx, zone, params)
		if err != nil {
			return fmt.Errorf("cloudflare: list records for %s: %w", fqdn, err)
		}

		ttl := record.TTL
		if ttl < 0 {
			ttl = 0
		}

		if len(matches) == 0 {
			_, err := p.api.CreateDNSRecord(ctx, zone, cf.CreateDNSRecordParams{
				Type:    record.Type,
				Name:    fqdn,
				Content: desiredContent,
				TTL:     ttl,
			})
			if err != nil {
				return fmt.Errorf("cloudflare: create %s %s: %w", record.Type, fqdn, err)
			}
			log.Printf("[cloudflare] created %s %s -> %s", record.Type, fqdn, desiredContent)
			continue
		}

		current := matches[0]
		if current.Content == desiredContent && (ttl == 0 || current.TTL == ttl) {
			continue
		}

		updateParams := cf.UpdateDNSRecordParams{
			ID:      current.ID,
			Type:    record.Type,
			Name:    fqdn,
			Content: desiredContent,
			TTL:     ttl,
		}
		_, err = p.api.UpdateDNSRecord(ctx, zone, updateParams)
		if err != nil {
			return fmt.Errorf("cloudflare: update %s %s: %w", record.Type, fqdn, err)
		}
		log.Printf("[cloudflare] updated %s %s -> %s", record.Type, fqdn, desiredContent)
	}
	return nil
}
