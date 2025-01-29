package cluster

import (
	"context"
	"fmt"

	"github.com/nais/narcos/internal/gcp"
	"github.com/urfave/cli/v3"
)

func listCmd() *cli.Command {
	return &cli.Command{
		Name:                   "list",
		Aliases:                []string{"l"},
		UseShortOptionHandling: true,
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			return ctx, gcp.ValidateUserLogin(ctx)
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			clusters, err := gcp.GetClusters(ctx)
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
