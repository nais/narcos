package tenant

import (
	"fmt"
	"strings"

	"github.com/nais/narcos/pkg/naisdevice"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slices"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:            "tenant",
		Aliases:         []string{"t"},
		Usage:           "Manage tenants.",
		HideHelpCommand: true,
		Subcommands:     subCommands(),
	}
}

func subCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "list",
			Usage: "narc device tenant list",
			Action: func(_ *cli.Context) error {
				for _, tenant := range naisdevice.Tenants {
					fmt.Println(tenant)
				}
				return nil
			},
		},
		{
			Name:      "set",
			Usage:     "narc device tenant set [tenant]",
			ArgsUsage: "name of the tenant",
			Action: func(ctx *cli.Context) error {
				if ctx.Args().Len() != 1 {
					return fmt.Errorf("missing required arguments: tenant name")
				}

				tenant := strings.TrimSpace(ctx.Args().First())
				if !slices.Contains(naisdevice.Tenants, tenant) {
					return fmt.Errorf("unknown tenant %v, must be one of: %v", tenant, naisdevice.Tenants)
				}

				err := naisdevice.SetTenant(ctx.Context, tenant)
				if err != nil {
					return err
				}

				fmt.Println("Tenant has been set to ", tenant)

				return nil
			},
		},
	}
}
