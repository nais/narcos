package application

import (
	"context"
	"io"
	"os"

	"github.com/nais/naistrix"
	jita "github.com/nais/narcos/internal/jita/command"
	kubeconfig "github.com/nais/narcos/internal/kubeconfig/command"
	"github.com/nais/narcos/internal/root"
	tenant "github.com/nais/narcos/internal/tenant/command"
	"github.com/nais/narcos/internal/version"
)

func newApplication(flags *root.Flags) *naistrix.Application {
	return &naistrix.Application{
		Name:    "narc",
		Title:   "Nais Administrator CLI",
		Version: version.Version,
		SubCommands: []*naistrix.Command{
			kubeconfig.Kubeconfig(flags),
			tenant.Tenant(flags),
			jita.Jita(flags),
		},
		StickyFlags: flags,
	}
}

func Run(ctx context.Context, w io.Writer) error {
	flags := &root.Flags{}
	_, err := newApplication(flags).Run(ctx, naistrix.NewWriter(w), os.Args[1:])
	return err
}
