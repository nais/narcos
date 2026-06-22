package flag

import "github.com/nais/naistrix"

type Loki struct {
	*naistrix.GlobalFlags
}

type Delete struct {
	*Loki
	Days      int    `name:"days" short:"d" usage:"Delete logs from this many days ago until now."`
	StartTime string `name:"start" short:"s" usage:"Starttime for deletion, until now, or end (format: 2006-01-02 15:04:05)."`
	EndTime   string `name:"end" short:"e" usage:"Endtime for deletion, used in range with --start (format: 2006-01-02 15:04:05)."`
	Query     string `name:"query" short:"q" usage:"Loki query code, copy/paste from Grafana."`
}

type List struct {
	*Loki
}
