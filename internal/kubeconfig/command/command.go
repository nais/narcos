package command

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/nais/naistrix"
	"github.com/nais/narcos/internal/gcp"
	"github.com/nais/narcos/internal/kubeconfig"
	"github.com/nais/narcos/internal/kubeconfig/command/flag"
)

func Kubeconfig(globalFlags *naistrix.GlobalFlags) *naistrix.Command {
	flags := &flag.KubeconfigFlags{GlobalFlags: globalFlags}
	return &naistrix.Command{
		Name:    "kubeconfig",
		Aliases: []string{"kc"},
		Title:   "Create a kubeconfig file for connecting to available clusters.",
		Description: heredoc.Doc(`
			For this command to success you need to be logged in:

			nais login
		`),
		Flags: flags,
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, _ *naistrix.OutputWriter) error {
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
