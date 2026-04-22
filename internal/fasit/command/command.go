package command

import (
	"context"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/nais/naistrix"
	"github.com/nais/narcos/internal/fasit"
	"github.com/nais/narcos/internal/fasit/command/flag"
	"golang.org/x/term"
)

type fasitTenantLister func(context.Context) ([]fasit.Tenant, error)
type fasitEnvironmentGetter func(context.Context, string, string) (*fasit.Environment, error)
type fasitFeatureLister func(context.Context) ([]fasit.Feature, error)
type fasitRolloutLister func(context.Context, string) ([]fasit.Rollout, error)

func Fasit(globalFlags *naistrix.GlobalFlags) *naistrix.Command {
	fasitFlags := &flag.Fasit{GlobalFlags: globalFlags}
	return &naistrix.Command{
		Name:        "fasit",
		Title:       "Manage Fasit configuration.",
		StickyFlags: fasitFlags,
		SubCommands: []*naistrix.Command{
			loginCmd(fasitFlags),
			tenantsCmd(fasitFlags),
			tenantCmd(fasitFlags),
			envCmd(fasitFlags),
			featuresCmd(fasitFlags),
			featureCmd(fasitFlags),
			rolloutsCmd(fasitFlags),
			rolloutCmd(fasitFlags),
			deploymentCmd(fasitFlags),
		},
	}
}

func tenantsCmd(parentFlags *flag.Fasit) *naistrix.Command {
	return &naistrix.Command{
		Name:  "tenants",
		Title: "List and inspect tenants.",
		SubCommands: []*naistrix.Command{
			tenantsListCmd(parentFlags),
		},
	}
}

func tenantCmd(parentFlags *flag.Fasit) *naistrix.Command {
	return &naistrix.Command{
		Name:  "tenant",
		Title: "Inspect a single tenant.",
		SubCommands: []*naistrix.Command{
			tenantGetCmd(parentFlags),
		},
	}
}

func envCmd(parentFlags *flag.Fasit) *naistrix.Command {
	return &naistrix.Command{
		Name:  "env",
		Title: "Inspect environments and features.",
		SubCommands: []*naistrix.Command{
			envGetCmd(parentFlags),
			envFeatureCmd(parentFlags),
		},
	}
}

func featuresCmd(parentFlags *flag.Fasit) *naistrix.Command {
	return &naistrix.Command{
		Name:  "features",
		Title: "List features.",
		SubCommands: []*naistrix.Command{
			featuresListCmd(parentFlags),
		},
	}
}

func featureCmd(parentFlags *flag.Fasit) *naistrix.Command {
	return &naistrix.Command{
		Name:  "feature",
		Title: "Inspect a single feature.",
		SubCommands: []*naistrix.Command{
			featureGetCmd(parentFlags),
			featureStatusCmd(parentFlags),
			featureRolloutsCmd(parentFlags),
		},
	}
}

func envFeatureCmd(parentFlags *flag.Fasit) *naistrix.Command {
	return &naistrix.Command{
		Name:  "feature",
		Title: "Inspect a feature in a specific environment.",
		SubCommands: []*naistrix.Command{
			envFeatureGetCmd(parentFlags),
			envFeatureLogsCmd(parentFlags),
			envFeatureHelmCmd(parentFlags),
			envFeatureRolloutsCmd(parentFlags),
			envFeatureAuditCmd(parentFlags),
			envFeatureEnableCmd(parentFlags),
			envFeatureDisableCmd(parentFlags),
			envFeatureConfigCmd(parentFlags),
		},
	}
}

func rolloutsCmd(parentFlags *flag.Fasit) *naistrix.Command {
	return &naistrix.Command{
		Name:  "rollouts",
		Title: "Inspect rollout history.",
		SubCommands: []*naistrix.Command{
			rolloutsListCmd(parentFlags),
		},
	}
}

func rolloutCmd(parentFlags *flag.Fasit) *naistrix.Command {
	return &naistrix.Command{
		Name:  "rollout",
		Title: "Inspect a single rollout.",
		SubCommands: []*naistrix.Command{
			rolloutGetCmd(parentFlags),
		},
	}
}

func deploymentCmd(parentFlags *flag.Fasit) *naistrix.Command {
	return &naistrix.Command{
		Name:  "deployment",
		Title: "Inspect a single deployment-backed rollout.",
		SubCommands: []*naistrix.Command{
			deploymentGetCmd(parentFlags),
		},
	}
}

