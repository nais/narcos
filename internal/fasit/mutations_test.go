package fasit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpdateConfigurationMutationBuildsTypedPayload(t *testing.T) {
	client := &mockGraphQLClient{}
	err := (&store{client: client}).updateConfiguration(context.Background(), "cfg-1", int64(42))
	require.NoError(t, err)
	require.Equal(t, "UpdateConfigurationMutation", client.lastReq.OpName)

	variables, err := json.Marshal(client.lastReq.Variables)
	require.NoError(t, err)
	require.JSONEq(t, `{"id":"cfg-1","configuration":{"description":null,"value":42}}`, string(variables))
}

func TestCreateConfigurationMutationBuildsTypedPayload(t *testing.T) {
	client := &mockGraphQLClient{}
	err := (&store{client: client}).createConfiguration(context.Background(), "env-1", "my-feature", "hosts", []string{"a", "b"})
	require.NoError(t, err)
	require.Equal(t, "CreateConfigurationMutation", client.lastReq.OpName)

	variables, err := json.Marshal(client.lastReq.Variables)
	require.NoError(t, err)
	require.JSONEq(t, `{"configuration":{"environmentID":"env-1","feature":"my-feature","description":null,"key":"hosts","value":["a","b"]}}`, string(variables))
}

func TestSetFeatureStateMutationBuildsPayload(t *testing.T) {
	client := &mockGraphQLClient{}
	err := (&store{client: client}).setFeatureState(context.Background(), "env-1", "my-feature", true)
	require.NoError(t, err)
	require.Equal(t, "SetFeatureStateMutation", client.lastReq.OpName)

	variables, err := json.Marshal(client.lastReq.Variables)
	require.NoError(t, err)
	require.JSONEq(t, `{"envID":"env-1","feature":"my-feature","enabled":true}`, string(variables))
}

func TestConfirmMutationRequiresConfirmation(t *testing.T) {
	out := &recordingOutput{}
	err := ConfirmMutation(out, strings.NewReader("n\n"), false, "About to mutate")
	require.ErrorIs(t, err, ErrMutationAborted)
	require.Equal(t, "About to mutate\nProceed? [y/N]: \n", out.String())
}

func TestConfirmMutationHonorsYesFlag(t *testing.T) {
	out := &recordingOutput{}
	err := ConfirmMutation(out, strings.NewReader(""), true, "About to mutate")
	require.NoError(t, err)
	require.Equal(t, "About to mutate\n", out.String())
}

func TestConfirmMutationAcceptsUppercaseY(t *testing.T) {
	out := &recordingOutput{}
	err := ConfirmMutation(out, strings.NewReader("Y\n"), false, "About to mutate")
	require.NoError(t, err)
	require.Equal(t, "About to mutate\nProceed? [y/N]: \n", out.String())
}

func TestResolveMutationValue(t *testing.T) {
	t.Run("non-secret requires value flag", func(t *testing.T) {
		_, err := ResolveMutationValue("", false, strings.NewReader("ignored\n"), 0, false, nil)
		require.EqualError(t, err, "--value is required for non-secret configuration mutations")
	})

	t.Run("non-secret keeps argv value", func(t *testing.T) {
		value, err := ResolveMutationValue("plain", false, strings.NewReader("ignored\n"), 0, false, nil)
		require.NoError(t, err)
		require.Equal(t, "plain", value)
	})

	t.Run("secret rejects value flag", func(t *testing.T) {
		_, err := ResolveMutationValue("super-secret", true, strings.NewReader("ignored\n"), 0, false, nil)
		require.EqualError(t, err, "secret configuration values must not be passed with --value; provide the value via stdin or the hidden prompt")
	})

	t.Run("secret reads from stdin when non-interactive", func(t *testing.T) {
		value, err := ResolveMutationValue("", true, strings.NewReader("s3cr3t\n"), 0, false, nil)
		require.NoError(t, err)
		require.Equal(t, "s3cr3t", value)
	})

	t.Run("secret trims trailing newline before typed parsing", func(t *testing.T) {
		raw, err := ResolveMutationValue("", true, strings.NewReader("1, 2, 3\n"), 0, false, nil)
		require.NoError(t, err)

		parsed, err := ParseConfigValue("STRING_ARRAY", raw)
		require.NoError(t, err)
		require.Equal(t, []string{"1", "2", "3"}, parsed)
	})

	t.Run("secret interactive path uses hidden prompt", func(t *testing.T) {
		value, err := ResolveMutationValue("", true, strings.NewReader("ignored\n"), 7, true, secretPromptStub{value: []byte("hidden\n")})
		require.NoError(t, err)
		require.Equal(t, "hidden", value)
	})

	t.Run("secret interactive path surfaces prompt errors", func(t *testing.T) {
		_, err := ResolveMutationValue("", true, strings.NewReader("ignored\n"), 7, true, secretPromptStub{err: errors.New("tty failed")})
		require.EqualError(t, err, "read secret value: tty failed")
	})

	t.Run("secret stdin cannot be empty", func(t *testing.T) {
		_, err := ResolveMutationValue("", true, strings.NewReader("\n"), 0, false, nil)
		require.EqualError(t, err, "secret configuration value cannot be empty")
	})

	t.Run("secret stdin read errors are surfaced", func(t *testing.T) {
		_, err := ResolveMutationValue("", true, errReader{}, 0, false, nil)
		require.EqualError(t, err, "read secret value from stdin: boom")
	})

	t.Run("secret confirmation display is masked", func(t *testing.T) {
		require.Equal(t, "***", DisplayMutationValue("super-secret", true))
		require.Equal(t, "42", DisplayMutationValue(int64(42), false))
	})
}

type recordingOutput struct {
	bytes.Buffer
}

type secretPromptStub struct {
	value []byte
	err   error
}

func (s secretPromptStub) ReadPassword(int) ([]byte, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.value, nil
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) {
	return 0, errors.New("boom")
}

func (errReader) ReadString(byte) (string, error) {
	return "", errors.New("boom")
}

func (o *recordingOutput) Println(values ...any) {
	_, _ = o.WriteString(joinValues(values...))
	_, _ = o.WriteString("\n")
}

func joinValues(values ...any) string {
	parts := make([]string, len(values))
	for i, value := range values {
		parts[i] = toString(value)
	}

	return strings.Join(parts, "")
}

func toString(value any) string {
	if s, ok := value.(string); ok {
		return s
	}

	return fmt.Sprint(value)
}
