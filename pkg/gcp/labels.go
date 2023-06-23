package gcp

import "strings"

type Kind int64

const (
	KindOnprem Kind = iota
	KindKNADA
	KindNAIS
	KindLegacy
	KindManagment
	KindUnknown
)

func ParseKind(in string) Kind {
	switch strings.ToLower(in) {
	case "knada":
		return KindKNADA
	case "onprem":
		return KindOnprem
	case "nais":
		return KindNAIS
	case "legacy":
		return KindLegacy
	case "managment":
		return KindManagment
	default:
		return KindUnknown
	}
}

type Environment int64

const (
	EnvironmentCi Environment = iota
	EnvironmentCiFSS
	EnvironmentCiGCP
	EnvironmentDev
	EnvironmentDevFSS
	EnvironmentDevGCP
	EnvironmentProd
	EnvironmentProdFSS
	EnvironmentProdGCP
	EnvironmentStaging
	EnvironmentUnknown
)

func (env Environment) String() string {
	switch env {
	case EnvironmentCi:
		return "ci"
	case EnvironmentCiFSS:
		return "ci-fss"
	case EnvironmentCiGCP:
		return "ci-gcp"
	case EnvironmentDev:
		return "dev"
	case EnvironmentDevFSS:
		return "dev-fss"
	case EnvironmentDevGCP:
		return "dev-gcp"
	case EnvironmentProd:
		return "prod"
	case EnvironmentProdFSS:
		return "prod-fss"
	case EnvironmentProdGCP:
		return "prod-gcp"
	case EnvironmentStaging:
		return "prod-fss"
	default:
		return "Unknown"
	}
}

func ParseEnvironment(in string) Environment {
	switch strings.ToLower(in) {
	case "ci":
		return EnvironmentCi
	case "ci-fss":
		return EnvironmentCiFSS
	case "ci-gcp":
		return EnvironmentCiGCP
	case "dev":
		return EnvironmentDev
	case "dev-fss":
		return EnvironmentDevFSS
	case "dev-gcp":
		return EnvironmentDevGCP
	case "prod":
		return EnvironmentProd
	case "prod-fss":
		return EnvironmentProdFSS
	case "prod-gcp":
		return EnvironmentProdGCP
	case "staging":
		return EnvironmentStaging
	default:
		return EnvironmentUnknown
	}
}

func GetClusterServerForLegacyGCP(env Environment) string {
	switch env {
	case EnvironmentProdGCP:
		return "https://10.255.240.6"
	case EnvironmentDevGCP:
		return "https://10.255.240.5"
	case EnvironmentCiGCP:
		return "https://10.255.240.7"
	default:
		return "unknown-cluster"
	}
}