func tenantsListCmd(parentFlags *flag.Fasit) *naistrix.Command {
	flags := &flag.TenantsList{Fasit: parentFlags}
	return &naistrix.Command{
		Name:  "list",
		Title: "List all tenants.",
		Flags: flags,
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			tenants, err := fasit.ListTenants(ctx)
			if err != nil {
				return err
			}

			tenants = filterTenants(tenants, flags.Tenant, flags.Kind)

			type tenantRow struct {
				Name         string `heading:"Name"`
				Environments string `heading:"Environments"`
			}

			rows := make([]tenantRow, 0, len(tenants))
			for _, tenant := range tenants {
				envs := make([]string, 0, len(tenant.Environments))
				for _, env := range tenant.Environments {
					envs = append(envs, env.Name)
				}

				rows = append(rows, tenantRow{Name: tenant.Name, Environments: strings.Join(envs, ", ")})
			}

			return fasit.RenderStructuredOutput(out, flags.Output, rows, tenants)
		},
	}
}

func tenantGetCmd(parentFlags *flag.Fasit) *naistrix.Command {
	flags := &flag.TenantGet{Fasit: parentFlags}
	return &naistrix.Command{
		Name:             "get",
		Title:            "Get tenant details.",
		Flags:            flags,
		AutoCompleteFunc: completeFasitTenantArg(flags.Fasit),
		Args:             []naistrix.Argument{{Name: "tenant"}},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			tenant, err := fasit.GetTenant(ctx, args.Get("tenant"))
			if err != nil {
				return err
			}

			type tenantRow struct {
				Tenant      string `heading:"Tenant"`
				Environment string `heading:"Environment"`
				Kind        string `heading:"Kind"`
				Reconcile   bool   `heading:"Reconcile"`
				Features    int    `heading:"Features"`
			}

			rows := make([]tenantRow, 0, len(tenant.Environments))
			for _, env := range tenant.Environments {
				rows = append(rows, tenantRow{
					Tenant:      tenant.Name,
					Environment: env.Name,
					Kind:        env.Kind,
					Reconcile:   env.Reconcile,
					Features:    len(env.Features),
				})
			}

			return fasit.RenderStructuredOutput(out, flags.Output, rows, tenant)
		},
	}
}

func envGetCmd(parentFlags *flag.Fasit) *naistrix.Command {
	flags := &flag.EnvGet{Fasit: parentFlags}
	return &naistrix.Command{
		Name:             "get",
		Title:            "Get environment details.",
		Flags:            flags,
		AutoCompleteFunc: completeFasitTenantEnvArgs(flags.Fasit),
		Args:             []naistrix.Argument{{Name: "tenant"}, {Name: "env"}},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			env, err := fasit.GetEnvironment(ctx, args.Get("tenant"), args.Get("env"))
			if err != nil {
				return err
			}

			type envRow struct {
				Environment string `heading:"Environment"`
				Kind        string `heading:"Kind"`
				Reconcile   bool   `heading:"Reconcile"`
				Features    string `heading:"Features"`
			}

			features := make([]string, 0, len(env.Features))
			for _, feature := range env.Features {
				state := "disabled"
				if feature.Enabled {
					state = "enabled"
				}
				features = append(features, feature.Name+" ("+state+")")
			}

			rows := []envRow{{
				Environment: env.Name,
				Kind:        env.Kind,
				Reconcile:   env.Reconcile,
				Features:    strings.Join(features, ", "),
			}}

			return fasit.RenderStructuredOutput(out, flags.Output, rows, env)
		},
	}
}

func featuresListCmd(parentFlags *flag.Fasit) *naistrix.Command {
	flags := &flag.FeaturesList{Fasit: parentFlags}
	return &naistrix.Command{
		Name:  "list",
		Title: "List features.",
		Flags: flags,
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			features, err := fasit.ListFeatures(ctx)
			if err != nil {
				return err
			}

			features = filterFeatures(features, flags.Feature, flags.Kind)

			type featureRow struct {
				Name    string `heading:"Name"`
				Chart   string `heading:"Chart"`
				Version string `heading:"Version"`
				Kinds   string `heading:"Environment kinds"`
			}

			rows := make([]featureRow, 0, len(features))
			for _, feature := range features {
				rows = append(rows, featureRow{
					Name:    feature.Name,
					Chart:   feature.Chart,
					Version: feature.Version,
					Kinds:   strings.Join(feature.EnvironmentKinds, ", "),
				})
			}

			return fasit.RenderStructuredOutput(out, flags.Output, rows, features)
		},
	}
}

