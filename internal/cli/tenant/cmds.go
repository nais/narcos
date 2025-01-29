package tenant

import (
	"fmt"
	"strings"

	"github.com/nais/narcos/internal/naisdevice"
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:            "tenant",
		Usage:           "Work with different Nais tenants.",
		HideHelpCommand: true,
		Subcommands:     subCommands(),
	}
}

func subCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "list",
			Usage: "narc tenant list",
			Action: func(ctx *cli.Context) error {
				tenants, err := naisdevice.ListTenants(ctx.Context)
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
			Action: func(ctx *cli.Context) error {
				if ctx.Args().Len() != 1 {
					return fmt.Errorf("missing required arguments: tenant name")
				}

				tenant := strings.TrimSpace(ctx.Args().First())

				err := naisdevice.SetTenant(ctx.Context, tenant)
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
			Action: func(ctx *cli.Context) error {
				tenant, err := naisdevice.GetTenant(ctx.Context)
				if err != nil {
					return err
				}

				fmt.Println(tenant)

				return nil
			},
		},
	}
}
