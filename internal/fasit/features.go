package fasit

import (
	"context"
	"encoding/json"
	"fmt"

	fasitgraphql "github.com/nais/narcos/internal/fasit/graphql"
)

func ListFeatures(ctx context.Context) ([]Feature, error) {
	store, err := newStore(ctx)
	if err != nil {
		return nil, err
	}

	return store.listFeatures(ctx)
}

func GetFeature(ctx context.Context, name string) (*Feature, error) {
	store, err := newStore(ctx)
	if err != nil {
		return nil, err
	}

	return store.getFeature(ctx, name)
}

func GetFeatureStatus(ctx context.Context, name string) ([]FeatureStatus, error) {
	store, err := newStore(ctx)
	if err != nil {
		return nil, err
	}

	return store.getFeatureStatus(ctx, name)
}

func (s *store) listFeatures(ctx context.Context) ([]Feature, error) {
	resp, err := fasitgraphql.GetFeatures(ctx, s.client)
	if err != nil {
		return nil, fmt.Errorf("get features: %w", err)
	}

	features := make([]Feature, 0, len(resp.Features))
	for _, feature := range resp.Features {
		if feature == nil {
			continue
		}

		features = append(features, *convertFeature(feature))
	}

	return features, nil
}

func (s *store) getFeature(ctx context.Context, name string) (*Feature, error) {
	features, err := s.listFeatures(ctx)
	if err != nil {
		return nil, err
	}

	for i := range features {
		if features[i].Name != name {
			continue
		}

		configuration, err := s.fetchFeatureConfig(ctx, name, "")
		if err != nil {
			return nil, err
		}

		features[i].Configuration = configuration
		features[i].Configurations = flattenConfigurationDetails(configuration)
		return &features[i], nil
	}

	return nil, fmt.Errorf("not found: feature %s", name)
}

func (s *store) getFeatureStatus(ctx context.Context, name string) ([]FeatureStatus, error) {
	resp, err := fasitgraphql.GetFeatureStatus(ctx, s.client)
	if err != nil {
		return nil, fmt.Errorf("get feature status: %w", err)
	}

	statuses := make([]FeatureStatus, 0)
	for _, tenant := range resp.Tenants {
		if tenant == nil {
			continue
		}

		for _, env := range tenant.Environments {
			if env == nil {
				continue
			}

			for _, feature := range env.Features {
				if feature == nil || feature.Name != name {
					continue
				}

				statuses = append(statuses, FeatureStatus{
					Tenant:      tenant.Name,
					Environment: env.Name,
					Kind:        string(env.Kind),
					Enabled:     feature.State != nil && feature.State.Enabled,
				})
			}
		}
	}

	return statuses, nil
}

func (s *store) fetchFeatureConfig(ctx context.Context, featureName, envID string) (*Configuration, error) {
	var envIDPtr *string
	if envID != "" {
		envIDPtr = &envID
	}

	resp, err := fasitgraphql.GetFeatureConfig(ctx, s.client, featureName, envIDPtr)
	if err != nil {
		return nil, fmt.Errorf("get feature config: %w", err)
	}

	if resp.Configuration == nil {
		return &Configuration{}, nil
	}

	items := make([]ConfigurationItem, 0, len(resp.Configuration.Configuration))
	for _, cfg := range resp.Configuration.Configuration {
		if cfg == nil || cfg.Value == nil {
			continue
		}

		var configMeta *ConfigMeta
		if cfg.Value.Config != nil {
			configMeta = &ConfigMeta{
				Type:   string(cfg.Value.Config.Type),
				Secret: cfg.Value.Config.Secret,
			}
		}

		var computedMeta *ComputedMeta
		if cfg.Value.Computed != nil {
			computedMeta = &ComputedMeta{Template: cfg.Value.Computed.Template}
		}

		var content any
		if cfg.Content != nil {
			copied := append(json.RawMessage(nil), (*cfg.Content)...)
			raw := json.RawMessage(copied)
			content = &raw
		}

		items = append(items, ConfigurationItem{
			ID: cfg.Id,
			Value: ConfigValue{
				Key:         cfg.Value.Key,
				DisplayName: cfg.Value.DisplayName,
				Description: cfg.Value.Description,
				Required:    cfg.Value.Required,
				Config:      configMeta,
				Computed:    computedMeta,
			},
			Content: content,
			Source:  string(cfg.Source),
		})
	}

	return &Configuration{Configuration: items}, nil
}

func convertFeature(gql *fasitgraphql.GetFeaturesFeaturesFeature) *Feature {
	if gql == nil {
		return &Feature{}
	}

	dependencies := make([]Dependency, 0, len(gql.Dependencies))
	for _, dep := range gql.Dependencies {
		if dep == nil {
			continue
		}

		dependencies = append(dependencies, Dependency{
			AnyOf: append([]string(nil), dep.AnyOf...),
			AllOf: append([]string(nil), dep.AllOf...),
		})
	}

	return &Feature{
		Name:             gql.Name,
		Chart:            gql.Chart,
		Version:          gql.Version,
		Source:           gql.Source,
		Description:      gql.Description,
		EnvironmentKinds: toStringSlice(gql.EnvironmentKinds),
		Dependencies:     dependencies,
	}
}

func flattenConfigurationDetails(configuration *Configuration) []ConfigurationDetail {
	if configuration == nil {
		return nil
	}

	items := make([]ConfigurationDetail, 0, len(configuration.Configuration))
	for _, item := range configuration.Configuration {
		items = append(items, ConfigurationDetail{
			ID:     item.ID,
			Key:    item.Value.Key,
			Value:  FormatDisplayValue(MaskConfigurationValue(item)),
			Source: item.Source,
		})
	}

	return items
}
