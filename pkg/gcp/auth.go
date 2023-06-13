package gcp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
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
		io.Copy(os.Stdout, buf)
		return fmt.Errorf("error running '%v' command: %w", cmd.String(), err)
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
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("you are missing Application Default Credentials, run `gcloud auth application-default login` first")
	}

	return nil
}
