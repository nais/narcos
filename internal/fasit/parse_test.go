package fasit

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseConfigValueInt(t *testing.T) {
	value, err := ParseConfigValue("INT", "42")
	require.NoError(t, err)
	require.Equal(t, int64(42), value)
}

func TestParseConfigValueBool(t *testing.T) {
	value, err := ParseConfigValue("BOOL", "true")
	require.NoError(t, err)
	require.Equal(t, true, value)
}

func TestParseConfigValueStringArrayJSON(t *testing.T) {
	value, err := ParseConfigValue("STRING_ARRAY", "[\"a\",\"b\"]")
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b"}, value)
}

func TestParseConfigValueStringArrayCommaSeparated(t *testing.T) {
	value, err := ParseConfigValue("STRING_ARRAY", "a, b, c")
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b", "c"}, value)
}

func TestParseConfigValueString(t *testing.T) {
	value, err := ParseConfigValue("STRING", "hello")
	require.NoError(t, err)
	require.Equal(t, "hello", value)
}

func TestParseConfigValueEmptyTypeDefaultsToString(t *testing.T) {
	value, err := ParseConfigValue("", "hello")
	require.NoError(t, err)
	require.Equal(t, "hello", value)
}
