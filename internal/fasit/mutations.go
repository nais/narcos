package fasit

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	fasitgraphql "github.com/nais/narcos/internal/fasit/graphql"
	"golang.org/x/term"
)

var ErrMutationAborted = errors.New("aborted")

type mutationOutput interface {
	Println(...any)
}

type stdinReader interface {
	ReadString(delim byte) (string, error)
}

type terminalPrompter struct{}

func (terminalPrompter) ReadPassword(fd int) ([]byte, error) {
	return term.ReadPassword(fd)
}

var terminalSecretPrompter interface{ ReadPassword(fd int) ([]byte, error) } = terminalPrompter{}

func UpdateConfiguration(ctx context.Context, configID string, value any) error {
	store, err := newStore(ctx)
	if err != nil {
		return err
	}

	return store.updateConfiguration(ctx, configID, value)
}

func CreateConfiguration(ctx context.Context, envID, feature, key string, value any) error {
	store, err := newStore(ctx)
	if err != nil {
		return err
	}

	return store.createConfiguration(ctx, envID, feature, key, value)
}

func SetFeatureState(ctx context.Context, envID, feature string, enabled bool) error {
	store, err := newStore(ctx)
	if err != nil {
		return err
	}

	return store.setFeatureState(ctx, envID, feature, enabled)
}

func ConfirmMutation(out mutationOutput, in io.Reader, yes bool, summary ...string) error {
	for _, line := range summary {
		out.Println(line)
	}

	if yes {
		return nil
	}

	out.Println("Proceed? [y/N]: ")
	answer, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	if strings.EqualFold(strings.TrimSpace(answer), "y") {
		return nil
	}

	return ErrMutationAborted
}

func ResolveMutationValue(flagValue string, secret bool, stdin io.Reader, stdinFD int, isTerminal bool, prompt interface{ ReadPassword(fd int) ([]byte, error) }) (string, error) {
	if !secret {
		if flagValue == "" {
			return "", errors.New("--value is required for non-secret configuration mutations")
		}

		return flagValue, nil
	}

	if flagValue != "" {
		return "", errors.New("secret configuration values must not be passed with --value; provide the value via stdin or the hidden prompt")
	}

	if isTerminal {
		if prompt == nil {
			prompt = terminalSecretPrompter
		}

		secretValue, err := prompt.ReadPassword(stdinFD)
		if err != nil {
			return "", fmt.Errorf("read secret value: %w", err)
		}

		return strings.TrimRight(string(secretValue), "\r\n"), nil
	}

	reader, ok := stdin.(stdinReader)
	if !ok {
		reader = bufio.NewReader(stdin)
	}

	value, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("read secret value from stdin: %w", err)
	}

	value = strings.TrimRight(value, "\r\n")
	if value == "" {
		return "", errors.New("secret configuration value cannot be empty")
	}

	return value, nil
}

func DisplayMutationValue(value any, secret bool) string {
	if secret {
		return "***"
	}

	return FormatDisplayValue(value)
}

func (s *store) updateConfiguration(ctx context.Context, configID string, value any) error {
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode value: %w", err)
	}

	config := &fasitgraphql.UpdateConfiguration{Value: valueJSON}
	if _, err := fasitgraphql.UpdateConfigurationMutation(ctx, s.client, configID, config); err != nil {
		return fmt.Errorf("update configuration: %w", err)
	}

	return nil
}

func (s *store) createConfiguration(ctx context.Context, envID, feature, key string, value any) error {
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode value: %w", err)
	}

	config := &fasitgraphql.NewConfiguration{
		EnvironmentID: &envID,
		Feature:       feature,
		Key:           key,
		Value:         valueJSON,
	}
	if _, err := fasitgraphql.CreateConfigurationMutation(ctx, s.client, config); err != nil {
		return fmt.Errorf("create configuration: %w", err)
	}

	return nil
}

func (s *store) setFeatureState(ctx context.Context, envID, feature string, enabled bool) error {
	if _, err := fasitgraphql.SetFeatureStateMutation(ctx, s.client, envID, feature, enabled); err != nil {
		return fmt.Errorf("set feature state: %w", err)
	}

	return nil
}
