package tenant

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/narcos/internal/naisdevice"
	"github.com/urfave/cli/v3"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:            "tenant",
		Usage:           "Work with different Nais tenants.",
		HideHelpCommand: true,
		Commands:        subCommands(),
	}
}

func subCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "list",
			Usage: "narc tenant list",
			Action: func(ctx context.Context, _ *cli.Command) error {
				tenants, err := naisdevice.ListTenants(ctx)
				if err != nil {
					return err
				}
				for _, tenant := range tenants {
					fmt.Println(tenant)
				}
				return nil
			},
		},
		{
			Name:      "set",
			Usage:     "narc tenant set <tenant>",
			ArgsUsage: "name of the tenant",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				if cmd.Args().Len() != 1 {
					return fmt.Errorf("missing required arguments: tenant name")
				}

				tenant := strings.TrimSpace(cmd.Args().First())

				err := naisdevice.SetTenant(ctx, tenant)
				if err != nil {
					return err
				}

				fmt.Println("Tenant has been set to ", tenant)

				return nil
			},
		},
		{
			Name:        "get",
			Usage:       "narc tenant get",
			Description: "Gets the name of the currently active tenant",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				tenant, err := naisdevice.GetTenant(ctx)
				if err != nil {
					return err
				}

				fmt.Println(tenant)

				return nil
			},
		},
	}
}
