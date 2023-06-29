package devicecmd

import (
	"github.com/nais/narcos/cmd/devicecmd/tenant"
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:            "device",
		Aliases:         []string{"d"},
		Description:     "Manage Naisdevice from the terminal.",
		HideHelpCommand: true,
		Subcommands:     subCommands(),
	}
}

func subCommands() []*cli.Command {
	return []*cli.Command{
		tenant.Command(),
	}
}
