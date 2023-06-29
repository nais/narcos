package clustercmd

import (
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:            "cluster",
		Aliases:         []string{"c"},
		Description:     "Operate on NAIS clusters",
		HideHelpCommand: true,
		Subcommands:     subCommands(),
	}
}

func subCommands() []*cli.Command {
	return []*cli.Command{
		kubeconfigCmd(),
		listCmd(),
	}
}
