package command

import (
	"context"
	"fmt"

	"github.com/nais/naistrix"
	"github.com/nais/narcos/internal/debug"
	"github.com/nais/narcos/internal/debug/command/flag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		AutoCompleteFunc: func(ctx context.Context, args *naistrix.Arguments, _ string) ([]string, string) {
			if args.Len() >= 1 {
				return nil, ""
			}

			if flags.Namespace == "" {
				return nil, "Specify namespace (-n) to get pod completions."
			}

			typedClient, _, _, err := debug.NewClients(flags.KubeContext)
			if err != nil {
				return nil, "Unable to create Kubernetes client for autocomplete."
			}

			pods, err := typedClient.CoreV1().Pods(flags.Namespace).List(ctx, metav1.ListOptions{
				FieldSelector: "status.phase=Running",
			})
			if err != nil {
				return nil, "Unable to list pods for autocomplete."
			}

			names := make([]string, 0, len(pods.Items))
			for _, pod := range pods.Items {
				names = append(names, pod.Name)
			}

			return names, "Choose the pod to debug."
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
