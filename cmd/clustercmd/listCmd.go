package clustercmd

import (
	"fmt"

	"github.com/nais/narcos/pkg/gcp"
	"github.com/urfave/cli/v2"
)

func listCmd() *cli.Command {
	return &cli.Command{
		Name:                   "list",
		Aliases:                []string{"l"},
		UseShortOptionHandling: true,
		Before: func(context *cli.Context) error {
			return gcp.ValidateUserLogin(context.Context)
		},
		Action: func(context *cli.Context) error {
			clusters, err := gcp.GetClusters(context.Context)
			if err != nil {
				return err
			}

			for _, cluster := range clusters {
				fmt.Println(cluster.Name)
			}

			return nil
		},
	}
}
