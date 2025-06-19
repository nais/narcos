package application

import (
	"context"
	"io"
	"os"

	"github.com/nais/cli/pkg/cli"
	"github.com/nais/narcos/internal/root"
)

func newApplication(flags *root.Flags) *cli.Application {
	return &cli.Application{
		Name:        "narc",
		Title:       "Narc CLI",
		Version:     getVersion(),
		SubCommands: []*cli.Command{},
		StickyFlags: flags,
	}
}

func Run(ctx context.Context, w io.Writer) error {
	flags := &root.Flags{}
	_, err := newApplication(flags).Run(ctx, cli.NewWriter(w), os.Args[1:])
	return err
}
