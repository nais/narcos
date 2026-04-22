package debug

import (
	"fmt"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewClients(kubeContext string) (kubernetes.Interface, dynamic.Interface, *rest.Config, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	overrides := &clientcmd.ConfigOverrides{}
	if kubeContext != "" {
		overrides.CurrentContext = kubeContext
	}

	loader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)
	config, err := loader.ClientConfig()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("loading kubeconfig: %w", err)
	}

	typed, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("creating kubernetes client: %w", err)
	}

	dyn, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("creating dynamic client: %w", err)
	}

	return typed, dyn, config, nil
}
