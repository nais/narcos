package jita

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/nais/naistrix"
	"github.com/nais/naistrix/output"
	"github.com/nais/narcos/internal/gcp"
	"github.com/nais/narcos/internal/jita/command/flag"
	"golang.org/x/sync/errgroup"
)

type Entitlement struct {
	Tenant        string
	Entitlement   string        `heading:"Entitlement"`
	TimeRemaining string        `heading:"Time remaining"`
	MaxDuration   time.Duration `heading:"Max. duration"`
	Roles         RoleList      `hidden:"true"`
}

type RoleList []string

func (r RoleList) String() string {
	return strings.Join(r, "\n")
}

func List(ctx context.Context, flags *flag.List, out *naistrix.OutputWriter) error {
	username, err := gcp.GCloudActiveUser(ctx)
	if err != nil {
		return err
	}

	tenants := flags.Tenants
	if len(tenants) == 0 {
		var err error
		tenants, err = gcp.FetchAllTenantNames(ctx)
		if err != nil {
			return err
		}
	}

	allEntitlements := make([]*Entitlement, 0)
	var mu sync.Mutex

	eg, ctx := errgroup.WithContext(ctx)
	for _, tenant := range tenants {
		eg.Go(func() error {
			entitlements, err := getEntitlementsForTenant(ctx, username, tenant)
			if err != nil {
				return err
			}

			mu.Lock()
			allEntitlements = append(allEntitlements, entitlements...)
			mu.Unlock()
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	slices.SortStableFunc(allEntitlements, func(a, b *Entitlement) int {
		if a.TimeRemaining != b.TimeRemaining {
			return cmp.Compare(b.TimeRemaining, a.TimeRemaining)
		}

		if a.Tenant != b.Tenant {
			return cmp.Compare(a.Tenant, b.Tenant)
		}

		return cmp.Compare(a.Entitlement, b.Entitlement)
	})

	opts := make([]output.TableOptionFunc, 0)
	if flags.IsVerbose() {
		opts = append(opts, output.TableWithShowHiddenColumns())
	}

	return out.
		Table(opts...).
		Render(allEntitlements)
}

func getEntitlementsForTenant(ctx context.Context, username, tenant string) ([]*Entitlement, error) {
	metadata, err := gcp.FetchTenantMetadata(ctx, tenant)
	if err != nil {
		return nil, fmt.Errorf("GCP error fetching tenant metadata: %w", err)
	}

	resp, err := gcp.ListEntitlements(ctx, metadata.NaisFolderID)
	if err != nil {
		return nil, fmt.Errorf("GCP error listing entitlements: %w", err)
	}

	entitlements := make([]*Entitlement, len(resp.Entitlements))
	for i, ent := range resp.Entitlements {
		grants, err := ent.ListActiveGrants(ctx, username)
		if err != nil {
			return nil, fmt.Errorf("fetching active grants: %w", err)
		}

		e := &Entitlement{
			Tenant:      tenant,
			Entitlement: ent.ShortName(),
			MaxDuration: ent.MaxDuration(),
			Roles: func() RoleList {
				roles := ent.Roles()
				slices.Sort(roles)
				return roles
			}(),
		}

		if len(grants) > 0 {
			e.TimeRemaining = grants[0].TimeRemaining().String()
		}

		entitlements[i] = e
	}

	return entitlements, nil
}
