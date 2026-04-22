package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/naistrix"
	"github.com/nais/narcos/internal/fasit"
	"github.com/nais/narcos/internal/fasit/command/flag"
)

func completeFasitTenantArg(_ *flag.Fasit) func(context.Context, *naistrix.Arguments, string) ([]string, string) {
	return func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
		if args.Len() >= 1 {
			return nil, ""
		}

		return completeTenantNames(ctx, fasit.ListTenants)
	}
}

func completeFasitTenantEnvArgs(_ *flag.Fasit) func(context.Context, *naistrix.Arguments, string) ([]string, string) {
	return func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
		switch args.Len() {
		case 0:
			return completeTenantNames(ctx, fasit.ListTenants)
		case 1:
			return completeEnvironmentNames(ctx, args.Get("tenant"), fasit.ListTenants)
		default:
			return nil, ""
		}
	}
}

func completeFasitTenantEnvFeatureArgs(_ *flag.Fasit) func(context.Context, *naistrix.Arguments, string) ([]string, string) {
	return func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
		switch args.Len() {
		case 0:
			return completeTenantNames(ctx, fasit.ListTenants)
		case 1:
			return completeEnvironmentNames(ctx, args.Get("tenant"), fasit.ListTenants)
		case 2:
			return completeEnvironmentFeatureNames(ctx, args.Get("tenant"), args.Get("env"), fasit.GetEnvironment)
		default:
			return nil, ""
		}
	}
}

func completeFasitFeatureArg(_ *flag.Fasit) func(context.Context, *naistrix.Arguments, string) ([]string, string) {
	return func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
		if args.Len() >= 1 {
			return nil, ""
		}

		return completeFeatureNames(ctx, fasit.ListFeatures)
	}
}

func completeFasitRolloutArgs(_ *flag.Fasit) func(context.Context, *naistrix.Arguments, string) ([]string, string) {
	return func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
		switch args.Len() {
		case 0:
			return completeFeatureNames(ctx, fasit.ListFeatures)
		case 1:
			return completeRolloutVersions(ctx, args.Get("feature"), fasit.ListRollouts)
		default:
			return nil, ""
		}
	}
}

func completeTenantNames(ctx context.Context, listTenants fasitTenantLister) ([]string, string) {
	tenants, err := listTenants(ctx)
	if err != nil {
		return nil, autocompleteErrorHint("list Fasit tenants", err)
	}

	return tenantNames(tenants), "Choose the tenant."
}

func completeEnvironmentNames(ctx context.Context, tenant string, listTenants fasitTenantLister) ([]string, string) {
	if tenant == "" {
		return nil, "Choose the tenant first."
	}

	tenants, err := listTenants(ctx)
	if err != nil {
		return nil, autocompleteErrorHint("list Fasit environments", err)
	}

	environments := environmentNamesForTenant(tenants, tenant)
	if environments == nil {
		return nil, fmt.Sprintf("Unknown tenant %q.", tenant)
	}

	return environments, fmt.Sprintf("Choose an environment in %s.", tenant)
}

func completeEnvironmentFeatureNames(ctx context.Context, tenant, env string, getEnvironment fasitEnvironmentGetter) ([]string, string) {
	if tenant == "" {
		return nil, "Choose the tenant first."
	}
	if env == "" {
		return nil, "Choose the environment first."
	}

	environment, err := getEnvironment(ctx, tenant, env)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Sprintf("Unknown environment %q in tenant %q.", env, tenant)
		}
		return nil, autocompleteErrorHint("list Fasit features for this environment", err)
	}

	return featureNames(environment.Features), fmt.Sprintf("Choose a feature in %s/%s.", tenant, env)
}

func completeFeatureNames(ctx context.Context, listFeatures fasitFeatureLister) ([]string, string) {
	features, err := listFeatures(ctx)
	if err != nil {
		return nil, autocompleteErrorHint("list Fasit features", err)
	}

	return featureNames(features), "Choose the feature."
}

func completeRolloutVersions(ctx context.Context, feature string, listRollouts fasitRolloutLister) ([]string, string) {
	if feature == "" {
		return nil, "Choose the feature first."
	}

	rollouts, err := listRollouts(ctx, feature)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, fmt.Sprintf("Unknown feature %q.", feature)
		}
		return nil, autocompleteErrorHint("list rollout versions", err)
	}

	return rolloutVersions(rollouts), fmt.Sprintf("Choose a rollout version for %s.", feature)
}

func autocompleteErrorHint(action string, err error) string {
	return fmt.Sprintf("Unable to %s for autocomplete; check Fasit auth (%v).", action, err)
}

func tenantNames(tenants []fasit.Tenant) []string {
	names := make([]string, 0, len(tenants))
	for _, tenant := range tenants {
		names = append(names, tenant.Name)
	}

	return names
}

func environmentNamesForTenant(tenants []fasit.Tenant, tenantName string) []string {
	for _, tenant := range tenants {
		if tenant.Name != tenantName {
			continue
		}

		names := make([]string, 0, len(tenant.Environments))
		for _, env := range tenant.Environments {
			names = append(names, env.Name)
		}

		return names
	}

	return nil
}

func featureNames(features []fasit.Feature) []string {
	names := make([]string, 0, len(features))
	for _, feature := range features {
		names = append(names, feature.Name)
	}

	return names
}

func rolloutVersions(rollouts []fasit.Rollout) []string {
	versions := make([]string, 0, len(rollouts))
	for _, rollout := range rollouts {
		versions = append(versions, rollout.Version)
	}

	return versions
}
