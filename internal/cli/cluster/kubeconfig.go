package cluster

import (
	"fmt"
	"strings"

	"github.com/nais/narcos/internal/gcp"
	"github.com/nais/narcos/internal/kubeconfig"
	"github.com/urfave/cli/v2"
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
		Before: func(context *cli.Context) error {
			return gcp.ValidateUserLogin(context.Context)
		},
		Action: func(context *cli.Context) error {
			overwrite := context.Bool("overwrite")
			clean := context.Bool("clean")
			verbose := context.Bool("verbose")

			fmt.Println("Getting clusters...")
			clusters, err := gcp.GetClusters(context.Context)
			if err != nil {
				return err
			}

			if len(clusters) == 0 {
				return fmt.Errorf("no clusters found")
			}

			fmt.Printf("Found %v clusters\n", len(clusters))

			emails, err := gcp.GetUserEmails(context.Context)
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
