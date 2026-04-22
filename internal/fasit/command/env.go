package command

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nais/naistrix"
	"github.com/nais/narcos/internal/fasit"
	"github.com/nais/narcos/internal/fasit/command/flag"
	"golang.org/x/term"
)

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
