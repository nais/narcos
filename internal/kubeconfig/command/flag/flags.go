package flag

import (
	"github.com/nais/naistrix"
)

type KubeconfigFlags struct {
	*naistrix.GlobalFlags
	Overwrite bool `name:"overwrite" short:"o" usage:"Will overwrite users, clusters, and contexts in your kubeconfig."`
	Clear     bool `name:"clear" short:"c" usage:"Clear existing kubeconfig before writing new data."`
}
