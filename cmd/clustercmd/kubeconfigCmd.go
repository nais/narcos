package clustercmd

import (
	"fmt"

	"github.com/nais/narcos/pkg/gcp"
	"github.com/nais/narcos/pkg/kubeconfig"
	"github.com/nais/narcos/pkg/naisdevice"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slices"
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
				Name:    "excludeManagement",
				Aliases: []string{"m"},
			},
			&cli.BoolFlag{
				Name:    "excludeOnprem",
				Aliases: []string{"o"},
			},
			&cli.BoolFlag{
				Name:    "excludeKnada",
				Aliases: []string{"k"},
			},
			&cli.BoolFlag{
				Name:  "prefixTenant",
				Value: true,
			},
			&cli.BoolFlag{
				Name: "skipNAVPrefix",
			},
			&cli.BoolFlag{
				Name:  "overwrite",
				Usage: "Will overwrite users, clusters, and contexts in your kubeconfig.",
			},
			&cli.BoolFlag{
				Name:    "seperateAdmin",
				Aliases: []string{"s"},
				Usage:   "Seperate cluster with admin user from cluster with team access. Required both a NAV and NAIS e-mail.",
			},
			&cli.BoolFlag{
				Name:  "clean",
				Usage: "Create kubeconfig from a clean slate, will remove all customization",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
			},
			&cli.StringFlag{
				Name:    "tenant",
				Aliases: []string{"t"},
				Usage:   "Specify which tenant you want config for. Default behaviour is all tenants.",
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
			excludeManagement := context.Bool("excludeManagement")
			excludeOnprem := context.Bool("excludeOnprem")
			excludeKnada := context.Bool("excludeKnada")
			prefixTenant := context.Bool("prefixTenant")
			skipNAVPrefix := context.Bool("skipNAVPrefix")
			overwrite := context.Bool("overwrite")
			seperateAdmin := context.Bool("seperateAdmin")
			clean := context.Bool("clean")
			verbose := context.Bool("verbose")
			tenant := context.String("tenant")

			fmt.Println("Getting clusters...")
			clusters, err := gcp.GetClusters(context.Context, !excludeManagement, !excludeOnprem, !excludeKnada, prefixTenant, skipNAVPrefix, tenant)
			if err != nil {
				return err
			}

			if len(clusters) == 0 {
				return fmt.Errorf("no clusters found")
			}

			fmt.Printf("Found %v clusters\n", len(clusters))

			emails, err := gcp.GetUserEmails(context.Context)

			err = kubeconfig.CreateKubeconfig(emails, clusters, overwrite, excludeOnprem, clean, verbose, seperateAdmin)
			if err != nil {
				return err
			}

			return nil
		},
	}
}
