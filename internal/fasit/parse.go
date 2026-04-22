package fasit

import (
	"encoding/json"
	"strconv"
	"strings"
)

func ParseConfigValue(configType, rawValue string) (any, error) {
	switch strings.ToUpper(configType) {
	case "INT":
		return strconv.ParseInt(rawValue, 10, 64)
	case "BOOL":
		return strconv.ParseBool(rawValue)
	case "STRING_ARRAY":
		var arr []string
		if err := json.Unmarshal([]byte(rawValue), &arr); err == nil {
			return arr, nil
		}

		parts := strings.Split(rawValue, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}

		return parts, nil
	default:
		return rawValue, nil
	}
}
