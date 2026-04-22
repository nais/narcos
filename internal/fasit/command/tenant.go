package command

import (
	"context"
	"strings"

	"github.com/nais/naistrix"
	"github.com/nais/narcos/internal/fasit"
	"github.com/nais/narcos/internal/fasit/command/flag"
)

func tenantCmd(parentFlags *flag.Fasit) *naistrix.Command {
	return &naistrix.Command{
		Name:  "tenant",
		Title: "Inspect tenants.",
		SubCommands: []*naistrix.Command{
			tenantsListCmd(parentFlags),
			tenantGetCmd(parentFlags),
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
