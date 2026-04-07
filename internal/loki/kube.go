package loki

import (
	"fmt"

	kubeClient "k8s.io/client-go/tools/clientcmd"
)

// currentContext returns the name of the active kubectl context from the
// local kubeconfig (respects $KUBECONFIG and ~/.kube/config).
func currentContext() (string, error) {
	config, err := kubeClient.NewDefaultClientConfigLoadingRules().Load()
	if err != nil {
		return "", fmt.Errorf("loading kubeconfig: %w", err)
	}
	if config.CurrentContext == "" {
		return "", fmt.Errorf("no active kubectl context found — run 'kubectl config use-context <context>' first")
	}
	return config.CurrentContext, nil
}
