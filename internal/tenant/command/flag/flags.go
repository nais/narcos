package flag

import (
	"github.com/nais/narcos/internal/root"
)

type TenantFlags struct {
	*root.Flags
}

type ListFlags struct {
	*TenantFlags
}

type SetFlags struct {
	*TenantFlags
}

type GetFlags struct {
	*TenantFlags
}
