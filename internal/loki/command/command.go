package command

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/nais/naistrix"
	"github.com/nais/narcos/internal/loki"
	"github.com/nais/narcos/internal/loki/command/flag"
)

func Loki(globalFlags *naistrix.GlobalFlags) *naistrix.Command {
	lokiFlags := &flag.Loki{GlobalFlags: globalFlags}
	return &naistrix.Command{
		Name:  "loki",
		Title: "Manage Loki log deletion requests.",
		SubCommands: []*naistrix.Command{
			deleteCmd(lokiFlags),
			listCmd(lokiFlags),
		},
	}
}

func deleteCmd(parentFlags *flag.Loki) *naistrix.Command {
	flags := &flag.Delete{Loki: parentFlags}
	return &naistrix.Command{
		Name:  "delete",
		Title: "Submit a log deletion request to Loki.",
		Description: heredoc.Doc(`
			Submits a log deletion request to the Loki compactor for the given application.

			There is one Loki instance per cluster, so make sure your kubeconfig is pointing
			at the correct cluster before running this command.

			The command will port-forward to the Loki compactor (loki-compactor-0:3100 in the
			nais-system namespace), send the deletion request, and print the updated list of
			pending delete requests.
		`),
		Flags: flags,
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			return loki.Delete(ctx, flags, out)
		},
	}
}

func listCmd(parentFlags *flag.Loki) *naistrix.Command {
	flags := &flag.List{Loki: parentFlags}
	return &naistrix.Command{
		Name:  "list",
		Title: "List pending log deletion requests in Loki.",
		Description: heredoc.Doc(`
			Lists all pending log deletion requests from the Loki compactor.

			There is one Loki instance per cluster, so make sure your kubeconfig is pointing
			at the correct cluster before running this command.
		`),
		Flags: flags,
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			return loki.List(ctx, out)
		},
	}
}
