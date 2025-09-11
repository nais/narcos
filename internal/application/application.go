package application

import (
	"context"
	"fmt"

	"github.com/nais/naistrix"
	jita "github.com/nais/narcos/internal/jita/command"
	kubeconfig "github.com/nais/narcos/internal/kubeconfig/command"
	tenant "github.com/nais/narcos/internal/tenant/command"
	"github.com/nais/narcos/internal/version"
)

func Run(ctx context.Context) error {
	app, flags, err := naistrix.NewApplication(
		"narc",
		"Nais Administrator CLI",
		version.Version,
	)
	if err != nil {
		return fmt.Errorf("unable to create application: %w", err)
	}

	err = app.AddCommand(
		kubeconfig.Kubeconfig(flags),
		tenant.Tenant(flags),
		jita.Jita(flags),
	)
	if err != nil {
		return fmt.Errorf("unable to add command: %w", err)
	}

	return app.Run(naistrix.RunWithContext(ctx))
}
