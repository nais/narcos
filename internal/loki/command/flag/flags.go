package flag

import "github.com/nais/naistrix"

type Loki struct {
	*naistrix.GlobalFlags
}

type Delete struct {
	*Loki
	Namespace string `name:"namespace" short:"n" usage:"Kubernetes |namespace| of the application."`
	App       string `name:"app" short:"a" usage:"Name of the |application|."`
	Days      int    `name:"days" short:"d" usage:"Delete logs from this many |days| ago until now."`
	Filter    string `name:"filter" short:"f" usage:"Additional LogQL |filter| expression (e.g. 'level=\"error\"')."`
	Regex     string `name:"regex" short:"r" usage:"Case-insensitive |regex| pattern to match log lines."`
}

type List struct {
	*Loki
}
