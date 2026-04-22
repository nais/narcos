package command

import (
	"context"
	"fmt"

	"github.com/nais/naistrix"
	"github.com/nais/narcos/internal/debug"
	"github.com/nais/narcos/internal/debug/command/flag"
)

func Debug(globalFlags *naistrix.GlobalFlags) *naistrix.Command {
	flags := &flag.Debug{GlobalFlags: globalFlags}
	return &naistrix.Command{
		Name:  "debug",
		Title: "Attach a privileged debug container to a running pod.",
		Description: "Creates an ephemeral debug container as root with NET_RAW and SYS_PTRACE capabilities. " +
			"Automatically creates Kyverno PolicyExceptions when needed.",
		Flags: flags,
		Args: []naistrix.Argument{
			{Name: "pod"},
		},
		ValidateFunc: func(_ context.Context, args *naistrix.Arguments) error {
			if args.Get("pod") == "" {
				return fmt.Errorf("pod name is required")
			}
			if flags.Namespace == "" {
				return fmt.Errorf("namespace is required (use -n)")
			}
			return nil
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, _ *naistrix.OutputWriter) error {
			return debug.Run(ctx, debug.Options{
				PodName:         args.Get("pod"),
				Namespace:       flags.Namespace,
				Image:           flags.Image,
				ExtraCaps:       flags.Cap,
				TargetContainer: flags.TargetContainer,
				KubeContext:     flags.KubeContext,
			})
		},
	}
}
