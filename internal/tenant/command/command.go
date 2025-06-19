package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/pkg/cli"
	"github.com/nais/narcos/internal/naisdevice"
	"github.com/nais/narcos/internal/root"
	"github.com/nais/narcos/internal/tenant/command/flag"
)

func Tenant(rootFlags *root.Flags) *cli.Command {
	tenantFlags := &flag.TenantFlags{Flags: rootFlags}
	return &cli.Command{
		Name:  "tenant",
		Title: "Work with different Nais tenants.",
		SubCommands: []*cli.Command{
			list(tenantFlags),
			set(tenantFlags),
			get(tenantFlags),
		},
	}
}

func list(parentFlags *flag.TenantFlags) *cli.Command {
	flags := &flag.ListFlags{TenantFlags: parentFlags}
	return &cli.Command{
		Name:  "list",
		Title: "List tenants.",
		Flags: flags,
		RunFunc: func(ctx context.Context, out cli.Output, args []string) error {
			tenants, err := naisdevice.ListTenants(ctx)
			if err != nil {
				return err
			}
			for _, tenant := range tenants {
				fmt.Println(tenant)
			}
			return nil
		},
	}
}

func set(parentFlags *flag.TenantFlags) *cli.Command {
	flags := &flag.SetFlags{TenantFlags: parentFlags}
	return &cli.Command{
		Name:         "set",
		Title:        "Set the active tenant.",
		ValidateFunc: cli.ValidateExactArgs(1),
		Args: []cli.Argument{
			{
				Name:     "tenant",
				Required: true,
			},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, out cli.Output, args []string) error {
			if err := naisdevice.SetTenant(ctx, args[0]); err != nil {
				return err
			}

			out.Println("Tenant has been set to ", args[0])
			return nil
		},
	}
}

func get(parentFlags *flag.TenantFlags) *cli.Command {
	flags := &flag.GetFlags{TenantFlags: parentFlags}
	return &cli.Command{
		Name:  "get",
		Title: "Get the active tenant.",
		Flags: flags,
		RunFunc: func(ctx context.Context, out cli.Output, args []string) error {
			tenant, err := naisdevice.GetTenant(ctx)
			if err != nil {
				return err
			}

			out.Println(tenant)
			return nil
		},
	}
}
