package cli

import (
	"context"
	"log"
	"os"

	"github.com/nais/narcos/internal/cli/cluster"
	"github.com/nais/narcos/internal/cli/tenant"
	"github.com/urfave/cli/v3"
)

func Run() {
	app := &cli.Command{
		Name:        "narc",
		Usage:       "Nais Administrator CLI",
		Version:     "v0.1",
		Description: "Nais Administrator CLI",
		Commands: []*cli.Command{
			tenant.Command(),
			cluster.Command(),
		},
		EnableShellCompletion: true,
		HideHelpCommand:       true,
	}

	ctx := context.Background()
	err := app.Run(ctx, os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
