package main

import (
	"log"
	"os"

	"github.com/nais/narcos/cmd/commands/clusterCmd"
	"github.com/nais/narcos/cmd/commands/deviceCmd"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:        "narc",
		Usage:       "NAIS Administrator CLI",
		Version:     "v0.1",
		Description: "NAIS Administrator CLI",
		Commands: []*cli.Command{
			deviceCmd.Command(),
			clusterCmd.Command(),
		},
		EnableBashCompletion: true,
		HideHelpCommand:      true,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
