package command

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/nais/naistrix"
	"github.com/nais/narcos/internal/fasit"
	"github.com/nais/narcos/internal/fasit/command/flag"
)

type oauthCallbackResult struct {
	code string
	err  error
}

type oauthTokenResponse struct {
	RefreshToken     string `json:"refresh_token"`
	IDToken          string `json:"id_token"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func loginCmd(_ *flag.Fasit) *naistrix.Command {
	return &naistrix.Command{
		Name:  "login",
		Title: "Log in to Fasit via desktop OAuth2.",
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			loginCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
			defer cancel()

			state, err := randomBase64URL(32)
			if err != nil {
				return fmt.Errorf("generate oauth state: %w", err)
			}

			codeVerifier, err := randomBase64URL(64)
			if err != nil {
				return fmt.Errorf("generate PKCE code verifier: %w", err)
			}

			codeChallenge := pkceCodeChallenge(codeVerifier)
			authURL := fasit.OAuthAuthEndpoint + "?" + url.Values{
				"client_id":             {fasit.DesktopOAuthClientID},
				"response_type":         {"code"},
				"scope":                 {fasit.OAuthScopes},
				"access_type":           {"offline"},
				"redirect_uri":          {fasit.OAuthRedirectURI},
				"code_challenge":        {codeChallenge},
				"code_challenge_method": {"S256"},
				"cred_ref":              {"true"},
				"state":                 {state},
			}.Encode()

			callbackCh, stopCallback, err := startOAuthCallbackServer(loginCtx, state)
			if err != nil {
				return err
			}
			defer stopCallback()

			out.Println("Open this URL to log in to Fasit:")
			out.Println(authURL)
			out.Println("")

			if err := openBrowser(loginCtx, authURL); err != nil {
				out.Println(fmt.Sprintf("Unable to open a browser automatically: %v", err))
				out.Println("Open the URL above manually to continue.")
				out.Println("")
			}

			out.Println("Waiting for OAuth callback on http://localhost:4444 ...")

			var code string
			select {
			case result := <-callbackCh:
				if result.err != nil {
					return result.err
				}
				code = result.code
			case <-loginCtx.Done():
				return fmt.Errorf("timed out waiting for OAuth callback: %w", loginCtx.Err())
			}

			response, err := exchangeAuthorizationCode(loginCtx, code, codeVerifier)
			if err != nil {
				return err
			}

			if response.RefreshToken == "" {
				return fmt.Errorf("oauth token response did not include a refresh_token")
			}

			if err := fasit.SaveRefreshToken(response.RefreshToken); err != nil {
				return err
			}

			out.Println("Fasit login successful. Refresh token saved.")
			return nil
		},
	}
}

func startOAuthCallbackServer(ctx context.Context, expectedState string) (<-chan oauthCallbackResult, func(), error) {
	listener, err := net.Listen("tcp", "127.0.0.1:4444")
	if err != nil {
		return nil, nil, fmt.Errorf("listen for OAuth callback on 127.0.0.1:4444: %w", err)
	}

	results := make(chan oauthCallbackResult, 1)
	mux := http.NewServeMux()
	server := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		switch {
		case query.Get("error") != "":
			http.Error(w, "Fasit login failed. You can close this window.", http.StatusBadRequest)
			sendOAuthCallbackResult(results, oauthCallbackResult{err: fmt.Errorf("oauth authorization failed: %s", query.Get("error"))})
		case query.Get("state") != expectedState:
			http.Error(w, "Invalid OAuth state. You can close this window.", http.StatusBadRequest)
			sendOAuthCallbackResult(results, oauthCallbackResult{err: fmt.Errorf("oauth callback state mismatch")})
		case query.Get("code") == "":
			http.Error(w, "Missing OAuth code. You can close this window.", http.StatusBadRequest)
			sendOAuthCallbackResult(results, oauthCallbackResult{err: fmt.Errorf("oauth callback missing authorization code")})
		default:
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, _ = io.WriteString(w, "Fasit login complete. You can close this window.\n")
			sendOAuthCallbackResult(results, oauthCallbackResult{code: query.Get("code")})
		}
	})

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			sendOAuthCallbackResult(results, oauthCallbackResult{err: fmt.Errorf("oauth callback server failed: %w", err)})
		}
	}()

	stop := func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}

	go func() {
		<-ctx.Done()
		stop()
	}()

	return results, stop, nil
}

func sendOAuthCallbackResult(results chan<- oauthCallbackResult, result oauthCallbackResult) {
	select {
	case results <- result:
	default:
	}
}

func exchangeAuthorizationCode(ctx context.Context, code, codeVerifier string) (*oauthTokenResponse, error) {
	values := url.Values{
		"client_id":     {fasit.DesktopOAuthClientID},
		"client_secret": {fasit.DesktopOAuthClientSecret},
		"code":          {code},
		"redirect_uri":  {fasit.OAuthRedirectURI},
		"grant_type":    {"authorization_code"},
		"code_verifier": {codeVerifier},
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, fasit.OAuthTokenEndpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create OAuth token exchange request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("exchange authorization code for tokens: %w", err)
	}
	defer func() { _ = response.Body.Close() }()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read OAuth token exchange response: %w", err)
	}

	var payload oauthTokenResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode OAuth token exchange response: %w", err)
	}

	if payload.Error != "" {
		return nil, fmt.Errorf("oauth token exchange failed: %s (%s)", payload.Error, payload.ErrorDescription)
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("oauth token exchange failed with status %s", response.Status)
	}

	return &payload, nil
}

func openBrowser(ctx context.Context, authURL string) error {
	var command string
	switch runtime.GOOS {
	case "darwin":
		command = "open"
	case "linux":
		command = "xdg-open"
	default:
		return fmt.Errorf("unsupported platform %q", runtime.GOOS)
	}

	return exec.CommandContext(ctx, command, authURL).Run() // #nosec G204 -- command is a fixed platform binary
}

func randomBase64URL(size int) (string, error) {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func pkceCodeChallenge(codeVerifier string) string {
	sum := sha256.Sum256([]byte(codeVerifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
