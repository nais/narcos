package application

import (
	"context"
	"io"
	"os"

	"github.com/nais/cli/pkg/cli"
	jita "github.com/nais/narcos/internal/jita/command"
	"github.com/nais/narcos/internal/root"
)

func newApplication(flags *root.Flags) *cli.Application {
	return &cli.Application{
		Name:    "narc",
		Title:   "Nais Administrator CLI",
		Version: getVersion(),
		SubCommands: []*cli.Command{
			jita.Jita(flags),
		},
		StickyFlags: flags,
	}
}

func Run(ctx context.Context, w io.Writer) error {
	flags := &root.Flags{}
	_, err := newApplication(flags).Run(ctx, cli.NewWriter(w), os.Args[1:])
	return err
}
