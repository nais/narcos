package kubeconfig

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/go-logr/logr"
	"github.com/nais/narcos/internal/gcp"
	kubeClient "k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog/v2"
)

func CreateKubeconfig(email string, clusters []gcp.Cluster, overwrite, clean, verbose bool) error {
	configLoad := kubeClient.NewDefaultClientConfigLoadingRules()

	// If KUBECONFIG is set, but the file does not exist, kubeClient will throw a warning.
	// since we're creating the file, we can safely ignore this warning.
	klog.SetLogger(logr.Discard())

	config, err := configLoad.Load()
	if err != nil {
		return err
	}

	if clean {
		config.AuthInfos = map[string]*api.AuthInfo{}
		config.Contexts = map[string]*api.Context{}
		config.Clusters = map[string]*api.Cluster{}
	}

	err = addUsers(config, clusters, email, overwrite, verbose)
	if err != nil {
		return err
	}

	err = addClusters(config, clusters, email, overwrite, verbose)
	if err != nil {
		return err
	}

	err = kubeClient.WriteToFile(*config, configLoad.GetDefaultFilename())
	if err != nil {
		return err
	}

	fmt.Println("Kubeconfig written to", configLoad.GetDefaultFilename())

	for _, user := range config.AuthInfos {
		if user == nil || user.Exec == nil {
			continue
		}
		_, err = exec.LookPath(user.Exec.Command)
		if err != nil {
			fmt.Printf("%v\nWARNING: %v not found in PATH.\n", os.Stderr, user.Exec.Command)
			fmt.Printf("%v\n%v\n", os.Stderr, user.Exec.InstallHint)
		}
	}
	return nil
}
