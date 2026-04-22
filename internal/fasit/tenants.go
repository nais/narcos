package fasit

import (
	"context"
	"fmt"

	genqlientgraphql "github.com/Khan/genqlient/graphql"
	fasitgraphql "github.com/nais/narcos/internal/fasit/graphql"
)

type store struct {
	client genqlientgraphql.Client
}

func newStore(ctx context.Context) (*store, error) {
	client, err := fasitgraphql.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return &store{client: client}, nil
}

func ListTenants(ctx context.Context) ([]Tenant, error) {
	store, err := newStore(ctx)
	if err != nil {
		return nil, err
	}

	return store.listTenants(ctx)
}

func GetTenant(ctx context.Context, slug string) (*Tenant, error) {
	store, err := newStore(ctx)
	if err != nil {
		return nil, err
	}

	return store.getTenant(ctx, slug)
}

func (s *store) listTenants(ctx context.Context) ([]Tenant, error) {
	resp, err := fasitgraphql.GetTenants(ctx, s.client)
	if err != nil {
		return nil, fmt.Errorf("get tenants: %w", err)
	}

	tenants := make([]Tenant, 0, len(resp.Tenants))
	for _, tenant := range resp.Tenants {
		if tenant == nil {
			continue
		}

		tenants = append(tenants, *convertTenant(tenant))
	}

	return tenants, nil
}

func (s *store) getTenant(ctx context.Context, slug string) (*Tenant, error) {
	resp, err := fasitgraphql.GetTenantOverview(ctx, s.client, slug)
	if err != nil {
		return nil, fmt.Errorf("get tenant overview: %w", err)
	}

	return convertTenantOverview(resp.Tenant), nil
}

func convertTenant(gql *fasitgraphql.GetTenantsTenantsTenant) *Tenant {
	if gql == nil {
		return &Tenant{}
	}

	environments := make([]Environment, 0, len(gql.Environments))
	for _, environment := range gql.Environments {
		if environment == nil {
			continue
		}

		environments = append(environments, Environment{
			ID:           environment.Id,
			Name:         environment.Name,
			Description:  environment.Description,
			Created:      environment.Created,
			LastModified: environment.LastModified,
			Kind:         string(environment.Kind),
			GCPProjectID: environment.GcpProjectID,
			Reconcile:    environment.Reconcile,
		})
	}

	return &Tenant{
		ID:           gql.Id,
		Name:         gql.Name,
		Environments: environments,
		Icon:         tenantIcon(gql.Name),
	}
}

func convertTenantOverview(gql *fasitgraphql.GetTenantOverviewTenant) *Tenant {
	if gql == nil {
		return &Tenant{}
	}

	environments := make([]Environment, 0, len(gql.Environments))
	for _, environment := range gql.Environments {
		if environment == nil {
			continue
		}

		features := make([]Feature, 0, len(environment.Features))
		for _, feature := range environment.Features {
			if feature == nil {
				continue
			}

			features = append(features, Feature{Name: feature.Name})
		}

		environments = append(environments, Environment{
			ID:           environment.Id,
			Name:         environment.Name,
			Description:  environment.Description,
			Created:      environment.Created,
			LastModified: environment.LastModified,
			Kind:         string(environment.Kind),
			GCPProjectID: environment.GcpProjectID,
			Reconcile:    environment.Reconcile,
			Features:     features,
		})
	}

	return &Tenant{
		ID:           gql.Id,
		Name:         gql.Name,
		Environments: environments,
		Icon:         tenantIcon(gql.Name),
	}
}

func tenantIcon(name string) string {
	icons := map[string]string{
		"nav":      "🧭",
		"devnais":  "🛠️",
		"testnais": "🧪",
		"cinais":   "🔒",
		"atil":     "📡",
		"ssb":      "📊",
		"ldir":     "🌿",
	}

	if icon, ok := icons[name]; ok {
		return icon
	}

	return "🏢"
}

func toStringSlice[T ~string](values []T) []string {
	out := make([]string, len(values))
	for i, value := range values {
		out[i] = string(value)
	}

	return out
}
