package cli

import (
	"context"
	"log"
	"os"

	"github.com/nais/narcos/internal/cli/jita"
	"github.com/nais/narcos/internal/cli/kubeconfig"
	"github.com/nais/narcos/internal/cli/tenant"
	"github.com/urfave/cli/v3"
)

var (
	// Is set during build
	version = "local"
	commit  = "uncommited"
)

func Run() {
	app := &cli.Command{
		Name:        "narc",
		Usage:       "Nais Administrator CLI",
		Version:     version + "-" + commit,
		Description: "Nais Administrator CLI",
		Commands: []*cli.Command{
			tenant.Command(),
			kubeconfig.Command(),
			jita.Command(),
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
