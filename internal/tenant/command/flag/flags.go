package flag

import (
	"github.com/nais/naistrix"
)

type TenantFlags struct {
	*naistrix.GlobalFlags
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