func featureGetCmd(parentFlags *flag.Fasit) *naistrix.Command {
	flags := &flag.FeatureGet{Fasit: parentFlags}
	return &naistrix.Command{
		Name:             "get",
		Title:            "Get feature details.",
		Flags:            flags,
		AutoCompleteFunc: completeFasitFeatureArg(flags.Fasit),
		Args:             []naistrix.Argument{{Name: "feature"}},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			feature, err := fasit.GetFeature(ctx, args.Get("feature"))
			if err != nil {
				return err
			}

			type featureRow struct {
				Name         string `heading:"Name"`
				Chart        string `heading:"Chart"`
				Version      string `heading:"Version"`
				Source       string `heading:"Source"`
				Kinds        string `heading:"Environment kinds"`
				Configs      int    `heading:"Configs"`
				Dependencies string `heading:"Dependencies"`
			}

			deps := make([]string, 0, len(feature.Dependencies))
			for _, dep := range feature.Dependencies {
				if len(dep.AnyOf) > 0 {
					deps = append(deps, "anyOf="+strings.Join(dep.AnyOf, ","))
				}
				if len(dep.AllOf) > 0 {
					deps = append(deps, "allOf="+strings.Join(dep.AllOf, ","))
				}
			}

			rows := []featureRow{{
				Name:         feature.Name,
				Chart:        feature.Chart,
				Version:      feature.Version,
				Source:       feature.Source,
				Kinds:        strings.Join(feature.EnvironmentKinds, ", "),
				Configs:      len(feature.Configurations),
				Dependencies: strings.Join(deps, "; "),
			}}

			return fasit.RenderStructuredOutput(out, flags.Output, rows, fasit.MaskedFeatureOutput(feature))
		},
	}
}

func featureStatusCmd(parentFlags *flag.Fasit) *naistrix.Command {
	flags := &flag.FeatureStatus{Fasit: parentFlags}
	return &naistrix.Command{
		Name:             "status",
		Title:            "Get feature status across environments.",
		Flags:            flags,
		AutoCompleteFunc: completeFasitFeatureArg(flags.Fasit),
		Args:             []naistrix.Argument{{Name: "feature"}},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			statuses, err := fasit.GetFeatureStatus(ctx, args.Get("feature"))
			if err != nil {
				return err
			}

			statuses, err = filterFeatureStatuses(statuses, flags.Tenant, flags.Env, flags.Kind, flags.Enabled)
			if err != nil {
				return err
			}

			return fasit.RenderStructuredOutput(out, flags.Output, statuses, statuses)
		},
	}
}

func featureRolloutsCmd(parentFlags *flag.Fasit) *naistrix.Command {
	flags := &flag.FeatureRollouts{Fasit: parentFlags}
	return &naistrix.Command{
		Name:             "rollouts",
		Title:            "Get rollout history for a feature.",
		Flags:            flags,
		AutoCompleteFunc: completeFasitFeatureArg(flags.Fasit),
		Args:             []naistrix.Argument{{Name: "feature"}},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			rollouts, err := fasit.ListFeatureRollouts(ctx, args.Get("feature"))
			if err != nil {
				return err
			}

			return fasit.RenderStructuredOutput(out, flags.Output, rolloutRows(rollouts), rollouts)
		},
	}
}

func envFeatureGetCmd(parentFlags *flag.Fasit) *naistrix.Command {
	flags := &flag.EnvFeatureGet{Fasit: parentFlags}
	return &naistrix.Command{
		Name:             "get",
		Title:            "Get feature configuration for an environment.",
		Flags:            flags,
		AutoCompleteFunc: completeFasitTenantEnvFeatureArgs(flags.Fasit),
		Args:             []naistrix.Argument{{Name: "tenant"}, {Name: "env"}, {Name: "feature"}},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			_, feature, err := fasit.GetEnvFeature(ctx, args.Get("tenant"), args.Get("env"), args.Get("feature"))
			if err != nil {
				return err
			}

			rows := fasit.MaskedConfigurationItems(feature.Configuration)
			return fasit.RenderStructuredOutput(out, flags.Output, rows, rows)
		},
	}
}

func envFeatureLogsCmd(parentFlags *flag.Fasit) *naistrix.Command {
	flags := &flag.EnvFeatureLogs{Fasit: parentFlags}
	return &naistrix.Command{
		Name:             "logs",
		Title:            "Get feature rollout log and helm diff.",
		Flags:            flags,
		AutoCompleteFunc: completeFasitTenantEnvFeatureArgs(flags.Fasit),
		Args:             []naistrix.Argument{{Name: "tenant"}, {Name: "env"}, {Name: "feature"}},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			featureLog, err := fasit.GetFeatureLog(ctx, args.Get("tenant"), args.Get("env"), args.Get("feature"))
			if err != nil {
				return err
			}

			return renderFeatureLog(out, flags.Output, featureLog)
		},
	}
}

func envFeatureHelmCmd(parentFlags *flag.Fasit) *naistrix.Command {
	flags := &flag.EnvFeatureHelm{Fasit: parentFlags}
	return &naistrix.Command{
		Name:             "helm",
		Title:            "Get computed helm values.",
		Flags:            flags,
		AutoCompleteFunc: completeFasitTenantEnvFeatureArgs(flags.Fasit),
		Args:             []naistrix.Argument{{Name: "tenant"}, {Name: "env"}, {Name: "feature"}},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			values, err := fasit.GetHelmValues(ctx, args.Get("feature"), args.Get("tenant"), args.Get("env"))
			if err != nil {
				return err
			}

			return renderHelmValues(out, flags.Output, values)
		},
	}
}

