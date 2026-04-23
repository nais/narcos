package flag

import (
	"context"

	"github.com/nais/naistrix"
	"github.com/nais/narcos/internal/debug"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

type Namespace string

func (*Namespace) AutoComplete(ctx context.Context, _ *naistrix.Arguments, _ string, flags any) ([]string, string) {
	f, ok := flags.(*Debug)
	if !ok {
		return nil, ""
	}

	client, _, _, err := debug.NewClients(string(f.KubeContext))
	if err != nil {
		return nil, "Unable to create Kubernetes client."
	}

	nsList, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, "Unable to list namespaces."
	}

	names := make([]string, 0, len(nsList.Items))
	for _, ns := range nsList.Items {
		names = append(names, ns.Name)
	}

	return names, "Choose the namespace."
}

type KubeContext string

func (*KubeContext) AutoComplete(_ context.Context, _ *naistrix.Arguments, _ string, _ any) ([]string, string) {
	config, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	if err != nil {
		return nil, "Unable to load kubeconfig."
	}

	names := make([]string, 0, len(config.Contexts))
	for name := range config.Contexts {
		names = append(names, name)
	}

	return names, "Choose the Kubernetes context."
}

type TargetContainer string

func (*TargetContainer) AutoComplete(ctx context.Context, args *naistrix.Arguments, _ string, flags any) ([]string, string) {
	f, ok := flags.(*Debug)
	if !ok {
		return nil, ""
	}

	podName := args.Get("pod")
	if podName == "" {
		return nil, "Specify the pod name first."
	}

	ns := string(f.Namespace)
	if ns == "" {
		return nil, "Specify namespace (-n) to get container completions."
	}

	client, _, _, err := debug.NewClients(string(f.KubeContext))
	if err != nil {
		return nil, "Unable to create Kubernetes client."
	}

	pod, err := client.CoreV1().Pods(ns).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, "Unable to get pod."
	}

	names := make([]string, 0, len(pod.Spec.Containers))
	for _, c := range pod.Spec.Containers {
		names = append(names, c.Name)
	}

	return names, "Choose the target container."
}

type Debug struct {
	*naistrix.GlobalFlags
	Namespace       Namespace       `name:"namespace" short:"n" usage:"Namespace of the target pod."`
	Image           string          `name:"image" short:"i" usage:"Debug container image."`
	Cap             []string        `name:"cap" usage:"Additional capabilities beyond NET_RAW,SYS_PTRACE."`
	TargetContainer TargetContainer `name:"target" short:"t" usage:"Target container for process namespace sharing. Defaults to first container."`
	KubeContext     KubeContext     `name:"context" usage:"Kubernetes context to use."`
}

func NewDebug(globalFlags *naistrix.GlobalFlags) *Debug {
	f := &Debug{
		GlobalFlags: globalFlags,
		Image:       debug.DefaultImage,
	}

	config, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	if err != nil {
		return f
	}

	f.KubeContext = KubeContext(config.CurrentContext)
	if ctx, ok := config.Contexts[config.CurrentContext]; ok && ctx.Namespace != "" {
		f.Namespace = Namespace(ctx.Namespace)
	}

	return f
}
