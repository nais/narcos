package fasit

type Tenant struct {
	ID           string
	Name         string
	Environments []Environment
	Icon         string
}

type Environment struct {
	ID           string
	Name         string
	Description  *string
	Created      string
	LastModified string
	Kind         string
	GCPProjectID *string
	Reconcile    bool
	Features     []Feature
	Values       []EnvironmentValue
}

type EnvironmentValue struct {
	Key   string
	Value string
}

type Feature struct {
	Name             string
	Chart            string
	Version          string
	Source           string
	Description      string
	EnvironmentKinds []string
	Enabled          bool
	Dependencies     []Dependency
	Configuration    *Configuration
	Configurations   []ConfigurationDetail
}

type Configuration struct {
	Configuration []ConfigurationItem
}

type ConfigurationDetail struct {
	ID     string
	Key    string
	Value  string
	Source string
}

type ConfigurationItem struct {
	ID      string
	Value   ConfigValue
	Content any
	Source  string
}

type ConfigValue struct {
	Key         string
	DisplayName string
	Description string
	Required    bool
	Config      *ConfigMeta
	Computed    *ComputedMeta
}

type ConfigMeta struct {
	Type   string
	Secret bool
}

type ComputedMeta struct {
	Template string
}

type Dependency struct {
	AnyOf []string
	AllOf []string
}

type Rollout struct {
	FeatureName  string         `json:"featureName" yaml:"featureName"`
	Version      string         `json:"version" yaml:"version"`
	Status       string         `json:"status" yaml:"status"`
	Created      string         `json:"created" yaml:"created"`
	Completed    string         `json:"completed,omitempty" yaml:"completed,omitempty"`
	Target       string         `json:"target,omitempty" yaml:"target,omitempty"`
	DeploymentID string         `json:"deploymentId,omitempty" yaml:"deploymentId,omitempty"`
	Events       []RolloutEvent `json:"events,omitempty" yaml:"events,omitempty"`
	Logs         []RolloutLog   `json:"logs,omitempty" yaml:"logs,omitempty"`
}

type RolloutDetail struct {
	ID          string         `json:"id" yaml:"id"`
	FeatureName string         `json:"featureName" yaml:"featureName"`
	Version     string         `json:"version" yaml:"version"`
	Status      string         `json:"status" yaml:"status"`
	Created     string         `json:"created" yaml:"created"`
	Completed   string         `json:"completed,omitempty" yaml:"completed,omitempty"`
	Events      []RolloutEvent `json:"events,omitempty" yaml:"events,omitempty"`
	Logs        []RolloutLog   `json:"logs,omitempty" yaml:"logs,omitempty"`
}

type RolloutSummary struct {
	FeatureName  string `json:"featureName" yaml:"featureName"`
	Version      string `json:"version" yaml:"version"`
	Status       string `json:"status" yaml:"status"`
	Created      string `json:"created" yaml:"created"`
	Completed    string `json:"completed,omitempty" yaml:"completed,omitempty"`
	Target       string `json:"target,omitempty" yaml:"target,omitempty"`
	DeploymentID string `json:"deploymentId,omitempty" yaml:"deploymentId,omitempty"`
}

type RolloutEvent struct {
	Message string `json:"message" yaml:"message"`
	Failure bool   `json:"failure" yaml:"failure"`
	Created string `json:"created,omitempty" yaml:"created,omitempty"`
}

type RolloutLog struct {
	TenantName  string    `json:"tenantName" yaml:"tenantName"`
	Environment string    `json:"environment" yaml:"environment"`
	Lines       []LogLine `json:"lines,omitempty" yaml:"lines,omitempty"`
}

type LogLine struct {
	Timestamp string `json:"timestamp" yaml:"timestamp"`
	Message   string `json:"message" yaml:"message"`
}

type FeatureLog struct {
	CurrentVersion string        `json:"currentVersion" yaml:"currentVersion"`
	CurrentStatus  string        `json:"currentStatus" yaml:"currentStatus"`
	LastModified   string        `json:"lastModified" yaml:"lastModified"`
	CurrentLog     []LogLine     `json:"currentLog,omitempty" yaml:"currentLog,omitempty"`
	HelmDiff       HelmValueDiff `json:"helmDiff" yaml:"helmDiff"`
}

type HelmValueDiff struct {
	Difference string `json:"difference" yaml:"difference"`
	Diff       string `json:"diff" yaml:"diff"`
}

type Deployment struct {
	ID          string             `json:"id" yaml:"id"`
	FeatureName string             `json:"featureName" yaml:"featureName"`
	Version     string             `json:"version" yaml:"version"`
	Description string             `json:"description,omitempty" yaml:"description,omitempty"`
	Created     string             `json:"created" yaml:"created"`
	Target      string             `json:"target" yaml:"target"`
	Statuses    []DeploymentStatus `json:"statuses,omitempty" yaml:"statuses,omitempty"`
}

type DeploymentStatus struct {
	TenantName      string `json:"tenantName" yaml:"tenantName"`
	EnvironmentName string `json:"environmentName" yaml:"environmentName"`
	State           string `json:"state" yaml:"state"`
	Message         string `json:"message" yaml:"message"`
	LastModified    string `json:"lastModified" yaml:"lastModified"`
}

type FeatureStatus struct {
	Tenant      string `json:"tenant" yaml:"tenant"`
	Environment string `json:"environment" yaml:"environment"`
	Kind        string `json:"kind" yaml:"kind"`
	Enabled     bool   `json:"enabled" yaml:"enabled"`
}