func envFeatureRolloutsCmd(parentFlags *flag.Fasit) *naistrix.Command {
	flags := &flag.EnvFeatureRollouts{Fasit: parentFlags}
	return &naistrix.Command{
		Name:             "rollouts",
		Title:            "Get feature rollout history relevant to an environment when possible.",
		Flags:            flags,
		AutoCompleteFunc: completeFasitTenantEnvFeatureArgs(flags.Fasit),
		Args:             []naistrix.Argument{{Name: "tenant"}, {Name: "env"}, {Name: "feature"}},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			if _, _, err := fasit.GetEnvFeature(ctx, args.Get("tenant"), args.Get("env"), args.Get("feature")); err != nil {
				return err
			}

			rollouts, err := fasit.ListFeatureRollouts(ctx, args.Get("feature"))
			if err != nil {
				return err
			}

			relevant := filterEnvironmentFeatureRollouts(rollouts, args.Get("tenant"), args.Get("env"))
			payload := struct {
				Tenant      string          `json:"tenant" yaml:"tenant"`
				Environment string          `json:"environment" yaml:"environment"`
				Feature     string          `json:"feature" yaml:"feature"`
				Scope       string          `json:"scope" yaml:"scope"`
				Warning     string          `json:"warning,omitempty" yaml:"warning,omitempty"`
				Rollouts    []fasit.Rollout `json:"rollouts" yaml:"rollouts"`
			}{
				Tenant:      args.Get("tenant"),
				Environment: args.Get("env"),
				Feature:     args.Get("feature"),
				Scope:       "environment-relevant",
				Rollouts:    relevant,
			}

			if len(relevant) == 0 {
				payload.Scope = "feature-wide"
				payload.Warning = "No environment-scoped deployment history was found for this tenant/environment. Showing feature-wide rollout history instead."
				relevant = rollouts
				payload.Rollouts = rollouts
			}

			switch fasit.NormalizeOutputFormat(flags.Output) {
			case fasit.OutputFormatJSON, fasit.OutputFormatYAML:
				return fasit.RenderDataOutput(out, flags.Output, payload)
			default:
				if payload.Warning != "" {
					out.Println(payload.Warning)
					out.Println("")
				}
				return fasit.RenderStructuredOutput(out, flags.Output, rolloutRows(relevant), relevant)
			}
		},
	}
}

func envFeatureAuditCmd(parentFlags *flag.Fasit) *naistrix.Command {
	flags := &flag.EnvFeatureAudit{Fasit: parentFlags}
	return &naistrix.Command{
		Name:             "audit",
		Title:            "Show audit history for an environment feature.",
		Flags:            flags,
		AutoCompleteFunc: completeFasitTenantEnvFeatureArgs(flags.Fasit),
		Args:             []naistrix.Argument{{Name: "tenant"}, {Name: "env"}, {Name: "feature"}},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			return renderEnvFeatureAuditPlaceholder(out, flags.Output, args.Get("tenant"), args.Get("env"), args.Get("feature"))
		},
	}
}

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

func renderEnvFeatureAuditPlaceholder(out *naistrix.OutputWriter, format, tenant, environment, feature string) error {
	placeholder := struct {
		Available   bool   `json:"available" yaml:"available"`
		Message     string `json:"message" yaml:"message"`
		Tenant      string `json:"tenant" yaml:"tenant"`
		Environment string `json:"environment" yaml:"environment"`
		Feature     string `json:"feature" yaml:"feature"`
	}{
		Available:   false,
		Message:     "Audit data is not available: the Fasit backend does not currently expose audit history.",
		Tenant:      tenant,
		Environment: environment,
		Feature:     feature,
	}

	switch fasit.NormalizeOutputFormat(format) {
	case fasit.OutputFormatJSON, fasit.OutputFormatYAML:
		return fasit.RenderDataOutput(out, format, placeholder)
	default:
		out.Println(placeholder.Message)
		return nil
	}
}

func envFeatureEnableCmd(parentFlags *flag.Fasit) *naistrix.Command {
	flags := &flag.EnvFeatureEnable{Fasit: parentFlags}
	return &naistrix.Command{
		Name:  "enable",
		Title: "Enable feature reconcile.",
		Flags: flags,
		Args:  []naistrix.Argument{{Name: "tenant"}, {Name: "env"}, {Name: "feature"}},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			return setFeatureState(ctx, args, out, flags.Fasit, flags.Yes, true)
		},
	}
}

func envFeatureDisableCmd(parentFlags *flag.Fasit) *naistrix.Command {
	flags := &flag.EnvFeatureDisable{Fasit: parentFlags}
	return &naistrix.Command{
		Name:  "disable",
		Title: "Disable feature reconcile.",
		Flags: flags,
		Args:  []naistrix.Argument{{Name: "tenant"}, {Name: "env"}, {Name: "feature"}},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			return setFeatureState(ctx, args, out, flags.Fasit, flags.Yes, false)
		},
	}
}

