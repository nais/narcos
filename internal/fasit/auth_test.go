package fasit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadRefreshTokenReturnsEmptyWhenStoreMissing(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	token, err := LoadRefreshToken()
	require.NoError(t, err)
	require.Empty(t, token)
}

func TestSaveRefreshTokenPersistsToken(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	require.NoError(t, SaveRefreshToken("refresh-token-value"))

	token, err := LoadRefreshToken()
	require.NoError(t, err)
	require.Equal(t, "refresh-token-value", token)

	data, err := os.ReadFile(filepath.Join(homeDir, ".config", "narc", "fasit-token.json"))
	require.NoError(t, err)
	require.JSONEq(t, `{"refresh_token":"refresh-token-value"}`, string(data))
}
