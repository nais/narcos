package fasit

import (
	"context"
	"fmt"

	fasitgraphql "github.com/nais/narcos/internal/fasit/graphql"
)

func GetEnvironment(ctx context.Context, tenantSlug, envName string) (*Environment, error) {
	store, err := newStore(ctx)
	if err != nil {
		return nil, err
	}

	_, env, err := store.getTenantEnvironment(ctx, tenantSlug, envName)
	if err != nil {
		return nil, err
	}

	return env, nil
}

func GetEnvFeature(ctx context.Context, tenantSlug, envName, featureName string) (*Environment, *Feature, error) {
	store, err := newStore(ctx)
	if err != nil {
		return nil, nil, err
	}

	return store.getEnvFeature(ctx, tenantSlug, envName, featureName)
}

func (s *store) getTenantEnvironment(ctx context.Context, tenantSlug, envName string) (*Tenant, *Environment, error) {
	resp, err := fasitgraphql.GetTenantEnvironment(ctx, s.client, tenantSlug, envName)
	if err != nil {
		return nil, nil, fmt.Errorf("get tenant environment: %w", err)
	}

	tenant := &Tenant{
		ID:   resp.Tenant.Id,
		Name: resp.Tenant.Name,
		Icon: tenantIcon(resp.Tenant.Name),
	}

	env := convertEnvironment(resp.Tenant.Environment)
	if env == nil {
		return nil, nil, fmt.Errorf("not found: environment %s/%s", tenantSlug, envName)
	}

	return tenant, env, nil
}

func (s *store) getEnvFeature(ctx context.Context, tenantSlug, envName, featureName string) (*Environment, *Feature, error) {
	_, env, err := s.getTenantEnvironment(ctx, tenantSlug, envName)
	if err != nil {
		return nil, nil, err
	}

	for i := range env.Features {
		if env.Features[i].Name != featureName {
			continue
		}

		configuration, err := s.fetchFeatureConfig(ctx, featureName, env.ID)
		if err != nil {
			return nil, nil, err
		}

		env.Features[i].Configuration = configuration
		return env, &env.Features[i], nil
	}

	return nil, nil, fmt.Errorf("not found: feature %s in environment %s/%s", featureName, tenantSlug, envName)
}

func convertEnvironment(gql *fasitgraphql.GetTenantEnvironmentTenantEnvironment) *Environment {
	if gql == nil {
		return nil
	}

	features := make([]Feature, 0, len(gql.Features))
	for _, feature := range gql.Features {
		if feature == nil {
			continue
		}

		features = append(features, Feature{
			Name:             feature.Name,
			EnvironmentKinds: toStringSlice(feature.EnvironmentKinds),
			Enabled:          feature.State != nil && feature.State.Enabled,
		})
	}

	values := make([]EnvironmentValue, 0, len(gql.Values))
	for _, value := range gql.Values {
		if value == nil {
			continue
		}

		values = append(values, EnvironmentValue{
			Key:   value.Key,
			Value: string(value.Value),
		})
	}

	return &Environment{
		ID:           gql.Id,
		Name:         gql.Name,
		Description:  gql.Description,
		Created:      gql.Created,
		LastModified: gql.LastModified,
		Kind:         string(gql.Kind),
		GCPProjectID: gql.GcpProjectID,
		Reconcile:    gql.Reconcile,
		Features:     features,
		Values:       values,
	}
}
