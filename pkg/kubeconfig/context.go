package kubeconfig

import (
	"fmt"

	"github.com/nais/narcos/pkg/gcp"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func addContext(config *clientcmdapi.Config, cluster gcp.Cluster, overwrite, verbose bool, email string) {
	contextName := cluster.Name
	namespace := "nais-system"

	if _, ok := config.Contexts[contextName]; ok && !overwrite {
		if verbose {
			fmt.Printf("Context %q already exists in kubeconfig, skipping\n", contextName)
		}
		return
	}

	user := email
	if cluster.Kind == gcp.KindOnprem {
		user = cluster.User.UserName
	}

	config.Contexts[contextName] = &clientcmdapi.Context{
		Cluster:   cluster.Name,
		AuthInfo:  user,
		Namespace: namespace,
	}

	fmt.Printf("Added context %v for %v to config\n", contextName, user)
}
