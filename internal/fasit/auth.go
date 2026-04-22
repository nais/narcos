package fasit

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	fasitgraphql "github.com/nais/narcos/internal/fasit/graphql"
)

const (
	OAuthTokenEndpoint = "https://oauth2.googleapis.com/token" // #nosec G101 -- not a credential
	OAuthAuthEndpoint  = "https://accounts.google.com/o/oauth2/v2/auth"
	OAuthScopes        = "openid email"
	OAuthRedirectURI   = "http://localhost:4444"

	tokenStoreRelativePath = ".config/narc/fasit-token.json" // #nosec G101 -- not a credential
)

var (
	DesktopOAuthClientID     string
	DesktopOAuthClientSecret string
)

type TokenStore struct {
	path string
}

type tokenStorePayload struct {
	RefreshToken string `json:"refresh_token"`
}

func init() {
	DesktopOAuthClientID = os.Getenv("NARC_FASIT_OAUTH_CLIENT_ID")
	DesktopOAuthClientSecret = os.Getenv("NARC_FASIT_OAUTH_CLIENT_SECRET")
	fasitgraphql.ConfigureDesktopOAuth(DesktopOAuthClientID, DesktopOAuthClientSecret, OAuthTokenEndpoint, LoadRefreshToken)
}

func LoadRefreshToken() (string, error) {
	store, err := defaultTokenStore()
	if err != nil {
		return "", err
	}

	return store.LoadRefreshToken()
}

func SaveRefreshToken(token string) error {
	store, err := defaultTokenStore()
	if err != nil {
		return err
	}

	return store.SaveRefreshToken(token)
}

func (s TokenStore) LoadRefreshToken() (string, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read token store %q: %w", s.path, err)
	}

	var payload tokenStorePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", fmt.Errorf("decode token store %q: %w", s.path, err)
	}

	return payload.RefreshToken, nil
}

func (s TokenStore) SaveRefreshToken(token string) error {
	if token == "" {
		return fmt.Errorf("refresh token is empty")
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("create token store directory: %w", err)
	}

	data, err := json.Marshal(tokenStorePayload{RefreshToken: token})
	if err != nil {
		return fmt.Errorf("encode token store payload: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0o600); err != nil {
		return fmt.Errorf("write token store %q: %w", s.path, err)
	}

	return nil
}

func defaultTokenStore() (TokenStore, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return TokenStore{}, fmt.Errorf("resolve user home directory: %w", err)
	}

	return TokenStore{path: filepath.Join(homeDir, tokenStoreRelativePath)}, nil
}
