package command

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/nais/naistrix"
	"github.com/nais/narcos/internal/fasit"
	"github.com/nais/narcos/internal/fasit/command/flag"
)

func featureCmd(parentFlags *flag.Fasit) *naistrix.Command {
	return &naistrix.Command{
		Name:  "feature",
		Title: "Inspect features.",
		SubCommands: []*naistrix.Command{
			featuresListCmd(parentFlags),
			featureGetCmd(parentFlags),
			featureStatusCmd(parentFlags),
			featureRolloutsCmd(parentFlags),
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
