package gcp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func activeGoogleCredentials(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "gcloud", "auth", "print-access-token")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("running 'gcloud auth print-access-token' command: %w", err)
	}
	// Ensure the output is not an empty token
	if strings.TrimSpace(string(output)) == "" {
		return fmt.Errorf("no active google credentials found")
	}

	return nil
}

func activeNaisUser(ctx context.Context) error {
	args := []string{
		"config",
		"list",
		"account",
		"--format", "value(core.account)",
	}

	buf := &bytes.Buffer{}
	cmd := exec.CommandContext(ctx, "gcloud", args...)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("running gcloud command %q: %w", strings.Join(args, " "), err)
	}

	user := strings.TrimSpace(buf.String())
	if !strings.HasSuffix(user, "@nais.io") {
		return fmt.Errorf("active gcloud user is not a nais.io user: %v", user)
	}

	return nil
}

func ValidateAndGetUserLogin(ctx context.Context, enforceNais bool) (string, error) {
	args := []string{
		"config",
		"list",
		"account",
		"--format", "value(core.account)",
	}

	buf := &bytes.Buffer{}
	cmd := exec.CommandContext(ctx, "gcloud", args...)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%v\nerror running '%v' command: %w", buf.String(), cmd.String(), err)
	}

	user := strings.TrimSpace(buf.String())
	if user == "" {
		return "", fmt.Errorf("missing active user, have you logged in with 'gcloud auth login --update-adc'")
	}

	if enforceNais && !strings.HasSuffix(user, "@nais.io") {
		return "", fmt.Errorf("active gcloud-user is not a nais.io-user: %v", user)
	}

	_, exists := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS")
	if exists {
		return user, nil
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	homedir += "/.config"

	if runtime.GOOS == "windows" {
		homedir = os.ExpandEnv("$APPDATA")
	}

	_, err = os.Stat(filepath.Clean(homedir + "/gcloud/application_default_credentials.json"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("you are missing Application Default Credentials, run `gcloud auth login --update-adc` first")
		}
		return "", err
	}

	return user, nil
}

func ValidateUserLogin(ctx context.Context) error {
	if err := activeGoogleCredentials(ctx); err != nil {
		return fmt.Errorf("checking active google credentials: %w", err)
	}
	if err := activeNaisUser(ctx); err != nil {
		return fmt.Errorf("checking for active Nais user: %w", err)
	}

	return nil
}

func shellCommandOutput(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	token := strings.TrimSpace(string(output))
	return token, nil
}

func GCloudAccessToken(ctx context.Context) (string, error) {
	return shellCommandOutput(ctx, "gcloud", "auth", "print-access-token")
}

func GCloudActiveUser(ctx context.Context) (string, error) {
	return shellCommandOutput(ctx, "gcloud", "auth", "list", "--filter", "status:ACTIVE", "--format", "value(account)")
}
