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
