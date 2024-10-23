package kubeconfig

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/nais/narcos/internal/gcp"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func addClusters(config *clientcmdapi.Config, clusters []gcp.Cluster, email string, overwrite, verbose bool) error {
	for _, cluster := range clusters {
		err := addCluster(config, cluster, overwrite, verbose)
		if err != nil {
			return err
		}

		addContext(config, cluster, overwrite, verbose, email)
	}

	return nil
}

func addCluster(config *clientcmdapi.Config, cluster gcp.Cluster, overwrite, verbose bool) error {
	if _, ok := config.Clusters[cluster.Name]; ok && !overwrite {
		if verbose {
			fmt.Printf("Cluster %q already exists in kubeconfig, skipping\n", cluster.Name)
		}
		return nil
	}

	var (
		ca  []byte
		err error
	)
	if len(cluster.CA) > 0 {
		ca, err = base64.StdEncoding.DecodeString(cluster.CA)
		if err != nil {
			return err
		}
	}

	kubeconfigCluster := &clientcmdapi.Cluster{
		Server:                   cluster.Endpoint,
		CertificateAuthorityData: ca,
	}
	isLegacy := cluster.Kind == gcp.KindLegacy || (strings.EqualFold("nav", cluster.Tenant) && strings.HasSuffix(cluster.Name, "gcp"))

	if isLegacy {
		kubeconfigCluster.CertificateAuthorityData = nil
		kubeconfigCluster.InsecureSkipTLSVerify = true
		kubeconfigCluster.Server = gcp.GetClusterServerForLegacyGCP(cluster.Name)
	}

	config.Clusters[cluster.Name] = kubeconfigCluster

	fmt.Printf("Added cluster %v to config\n", cluster.Name)

	return nil
}
