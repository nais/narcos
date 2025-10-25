package command

import (
	"context"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/nais/naistrix"
	"github.com/nais/narcos/internal/jita"
	"github.com/nais/narcos/internal/jita/command/flag"
)

func Jita(globalFlags *naistrix.GlobalFlags) *naistrix.Command {
	jitaFlags := &flag.Jita{GlobalFlags: globalFlags}
	return &naistrix.Command{
		Name:  "jita",
		Title: "Just-in-time privilege elevation for tenants.",
		SubCommands: []*naistrix.Command{
			list(jitaFlags),
			grant(jitaFlags),
		},
	}
}

func list(parentFlags *flag.Jita) *naistrix.Command {
	flags := &flag.List{Jita: parentFlags}
	return &naistrix.Command{
		Name:        "list",
		Title:       "List active and potential privilege elevations",
		Description: "To include the roles associated with each entitlement in the output, use verbose (-v) mode.",
		Flags:       flags,
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			return jita.List(ctx, flags, out)
		},
	}
}

func grant(parentFlags *flag.Jita) *naistrix.Command {
	flags := &flag.Grant{Jita: parentFlags}
	return &naistrix.Command{
		Name:  "grant",
		Title: "Elevate privileges for this tenant.",
		Description: heredoc.Doc(`
			ENTITLEMENT is one the entitlements given by "narc jita list TENANT"
			TENANT is one of the tenants given by "narc tenant list"
			DURATION is the amount of time you need privileges for, given as "0h0m"
			REASON is a human-readable description of why you need to elevate privileges.
		`),
		Flags: flags,
		Args: []naistrix.Argument{
			{Name: "entitlement"},
			{Name: "tenant"},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			return jita.Grant(ctx, flags, args.Get("entitlement"), args.Get("tenant"))
		},
	}
}
