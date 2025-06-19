package command

import (
	"context"

	"github.com/nais/cli/pkg/cli"
	"github.com/nais/narcos/internal/jita"
	"github.com/nais/narcos/internal/jita/command/flag"
	"github.com/nais/narcos/internal/root"
)

func Jita(rootFlags *root.Flags) *cli.Command {
	jitaFlags := &flag.JitaFlags{Flags: rootFlags}
	return &cli.Command{
		Name:  "jita",
		Title: "Just-in-time privilege elevation for tenants.",
		SubCommands: []*cli.Command{
			list(jitaFlags),
			grant(jitaFlags),
		},
	}
}

func list(parentFlags *flag.JitaFlags) *cli.Command {
	flags := &flag.ListFlags{
		JitaFlags: parentFlags,
	}
	return &cli.Command{
		Name:  "list",
		Title: "List active and potential privilege elevations",
		Flags: flags,
		Args: []cli.Argument{
			{
				Name:       "TENANT",
				Repeatable: true,
			},
		},
		RunFunc: func(ctx context.Context, out cli.Output, args []string) error {
			return jita.List(ctx, flags, out, args)
		},
	}
}

func grant(parentFlags *flag.JitaFlags) *cli.Command {
	flags := &flag.GrantFlags{
		JitaFlags: parentFlags,
	}
	return &cli.Command{
		Name:        "grant",
		Title:       "Elevate privileges for this tenant",
		Description: "TENANT is one of the tenants given by `narc tenant list`\nENTITLEMENT is one the entitlements given by `narc jita list <TENANT>`\nDURATION is the amount of time you need privileges for, given as 0h0m\nREASON is a human-readable description of why you need to elevate privileges.",
		Flags:       flags,
		Args: []cli.Argument{
			{
				Name:     "tenant",
				Required: true,
			},
			{
				Name:     "entitlement",
				Required: true,
			},
		},
		RunFunc: func(ctx context.Context, out cli.Output, args []string) error {
			return jita.Grant(ctx, flags, out, args)
		},
	}
}
