package cluster

import (
	"github.com/urfave/cli/v3"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:            "cluster",
		Aliases:         []string{"c"},
		Description:     "Operate on Nais clusters",
		HideHelpCommand: true,
		Commands:        subCommands(),
	}
}

func subCommands() []*cli.Command {
	return []*cli.Command{
		kubeconfigCmd(),
		listCmd(),
	}
}
