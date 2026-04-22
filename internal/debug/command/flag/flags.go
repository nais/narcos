package flag

import "github.com/nais/naistrix"

type Debug struct {
	*naistrix.GlobalFlags
	Namespace       string   `name:"namespace" short:"n" usage:"Namespace of the target pod."`
	Image           string   `name:"image" short:"i" usage:"Debug container image." default:"europe-north1-docker.pkg.dev/nais-io/nais/images/debug:latest"`
	Cap             []string `name:"cap" usage:"Additional capabilities beyond NET_RAW,SYS_PTRACE."`
	TargetContainer string   `name:"target" short:"t" usage:"Target container for process namespace sharing. Defaults to first container."`
	KubeContext     string   `name:"context" usage:"Kubernetes context to use."`
}
