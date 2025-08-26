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
	jitaFlags := &flag.Jita{Flags: rootFlags}
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
		RunFunc: func(ctx context.Context, out naistrix.Output, _ []string) error {
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
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			return jita.Grant(ctx, flags, args[0], args[1])
		},
	}
}
