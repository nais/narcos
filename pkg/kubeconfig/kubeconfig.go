package kubeconfig

import (
	"fmt"
	"k8s.io/client-go/tools/clientcmd/api"
	"os"
	"os/exec"
	"strings"

	"github.com/go-logr/logr"
	"github.com/nais/narcos/pkg/gcp"
	"golang.org/x/exp/slices"
	kubeClient "k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func CreateKubeconfig(emails []string, clusters []gcp.Cluster, overwrite, excludeOnprem, clean, verbose, seperateAdmin bool) error {
	slices.Sort(emails)
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

	for _, email := range emails {
		err = addUsers(config, clusters, email, overwrite, excludeOnprem, verbose)
		if err != nil {
			return err
		}

		err = addClusters(config, clusters, email, overwrite, seperateAdmin, verbose)
		if err != nil {
			return err
		}
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

func isEmailNais(email string) bool {
	return strings.HasSuffix(email, "@nais.io")
}
func isEmailNav(email string) bool {
	return strings.HasSuffix(email, "@nav.no")
}
