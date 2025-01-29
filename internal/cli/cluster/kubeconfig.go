package cluster

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/narcos/internal/gcp"
	"github.com/nais/narcos/internal/kubeconfig"
	"github.com/urfave/cli/v3"
)

func kubeconfigCmd() *cli.Command {
	return &cli.Command{
		Name:    "kubeconfig",
		Aliases: []string{"kc"},
		Description: `Create a kubeconfig file for connecting to available clusters.
This requires that you have the gcloud command line tool installed, configured and logged in using:
gcloud auth login --update-adc`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "overwrite",
				Usage: "Will overwrite users, clusters, and contexts in your kubeconfig.",
			},
			&cli.BoolFlag{
				Name:  "clean",
				Usage: "Recreate the entire kubeconfig.",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
			},
		},
		UseShortOptionHandling: true,
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			return ctx, gcp.ValidateUserLogin(ctx)
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			overwrite := cmd.Bool("overwrite")
			clean := cmd.Bool("clean")
			verbose := cmd.Bool("verbose")

			fmt.Println("Getting clusters...")
			clusters, err := gcp.GetClusters(ctx)
			if err != nil {
				return err
			}

			if len(clusters) == 0 {
				return fmt.Errorf("no clusters found")
			}

			fmt.Printf("Found %v clusters\n", len(clusters))

			emails, err := gcp.GetUserEmails(ctx)
			if err != nil {
				return err
			}

			hasSuffix := func(emails []string, suffix string) string {
				for _, email := range emails {
					if strings.HasSuffix(email, suffix) {
						return email
					}
				}
				panic("no user with suffix " + suffix + " found")
			}

			err = kubeconfig.CreateKubeconfig(hasSuffix(emails, "@nais.io"), clusters, overwrite, clean, verbose)
			if err != nil {
				return err
			}

			return nil
		},
	}
}
