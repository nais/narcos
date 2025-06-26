package flag

import (
	"time"

	"github.com/nais/narcos/internal/root"
)

type JitaFlags struct {
	*root.Flags
}

type ListFlags struct {
	*JitaFlags
}

type GrantFlags struct {
	*JitaFlags

	Duration time.Duration `name:"duration" short:"d" usage:"How long you need privileges for."`
	Reason   string        `name:"reason" short:"r" usage:"Human-readable description of why you need to elevate privileges. This value is read by the tenant."`
}
