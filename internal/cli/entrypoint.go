package cli

import (
	"log"
	"os"

	"github.com/nais/narcos/internal/cli/cluster"
	"github.com/nais/narcos/internal/cli/tenant"
	"github.com/urfave/cli/v2"
)

func Run() {
	app := &cli.App{
		Name:        "narc",
		Usage:       "NAIS Administrator CLI",
		Version:     "v0.1",
		Description: "NAIS Administrator CLI",
		Commands: []*cli.Command{
			tenant.Command(),
			cluster.Command(),
		},
		EnableBashCompletion: true,
		HideHelpCommand:      true,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
