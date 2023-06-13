package clusterCmd

import (
	"fmt"
	"github.com/nais/narcos/pkg/gcp"
	"github.com/nais/narcos/pkg/naisdevice"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slices"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:            "cluster",
		Aliases:         []string{"c"},
		Description:     "Operate on NAIS clusters",
		HideHelpCommand: true,
		Subcommands:     subCommands(),
	}
}

func subCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name: "list",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "includeManagement",
					Aliases: []string{"m"},
				},
				&cli.BoolFlag{
					Name:    "includeOnprem",
					Aliases: []string{"o"},
					Value:   true,
				},
				&cli.StringFlag{
					Name:    "tenant",
					Aliases: []string{"t"},
					Action: func(context *cli.Context, tenant string) error {
						if !slices.Contains(naisdevice.Tenants, tenant) {
							return fmt.Errorf("%v is not a valid tenant", tenant)
						}

						return nil
					},
				},
			},
			Before: func(context *cli.Context) error {
				return gcp.ValidateUserLogin(context.Context)
			},
			Action: func(context *cli.Context) error {
				includeManagement := context.Bool("includeManagement")
				includeOnprem := context.Bool("includeOnprem")
				tenant := context.String("tenant")

				_, err := gcp.GetClusters(context.Context, includeManagement, includeOnprem, tenant)
				if err != nil {
					return err
				}

				return nil
			},
		},
	}
}
