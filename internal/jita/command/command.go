package command

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/nais/naistrix"
	"github.com/nais/narcos/internal/jita"
	"github.com/nais/narcos/internal/jita/command/flag"
	"github.com/nais/narcos/internal/root"
)

func Jita(rootFlags *root.Flags) *naistrix.Command {
	jitaFlags := &flag.JitaFlags{Flags: rootFlags}
	return &naistrix.Command{
		Name:  "jita",
		Title: "Just-in-time privilege elevation for tenants.",
		SubCommands: []*naistrix.Command{
			list(jitaFlags),
			grant(jitaFlags),
		},
	}
}

func list(parentFlags *flag.JitaFlags) *naistrix.Command {
	flags := &flag.ListFlags{JitaFlags: parentFlags}
	return &naistrix.Command{
		Name:  "list",
		Title: "List active and potential privilege elevations",
		Flags: flags,
		Args: []naistrix.Argument{
			{Name: "tenant", Repeatable: true},
		},
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			return jita.List(ctx, flags, out, args)
		},
	}
}

func grant(parentFlags *flag.JitaFlags) *naistrix.Command {
	flags := &flag.GrantFlags{JitaFlags: parentFlags}
	return &naistrix.Command{
		Name:  "grant",
		Title: "Elevate privileges for this tenant.",
		Description: heredoc.Doc(`
			TENANT is one of the tenants given by "narc tenant list"
			ENTITLEMENT is one the entitlements given by "narc jita list TENANT"
			DURATION is the amount of time you need privileges for, given as "0h0m"
			REASON is a human-readable description of why you need to elevate privileges.
		`),
		Flags: flags,
		Args: []naistrix.Argument{
			{Name: "tenant"},
			{Name: "entitlement"},
		},
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			return jita.Grant(ctx, flags, out, args)
		},
	}
}