func envFeatureConfigCmd(parentFlags *flag.Fasit) *naistrix.Command {
	return &naistrix.Command{
		Name:  "config",
		Title: "Manage environment feature configuration.",
		SubCommands: []*naistrix.Command{
			envFeatureConfigSetCmd(parentFlags),
			envFeatureConfigOverrideCmd(parentFlags),
		},
	}
}

func envFeatureConfigSetCmd(parentFlags *flag.Fasit) *naistrix.Command {
	flags := &flag.EnvFeatureConfigSet{Fasit: parentFlags}
	return &naistrix.Command{
		Name:  "set",
		Title: "Update an existing configuration.",
		Flags: flags,
		Args:  []naistrix.Argument{{Name: "tenant"}, {Name: "env"}, {Name: "feature"}, {Name: "config-id"}},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			tenant := args.Get("tenant")
			envName := args.Get("env")
			featureName := args.Get("feature")
			configID := args.Get("config-id")

			_, feature, err := fasit.GetEnvFeature(ctx, tenant, envName, featureName)
			if err != nil {
				return err
			}

			config, err := findConfigurationByID(feature.Configuration, configID)
			if err != nil {
				return err
			}

			rawValue, err := resolveConfigMutationInput(out, flags.Value, configIsSecret(config))
			if err != nil {
				return err
			}

			parsedValue, err := fasit.ParseConfigValue(configType(config), rawValue)
			if err != nil {
				return err
			}

			if err := fasit.ConfirmMutation(out, os.Stdin, flags.Yes,
				fmt.Sprintf("About to update configuration %q (%s) for feature %q in %s/%s.", configID, config.Value.Key, featureName, tenant, envName),
				fmt.Sprintf("New value: %s", fasit.DisplayMutationValue(parsedValue, configIsSecret(config))),
			); err != nil {
				return err
			}

			if err := fasit.UpdateConfiguration(ctx, configID, parsedValue); err != nil {
				return err
			}

			out.Println("Configuration updated.")
			return nil
		},
	}
}

func envFeatureConfigOverrideCmd(parentFlags *flag.Fasit) *naistrix.Command {
	flags := &flag.EnvFeatureConfigOverride{Fasit: parentFlags}
	return &naistrix.Command{
		Name:  "override",
		Title: "Create a configuration override.",
		Flags: flags,
		Args:  []naistrix.Argument{{Name: "tenant"}, {Name: "env"}, {Name: "feature"}, {Name: "key"}},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			tenant := args.Get("tenant")
			envName := args.Get("env")
			featureName := args.Get("feature")
			key := args.Get("key")

			env, feature, err := fasit.GetEnvFeature(ctx, tenant, envName, featureName)
			if err != nil {
				return err
			}

			config, err := findConfigurationByKey(feature.Configuration, key)
			if err != nil {
				return err
			}

			rawValue, err := resolveConfigMutationInput(out, flags.Value, configIsSecret(config))
			if err != nil {
				return err
			}

			parsedValue, err := fasit.ParseConfigValue(configType(config), rawValue)
			if err != nil {
				return err
			}

			if err := fasit.ConfirmMutation(out, os.Stdin, flags.Yes,
				fmt.Sprintf("About to create configuration override %q for feature %q in %s/%s.", key, featureName, tenant, envName),
				fmt.Sprintf("New value: %s", fasit.DisplayMutationValue(parsedValue, configIsSecret(config))),
			); err != nil {
				return err
			}

			if err := fasit.CreateConfiguration(ctx, env.ID, featureName, key, parsedValue); err != nil {
				return err
			}

			out.Println("Configuration override created.")
			return nil
		},
	}
}

func rolloutsListCmd(parentFlags *flag.Fasit) *naistrix.Command {
	flags := &flag.RolloutsList{Fasit: parentFlags}
	return &naistrix.Command{
		Name:  "list",
		Title: "List all rollouts.",
		Flags: flags,
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			rollouts, err := fasit.ListAllRollouts(ctx)
			if err != nil {
				return err
			}

			rollouts = filterRolloutSummaries(rollouts, flags.Feature, flags.Status)

			return fasit.RenderStructuredOutput(out, flags.Output, rolloutSummaryRows(rollouts), rollouts)
		},
	}
}

func filterTenants(tenants []fasit.Tenant, tenantName, kind string) []fasit.Tenant {
	filtered := make([]fasit.Tenant, 0, len(tenants))
	for _, tenant := range tenants {
		if tenantName != "" && tenant.Name != tenantName {
			continue
		}

		if kind == "" {
			filtered = append(filtered, tenant)
			continue
		}

		envs := make([]fasit.Environment, 0, len(tenant.Environments))
		for _, env := range tenant.Environments {
			if env.Kind == kind {
				envs = append(envs, env)
			}
		}

		if len(envs) == 0 {
			continue
		}

		tenant.Environments = envs
		filtered = append(filtered, tenant)
	}

	return filtered
}

