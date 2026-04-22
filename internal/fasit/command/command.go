package command

import (
	"context"

	"github.com/nais/naistrix"
	"github.com/nais/narcos/internal/fasit"
	"github.com/nais/narcos/internal/fasit/command/flag"
)

type (
	fasitTenantLister      func(context.Context) ([]fasit.Tenant, error)
	fasitEnvironmentGetter func(context.Context, string, string) (*fasit.Environment, error)
	fasitFeatureLister     func(context.Context) ([]fasit.Feature, error)
	fasitRolloutLister     func(context.Context, string) ([]fasit.Rollout, error)
)

func Fasit(globalFlags *naistrix.GlobalFlags) *naistrix.Command {
	fasitFlags := &flag.Fasit{GlobalFlags: globalFlags}
	return &naistrix.Command{
		Name:        "fasit",
		Title:       "Manage Fasit configuration.",
		StickyFlags: fasitFlags,
		SubCommands: []*naistrix.Command{
			loginCmd(fasitFlags),
			tenantCmd(fasitFlags),
			envCmd(fasitFlags),
			featureCmd(fasitFlags),
			rolloutCmd(fasitFlags),
			deploymentCmd(fasitFlags),
		},
	}
}
