package flag

import "github.com/nais/naistrix"

type Fasit struct {
	*naistrix.GlobalFlags
	Output string `name:"output" short:"o" usage:"Output |format|. Valid values: table, json, yaml."`
}

type TenantsList struct {
	*Fasit
	Tenant string `name:"tenant" usage:"Filter by tenant |name|."`
	Kind   string `name:"kind" usage:"Filter by environment |kind|."`
}

type TenantGet struct {
	*Fasit
}

type EnvGet struct {
	*Fasit
}

type FeaturesList struct {
	*Fasit
	Kind    string `name:"kind" usage:"Filter by environment |kind| compatibility."`
	Feature string `name:"feature" usage:"Filter by feature |name|."`
}

type FeatureGet struct {
	*Fasit
}

type FeatureStatus struct {
	*Fasit
	Tenant  string `name:"tenant" usage:"Filter by tenant |name|."`
	Env     string `name:"env" usage:"Filter by environment |name|."`
	Kind    string `name:"kind" usage:"Filter by environment |kind|."`
	Enabled string `name:"enabled" usage:"Filter by enabled state (true|false)."`
}

type FeatureRollouts struct {
	*Fasit
}

type EnvFeatureGet struct {
	*Fasit
}

type EnvFeatureLogs struct {
	*Fasit
}

type EnvFeatureHelm struct {
	*Fasit
}

type EnvFeatureRollouts struct {
	*Fasit
}

type EnvFeatureAudit struct {
	*Fasit
}

type EnvFeatureEnable struct {
	*Fasit
	Yes bool `name:"yes" short:"y" usage:"Skip confirmation prompt."`
}

type EnvFeatureDisable struct {
	*Fasit
	Yes bool `name:"yes" short:"y" usage:"Skip confirmation prompt."`
}

type EnvFeatureConfigSet struct {
	*Fasit
	Value string `name:"value" usage:"New value for the configuration. Required for non-secret configs; secret configs must be provided via stdin or the hidden prompt."`
	Yes   bool   `name:"yes" short:"y" usage:"Skip confirmation prompt."`
}

type EnvFeatureConfigOverride struct {
	*Fasit
	Value string `name:"value" usage:"Value for the configuration override. Required for non-secret configs; secret configs must be provided via stdin or the hidden prompt."`
	Yes   bool   `name:"yes" short:"y" usage:"Skip confirmation prompt."`
}

type RolloutsList struct {
	*Fasit
	Feature string `name:"feature" usage:"Filter by feature |name|."`
	Status  string `name:"status" usage:"Filter by rollout |status|."`
}

type RolloutGet struct {
	*Fasit
}

type DeploymentGet struct {
	*Fasit
}
