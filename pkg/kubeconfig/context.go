package kubeconfig

import (
	"fmt"
	"github.com/nais/narcos/pkg/gcp"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"strings"
)

func addContext(config *clientcmdapi.Config, cluster gcp.Cluster, overwrite, seperateAdmin, verbose bool, email string) {
	if contextShouldNotBeInKubeconfig(email, cluster) {
		return
	}

	contextName := cluster.Name
	namespace := "default"
	if contextShouldHaveSeperateAdminContext(email, cluster, seperateAdmin) {
		contextName += "-nais"
		namespace = "nais-system"
	}

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

func contextShouldNotBeInKubeconfig(email string, cluster gcp.Cluster) bool {
	return (isEmailNais(email) && cluster.Kind == gcp.KindKNADA) ||
		(isEmailNav(email) && cluster.Tenant != "nav") ||
		(isEmailNav(email) && cluster.Kind == gcp.KindManagment) ||
		(isEmailNav(email) && cluster.Name == "ci-gcp")
}

func contextShouldHaveSeperateAdminContext(email string, cluster gcp.Cluster, seperateAdmin bool) bool {
	return seperateAdmin &&
		isEmailNais(email) &&
		cluster.Tenant == "nav" &&
		strings.HasSuffix(cluster.Environment, "-gcp")
}