func filterFeatures(features []fasit.Feature, featureName, kind string) []fasit.Feature {
	filtered := make([]fasit.Feature, 0, len(features))
	for _, feature := range features {
		if featureName != "" && feature.Name != featureName {
			continue
		}

		if kind != "" && !slices.Contains(feature.EnvironmentKinds, kind) {
			continue
		}

		filtered = append(filtered, feature)
	}

	return filtered
}

func filterFeatureStatuses(statuses []fasit.FeatureStatus, tenantName, envName, kind, enabled string) ([]fasit.FeatureStatus, error) {
	var enabledFilter *bool
	if enabled != "" {
		parsed, err := strconv.ParseBool(enabled)
		if err != nil {
			return nil, fmt.Errorf("invalid --enabled value %q: must be true or false", enabled)
		}
		enabledFilter = &parsed
	}

	filtered := make([]fasit.FeatureStatus, 0, len(statuses))
	for _, status := range statuses {
		if tenantName != "" && status.Tenant != tenantName {
			continue
		}
		if envName != "" && status.Environment != envName {
			continue
		}
		if kind != "" && status.Kind != kind {
			continue
		}
		if enabledFilter != nil && status.Enabled != *enabledFilter {
			continue
		}

		filtered = append(filtered, status)
	}

	return filtered, nil
}

func filterRolloutSummaries(rollouts []fasit.RolloutSummary, featureName, status string) []fasit.RolloutSummary {
	filtered := make([]fasit.RolloutSummary, 0, len(rollouts))
	for _, rollout := range rollouts {
		if featureName != "" && rollout.FeatureName != featureName {
			continue
		}
		if status != "" && !strings.EqualFold(rollout.Status, status) {
			continue
		}

		filtered = append(filtered, rollout)
	}

	return filtered
}

func rolloutGetCmd(parentFlags *flag.Fasit) *naistrix.Command {
	flags := &flag.RolloutGet{Fasit: parentFlags}
	return &naistrix.Command{
		Name:             "get",
		Title:            "Get rollout detail.",
		Flags:            flags,
		AutoCompleteFunc: completeFasitRolloutArgs(flags.Fasit),
		Args:             []naistrix.Argument{{Name: "feature"}, {Name: "version"}},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			detail, err := fasit.GetRollout(ctx, args.Get("feature"), args.Get("version"))
			if err != nil {
				return err
			}

			return renderRolloutDetail(out, flags.Output, detail)
		},
	}
}

func deploymentGetCmd(parentFlags *flag.Fasit) *naistrix.Command {
	flags := &flag.DeploymentGet{Fasit: parentFlags}
	return &naistrix.Command{
		Name:  "get",
		Title: "Get deployment-backed rollout detail.",
		Flags: flags,
		Args:  []naistrix.Argument{{Name: "id"}},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			detail, err := fasit.GetDeployment(ctx, args.Get("id"))
			if err != nil {
				return err
			}

			return renderDeploymentDetail(out, flags.Output, detail)
		},
	}
}

type rolloutRow struct {
	Feature   string `heading:"Feature" json:"feature" yaml:"feature"`
	Version   string `heading:"Version" json:"version" yaml:"version"`
	Status    string `heading:"Status" json:"status" yaml:"status"`
	Target    string `heading:"Target" json:"target,omitempty" yaml:"target,omitempty"`
	DetailRef string `heading:"Detail" json:"detailRef,omitempty" yaml:"detailRef,omitempty"`
	Created   string `heading:"Created" json:"created" yaml:"created"`
	Completed string `heading:"Completed" json:"completed" yaml:"completed"`
}

type rolloutEventRow struct {
	Created string `heading:"Created" json:"created" yaml:"created"`
	Failure bool   `heading:"Failure" json:"failure" yaml:"failure"`
	Message string `heading:"Message" json:"message" yaml:"message"`
}

type logLineRow struct {
	Timestamp string `heading:"Timestamp" json:"timestamp" yaml:"timestamp"`
	Message   string `heading:"Message" json:"message" yaml:"message"`
}

type deploymentRow struct {
	ID          string `heading:"Deployment ID" json:"id" yaml:"id"`
	FeatureName string `heading:"Feature" json:"featureName" yaml:"featureName"`
	Version     string `heading:"Version" json:"version" yaml:"version"`
	Target      string `heading:"Target" json:"target" yaml:"target"`
	Created     string `heading:"Created" json:"created" yaml:"created"`
	Description string `heading:"Description" json:"description,omitempty" yaml:"description,omitempty"`
}

type deploymentStatusRow struct {
	Tenant       string `heading:"Tenant" json:"tenant" yaml:"tenant"`
	Environment  string `heading:"Environment" json:"environment" yaml:"environment"`
	State        string `heading:"State" json:"state" yaml:"state"`
	Message      string `heading:"Message" json:"message" yaml:"message"`
	LastModified string `heading:"Last modified" json:"lastModified" yaml:"lastModified"`
}

