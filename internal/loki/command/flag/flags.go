package flag

import "github.com/nais/naistrix"

type Loki struct {
	*naistrix.GlobalFlags
}

type Delete struct {
	*Loki
	Namespace string `name:"namespace" short:"n" usage:"Kubernetes namespace of the application."`
	App       string `name:"app" short:"a" usage:"Name of the application."`
	Days      int    `name:"days" short:"d" usage:"Delete logs from this many days ago until now."`
	StartTime string `name:"start" short:"s" usage:"Starttime for deletion, until now, or end (format: 2006-01-02 15:04:05)."`
	EndTime   string `name:"end" short:"e" usage:"Endtime for deletion, used in range with --start (format: 2006-01-02 15:04:05)."`
	Contains  string `name:"contains" short:"c" usage:"Line filtering (e.g. '|= \"personident\"', or '!= \"personident\"')."`
	Filter    string `name:"filter" short:"f" usage:"Additional LogQL label filter expression (e.g. 'level=\"error\"')."`
	Regex     string `name:"regex" short:"r" usage:"Case-insensitive regex pattern to match log lines."`
}

type List struct {
	*Loki
}
