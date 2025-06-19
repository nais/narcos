package command

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/nais/cli/pkg/cli"
	"github.com/nais/narcos/internal/gcp"
	"github.com/nais/narcos/internal/kubeconfig"
	"github.com/nais/narcos/internal/kubeconfig/command/flag"
	"github.com/nais/narcos/internal/root"
)

func Kubeconfig(parentFlags *root.Flags) *cli.Command {
	flags := &flag.KubeconfigFlags{Flags: parentFlags}
	return &cli.Command{
		Name:    "kubeconfig",
		Aliases: []string{"kc"},
		Title:   "Create a kubeconfig file for connecting to available clusters.",
		Description: heredoc.Doc(`
			For this command to success you need to be logged in:

			nais login
		`),
		Flags: flags,
		RunFunc: func(ctx context.Context, out cli.Output, args []string) error {
			email, err := gcp.ValidateAndGetUserLogin(ctx, true)
			if err != nil {
				return err
			}

			return kubeconfig.CreateKubeconfig(
				ctx,
				email,
				kubeconfig.WithOverwriteData(flags.Overwrite),
				kubeconfig.WithFromScratch(flags.Clear),
				kubeconfig.WithOnpremClusters(true),
				kubeconfig.WithCiClusters(true),
				kubeconfig.WithManagementClusters(true),
				kubeconfig.WithPrefixedTenants(true),
				kubeconfig.WithVerboseLogging(flags.IsVerbose()),
			)
		},
	}
}
