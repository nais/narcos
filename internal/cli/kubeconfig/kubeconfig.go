package kubeconfig

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/narcos/internal/gcp"
	"github.com/nais/narcos/internal/kubeconfig"
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
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			return ctx, gcp.ValidateUserLogin(ctx)
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			overwrite := cmd.Bool("overwrite")
			clear := cmd.Bool("clear")
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

			var currentUser string
			for _, email := range emails {
				if strings.HasSuffix(email, "@nais.io") {
					currentUser = email
				}
			}

			if currentUser == "" {
				return fmt.Errorf("no user found with nais.io email")
			}

			err = kubeconfig.CreateKubeconfig(currentUser, clusters, overwrite, clear, verbose)
			if err != nil {
				return err
			}

			return nil
		},
	}
}