func rolloutRows(rollouts []fasit.Rollout) []rolloutRow {
	rows := make([]rolloutRow, 0, len(rollouts))
	for _, rollout := range rollouts {
		rows = append(rows, rolloutRow{
			Feature:   rollout.FeatureName,
			Version:   rollout.Version,
			Status:    rollout.Status,
			Target:    rollout.Target,
			DetailRef: rolloutDetailRef(rollout.DeploymentID),
			Created:   rollout.Created,
			Completed: rollout.Completed,
		})
	}

	return rows
}

func rolloutSummaryRows(rollouts []fasit.RolloutSummary) []rolloutRow {
	rows := make([]rolloutRow, 0, len(rollouts))
	for _, rollout := range rollouts {
		rows = append(rows, rolloutRow{
			Feature:   rollout.FeatureName,
			Version:   rollout.Version,
			Status:    rollout.Status,
			Target:    rollout.Target,
			DetailRef: rolloutDetailRef(rollout.DeploymentID),
			Created:   rollout.Created,
			Completed: rollout.Completed,
		})
	}

	return rows
}

func renderFeatureLog(out *naistrix.OutputWriter, format string, featureLog *fasit.FeatureLog) error {
	switch fasit.NormalizeOutputFormat(format) {
	case fasit.OutputFormatJSON, fasit.OutputFormatYAML:
		return fasit.RenderDataOutput(out, format, featureLog)
	default:
		summary := []struct {
			Version      string `heading:"Version"`
			Status       string `heading:"Status"`
			LastModified string `heading:"Last modified"`
		}{{
			Version:      featureLog.CurrentVersion,
			Status:       featureLog.CurrentStatus,
			LastModified: featureLog.LastModified,
		}}

		if err := out.Table().Render(summary); err != nil {
			return err
		}

		if len(featureLog.CurrentLog) > 0 {
			out.Println("")
			if err := out.Table().Render(logLineRows(featureLog.CurrentLog)); err != nil {
				return err
			}
		}

		if featureLog.HelmDiff.Diff != "" {
			out.Println("")
			out.Println("Helm diff:")
			out.Println(featureLog.HelmDiff.Diff)
		}

		return nil
	}
}

func renderHelmValues(out *naistrix.OutputWriter, format, values string) error {
	switch fasit.NormalizeOutputFormat(format) {
	case fasit.OutputFormatJSON, fasit.OutputFormatYAML:
		return fasit.RenderDataOutput(out, format, map[string]string{"helmValues": values})
	default:
		out.Println(values)
		return nil
	}
}

func renderRolloutDetail(out *naistrix.OutputWriter, format string, detail *fasit.RolloutDetail) error {
	switch fasit.NormalizeOutputFormat(format) {
	case fasit.OutputFormatJSON, fasit.OutputFormatYAML:
		return fasit.RenderDataOutput(out, format, detail)
	default:
		if err := out.Table().Render([]rolloutRow{{
			Feature:   detail.FeatureName,
			Version:   detail.Version,
			Status:    detail.Status,
			Target:    "",
			DetailRef: "",
			Created:   detail.Created,
			Completed: detail.Completed,
		}}); err != nil {
			return err
		}

		out.Println("")
		out.Println("Events:")
		if err := out.Table().Render(rolloutEventRows(detail.Events)); err != nil {
			return err
		}

		for _, log := range detail.Logs {
			out.Println("")
			out.Println(fmt.Sprintf("Logs for %s/%s:", log.TenantName, log.Environment))
			if err := out.Table().Render(logLineRows(log.Lines)); err != nil {
				return err
			}
		}

		return nil
	}
}

func renderDeploymentDetail(out *naistrix.OutputWriter, format string, detail *fasit.Deployment) error {
	switch fasit.NormalizeOutputFormat(format) {
	case fasit.OutputFormatJSON, fasit.OutputFormatYAML:
		return fasit.RenderDataOutput(out, format, detail)
	default:
		if err := out.Table().Render([]deploymentRow{{
			ID:          detail.ID,
			FeatureName: detail.FeatureName,
			Version:     detail.Version,
			Target:      detail.Target,
			Created:     detail.Created,
			Description: detail.Description,
		}}); err != nil {
			return err
		}

		out.Println("")
		out.Println("Environment statuses:")
		if len(detail.Statuses) == 0 {
			out.Println("No results")
			return nil
		}

		return out.Table().Render(deploymentStatusRows(detail.Statuses))
	}
}

func rolloutEventRows(events []fasit.RolloutEvent) []rolloutEventRow {
	rows := make([]rolloutEventRow, 0, len(events))
	for _, event := range events {
		rows = append(rows, rolloutEventRow{Created: event.Created, Failure: event.Failure, Message: event.Message})
	}

	return rows
}

