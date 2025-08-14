package application

import (
	"context"
	"io"

	"github.com/nais/naistrix"
	jita "github.com/nais/narcos/internal/jita/command"
	kubeconfig "github.com/nais/narcos/internal/kubeconfig/command"
	"github.com/nais/narcos/internal/root"
	tenant "github.com/nais/narcos/internal/tenant/command"
	"github.com/nais/narcos/internal/version"
)

func Run(ctx context.Context, w io.Writer) error {
	flags := &root.Flags{}
	app := &naistrix.Application{
		Name:        "narc",
		Title:       "Nais Administrator CLI",
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			kubeconfig.Kubeconfig(flags),
			tenant.Tenant(flags),
			jita.Jita(flags),
		},
		Version: version.Version,
	}
	return app.Run(naistrix.RunWithContext(ctx), naistrix.RunWithOutput(naistrix.NewWriter(w)))
}
