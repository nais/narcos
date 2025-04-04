package kubeconfig

import (
	"context"

	"github.com/nais/cli/pkg/gcp"
	"github.com/nais/cli/pkg/kubeconfig"
	"github.com/urfave/cli/v3"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:    "kubeconfig",
		Aliases: []string{"kc"},
		Usage:   "Create a kubeconfig file for connecting to available clusters",
		Description: `Create a kubeconfig file for connecting to available clusters.
This requires that you have the gcloud command line tool installed, configured and logged
in using:
gcloud auth login --update-adc`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "overwrite",
				Usage:   "Will overwrite users, clusters, and contexts in your kubeconfig.",
				Aliases: []string{"o"},
			},
			&cli.BoolFlag{
				Name:    "clear",
				Usage:   "Clear existing kubeconfig before writing new data",
				Aliases: []string{"c"},
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
			},
		},
		UseShortOptionHandling: true,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			clear := cmd.Bool("clear")
			overwrite := cmd.Bool("overwrite")
			verbose := cmd.Bool("verbose")

			email, err := gcp.ValidateAndGetUserLogin(ctx, true)
			if err != nil {
				return err
			}

			err = kubeconfig.CreateKubeconfig(ctx, email,
				kubeconfig.WithOverwriteData(overwrite),
				kubeconfig.WithFromScratch(clear),
				kubeconfig.WithOnpremClusters(true),
				kubeconfig.WithCiClusters(true),
				kubeconfig.WithManagementClusters(true),
				kubeconfig.WithPrefixedTenants(true),
				kubeconfig.WithVerboseLogging(verbose))
			if err != nil {
				return err
			}

			return nil
		},
	}
}