func logLineRows(lines []fasit.LogLine) []logLineRow {
	rows := make([]logLineRow, 0, len(lines))
	for _, line := range lines {
		rows = append(rows, logLineRow{Timestamp: line.Timestamp, Message: line.Message})
	}

	return rows
}

func deploymentStatusRows(statuses []fasit.DeploymentStatus) []deploymentStatusRow {
	rows := make([]deploymentStatusRow, 0, len(statuses))
	for _, status := range statuses {
		rows = append(rows, deploymentStatusRow{
			Tenant:       status.TenantName,
			Environment:  status.EnvironmentName,
			State:        status.State,
			Message:      status.Message,
			LastModified: status.LastModified,
		})
	}

	return rows
}

func rolloutDetailRef(deploymentID string) string {
	if deploymentID == "" {
		return "rollout"
	}

	return "deployment:" + deploymentID
}

func filterEnvironmentFeatureRollouts(rollouts []fasit.Rollout, tenant, env string) []fasit.Rollout {
	if tenant == "" || env == "" {
		return nil
	}

	filtered := make([]fasit.Rollout, 0, len(rollouts))
	for _, rollout := range rollouts {
		if rollout.DeploymentID == "" {
			continue
		}

		labels := map[string]string{}
		for target := range strings.SplitSeq(rollout.Target, ",") {
			parts := strings.SplitN(strings.TrimSpace(target), "=", 2)
			if len(parts) != 2 {
				continue
			}
			labels[parts[0]] = parts[1]
		}

		tenantLabel, tenantOK := labels["tenant"]
		envLabel, envOK := labels["environment"]
		if !envOK {
			envLabel, envOK = labels["env"]
		}
		if tenantOK && envOK && tenantLabel == tenant && envLabel == env {
			filtered = append(filtered, rollout)
			continue
		}
		if combined, ok := labels[tenant]; ok && combined == env {
			filtered = append(filtered, rollout)
			continue
		}
		if len(labels) == 0 && strings.TrimSpace(rollout.Target) == tenant+"="+env {
			filtered = append(filtered, rollout)
			continue
		}
	}

	return filtered
}

func setFeatureState(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter, _ *flag.Fasit, yes, enabled bool) error {
	tenant := args.Get("tenant")
	envName := args.Get("env")
	featureName := args.Get("feature")

	env, _, err := fasit.GetEnvFeature(ctx, tenant, envName, featureName)
	if err != nil {
		return err
	}

	action := "disable"
	newState := "disabled"
	if enabled {
		action = "enable"
		newState = "enabled"
	}

	if err := fasit.ConfirmMutation(out, os.Stdin, yes,
		fmt.Sprintf("About to %s reconcile for feature %q in %s/%s.", action, featureName, tenant, envName),
		fmt.Sprintf("New state: %s", newState),
	); err != nil {
		return err
	}

	if err := fasit.SetFeatureState(ctx, env.ID, featureName, enabled); err != nil {
		return err
	}

	out.Println("Feature state updated.")
	return nil
}

func findConfigurationByID(configuration *fasit.Configuration, id string) (*fasit.ConfigurationItem, error) {
	if configuration == nil {
		return nil, fmt.Errorf("configuration %q not found", id)
	}

	for i := range configuration.Configuration {
		if configuration.Configuration[i].ID == id {
			return &configuration.Configuration[i], nil
		}
	}

	return nil, fmt.Errorf("configuration %q not found", id)
}

func findConfigurationByKey(configuration *fasit.Configuration, key string) (*fasit.ConfigurationItem, error) {
	if configuration == nil {
		return nil, fmt.Errorf("configuration key %q not found", key)
	}

	for i := range configuration.Configuration {
		if configuration.Configuration[i].Value.Key == key {
			return &configuration.Configuration[i], nil
		}
	}

	return nil, fmt.Errorf("configuration key %q not found", key)
}

func configType(item *fasit.ConfigurationItem) string {
	if item == nil || item.Value.Config == nil {
		return ""
	}

	return item.Value.Config.Type
}

func configIsSecret(item *fasit.ConfigurationItem) bool {
	return item != nil && item.Value.Config != nil && item.Value.Config.Secret
}

func resolveConfigMutationInput(out *naistrix.OutputWriter, flagValue string, secret bool) (string, error) {
	stdinFD := int(os.Stdin.Fd())
	return resolveConfigMutationInputWithTerminalPrompt(out, os.Stdin, stdinFD, term.IsTerminal(stdinFD), flagValue, secret, nil)
}

func resolveConfigMutationInputWithTerminalPrompt(out *naistrix.OutputWriter, stdin io.Reader, stdinFD int, isTerminal bool, flagValue string, secret bool, prompt interface{ ReadPassword(fd int) ([]byte, error) }) (string, error) {
	if secret && isTerminal {
		out.Println("Enter new secret value (input hidden):")
	}

	return fasit.ResolveMutationValue(flagValue, secret, stdin, stdinFD, isTerminal, prompt)
}
