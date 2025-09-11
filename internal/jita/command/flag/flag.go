package flag

import (
	"time"

	"github.com/nais/naistrix"
)

type Jita struct {
	*naistrix.GlobalFlags
}

type List struct {
	*Jita
	Tenants []string `name:"tenant" short:"t" usage:"Specify one or more |tenant|s. If not specified, all supported tenants will be used. Can be repeated."`
}

type Grant struct {
	*Jita
	Duration time.Duration `name:"duration" short:"d" usage:"How long you need privileges for."`
	Reason   string        `name:"reason" short:"r" usage:"Human-readable description of why you need to elevate privileges. This value is read by the tenant."`
}
