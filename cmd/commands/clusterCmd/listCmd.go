package clusterCmd

import (
	"fmt"
	"github.com/nais/narcos/pkg/gcp"
	"github.com/nais/narcos/pkg/naisdevice"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slices"
	"strings"
)

func listCmd() *cli.Command {
	return &cli.Command{
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
			&cli.BoolFlag{
				Name: "prefixTenant",
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
		UseShortOptionHandling: true,
		Before: func(context *cli.Context) error {
			return gcp.ValidateUserLogin(context.Context)
		},
		Action: func(context *cli.Context) error {
			includeManagement := context.Bool("includeManagement")
			includeOnprem := context.Bool("includeOnprem")
			prefixTenant := context.Bool("prefixTenant")
			tenant := context.String("tenant")

			clusters, err := gcp.GetClusters(context.Context, includeManagement, includeOnprem, tenant)
			if err != nil {
				return err
			}

			for _, cluster := range clusters {
				name := cluster.Name
				if prefixTenant {
					name = cluster.Tenant + "-" + strings.TrimPrefix(name, "nais-")
				}
				fmt.Println(name)
			}

			return nil
		},
	}
}
