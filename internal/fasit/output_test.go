package fasit

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMaskedConfigurationItemsMasksSecretValues(t *testing.T) {
	raw := json.RawMessage(`"super-secret"`)
	items := MaskedConfigurationItems(&Configuration{Configuration: []ConfigurationItem{{
		ID: "cfg-1",
		Value: ConfigValue{
			Key:         "password",
			DisplayName: "Password",
			Description: "Sensitive",
			Config: &ConfigMeta{
				Type:   "STRING",
				Secret: true,
			},
		},
		Content: &raw,
		Source:  "ENV",
	}}})

	require.Len(t, items, 1)
	require.Equal(t, "***", items[0].Value)
}

func TestMaskedConfigurationItemsPreservesNonSecretValues(t *testing.T) {
	raw := json.RawMessage(`123`)
	items := MaskedConfigurationItems(&Configuration{Configuration: []ConfigurationItem{{
		ID: "cfg-2",
		Value: ConfigValue{
			Key:         "replicas",
			DisplayName: "Replicas",
			Description: "Replica count",
			Config: &ConfigMeta{
				Type:   "INT",
				Secret: false,
			},
		},
		Content: &raw,
		Source:  "GLOBAL",
	}}})

	require.Len(t, items, 1)
	require.Equal(t, float64(123), items[0].Value)
}

func TestMaskedFeatureOutputMasksSecretConfigurationValues(t *testing.T) {
	raw := json.RawMessage(`"super-secret"`)
	feature := &Feature{
		Name:    "my-feature",
		Chart:   "demo",
		Version: "1.2.3",
		Configuration: &Configuration{Configuration: []ConfigurationItem{{
			ID: "cfg-1",
			Value: ConfigValue{
				Key:         "password",
				DisplayName: "Password",
				Description: "Sensitive",
				Config: &ConfigMeta{
					Type:   "STRING",
					Secret: true,
				},
			},
			Content: &raw,
			Source:  "GLOBAL",
		}}},
		Configurations: []ConfigurationDetail{{ID: "cfg-1", Key: "password", Value: "***", Source: "GLOBAL"}},
	}

	masked := MaskedFeatureOutput(feature)
	require.NotNil(t, masked)
	require.Len(t, masked.Configuration, 1)
	require.Equal(t, "***", masked.Configuration[0].Value)
	require.Equal(t, "***", masked.Configurations[0].Value)
}
