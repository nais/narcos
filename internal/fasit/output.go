package fasit

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/nais/naistrix"
	naistrixoutput "github.com/nais/naistrix/output"
)

const (
	OutputFormatTable = "table"
	OutputFormatJSON  = "json"
	OutputFormatYAML  = "yaml"
)

type ConfigurationOutput struct {
	ID               string `json:"id" yaml:"id"`
	Key              string `json:"key" yaml:"key"`
	DisplayName      string `json:"displayName" yaml:"displayName"`
	Description      string `json:"description" yaml:"description"`
	Required         bool   `json:"required" yaml:"required"`
	Type             string `json:"type,omitempty" yaml:"type,omitempty"`
	Secret           bool   `json:"secret" yaml:"secret"`
	Source           string `json:"source" yaml:"source"`
	ComputedTemplate string `json:"computedTemplate,omitempty" yaml:"computedTemplate,omitempty"`
	Value            any    `json:"value,omitempty" yaml:"value,omitempty"`
}

type FeatureOutput struct {
	Name             string                `json:"name" yaml:"name"`
	Chart            string                `json:"chart" yaml:"chart"`
	Version          string                `json:"version" yaml:"version"`
	Source           string                `json:"source" yaml:"source"`
	Description      string                `json:"description" yaml:"description"`
	EnvironmentKinds []string              `json:"environmentKinds,omitempty" yaml:"environmentKinds,omitempty"`
	Enabled          bool                  `json:"enabled" yaml:"enabled"`
	Dependencies     []Dependency          `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	Configuration    []ConfigurationOutput `json:"configuration,omitempty" yaml:"configuration,omitempty"`
	Configurations   []ConfigurationDetail `json:"configurations,omitempty" yaml:"configurations,omitempty"`
}

func NormalizeOutputFormat(format string) string {
	if format == "" {
		return OutputFormatTable
	}

	return strings.ToLower(format)
}

func RenderStructuredOutput(out *naistrix.OutputWriter, format string, tableRows, data any) error {
	switch NormalizeOutputFormat(format) {
	case OutputFormatTable:
		return renderTable(out, tableRows)
	case OutputFormatJSON:
		return out.JSON(naistrixoutput.JSONWithPrettyOutput()).Render(data)
	case OutputFormatYAML:
		return out.YAML().Render(data)
	default:
		return fmt.Errorf("unsupported output format: %q", format)
	}
}

func RenderDataOutput(out *naistrix.OutputWriter, format string, data any) error {
	switch NormalizeOutputFormat(format) {
	case OutputFormatJSON:
		return out.JSON(naistrixoutput.JSONWithPrettyOutput()).Render(data)
	case OutputFormatYAML:
		return out.YAML().Render(data)
	default:
		return fmt.Errorf("unsupported output format: %q", format)
	}
}

func MaskedConfigurationItems(configuration *Configuration) []ConfigurationOutput {
	if configuration == nil {
		return nil
	}

	items := make([]ConfigurationOutput, 0, len(configuration.Configuration))
	for _, item := range configuration.Configuration {
		configType := ""
		secret := false
		if item.Value.Config != nil {
			configType = item.Value.Config.Type
			secret = item.Value.Config.Secret
		}

		computedTemplate := ""
		if item.Value.Computed != nil {
			computedTemplate = item.Value.Computed.Template
		}

		items = append(items, ConfigurationOutput{
			ID:               item.ID,
			Key:              item.Value.Key,
			DisplayName:      item.Value.DisplayName,
			Description:      item.Value.Description,
			Required:         item.Value.Required,
			Type:             configType,
			Secret:           secret,
			Source:           item.Source,
			ComputedTemplate: computedTemplate,
			Value:            MaskConfigurationValue(item),
		})
	}

	return items
}

func MaskedFeatureOutput(feature *Feature) *FeatureOutput {
	if feature == nil {
		return nil
	}

	return &FeatureOutput{
		Name:             feature.Name,
		Chart:            feature.Chart,
		Version:          feature.Version,
		Source:           feature.Source,
		Description:      feature.Description,
		EnvironmentKinds: append([]string(nil), feature.EnvironmentKinds...),
		Enabled:          feature.Enabled,
		Dependencies:     append([]Dependency(nil), feature.Dependencies...),
		Configuration:    MaskedConfigurationItems(feature.Configuration),
		Configurations:   append([]ConfigurationDetail(nil), feature.Configurations...),
	}
}

func MaskConfigurationValue(item ConfigurationItem) any {
	if item.Value.Config != nil && item.Value.Config.Secret {
		return "***"
	}

	return decodeConfigurationContent(item.Content)
}

func FormatDisplayValue(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	default:
		b, err := json.Marshal(v)
		if err == nil {
			return string(b)
		}

		return fmt.Sprint(v)
	}
}

func decodeConfigurationContent(content any) any {
	switch value := content.(type) {
	case nil:
		return nil
	case *json.RawMessage:
		if value == nil {
			return nil
		}
		return decodeRawMessage(*value)
	case json.RawMessage:
		return decodeRawMessage(value)
	default:
		return value
	}
}

func decodeRawMessage(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}

	var value any
	if err := json.Unmarshal(raw, &value); err == nil {
		return value
	}

	return string(raw)
}

func renderTable(out *naistrix.OutputWriter, rows any) error {
	v := reflect.ValueOf(rows)
	if !v.IsValid() || v.Kind() != reflect.Slice {
		return out.Table().Render(rows)
	}

	if v.Len() == 0 {
		out.Println("No results")
		return nil
	}

	return out.Table().Render(rows)
}
