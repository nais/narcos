package gcp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func ValidateUserLogin(ctx context.Context) error {
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
		return fmt.Errorf("%v\nerror running '%v' command: %w", buf.String(), cmd.String(), err)
	}

	user := strings.TrimSpace(buf.String())
	if !strings.HasSuffix(user, "@nais.io") {
		return fmt.Errorf("active gcloud-user is not a nais.io-user: %v", user)
	}

	_, exists := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS")
	if exists {
		return nil
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	_, err = os.Stat(homedir + "/.config/gcloud/application_default_credentials.json")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("you are missing Application Default Credentials, run `gcloud auth application-default login` first")
		}
		return err
	}

	return nil
}

func GetUserEmails(ctx context.Context) ([]string, error) {
	args := []string{
		"auth",
		"list",
		"--format", "value(account)",
	}

	buf := &bytes.Buffer{}
	cmd := exec.CommandContext(ctx, "gcloud", args...)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("%v\nerror running '%v' command: %w", buf.String(), cmd.String(), err)
	}

	users := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(users) == 0 {
		return nil, fmt.Errorf("no users found, are you logged in")
	}

	return users, nil
}
