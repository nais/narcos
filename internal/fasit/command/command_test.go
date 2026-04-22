package command

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/nais/naistrix"
	"github.com/nais/narcos/internal/fasit"
	"github.com/nais/narcos/internal/fasit/command/flag"
)

func runFasitCommand(t *testing.T, args ...string) string {
	t.Helper()

	var buf bytes.Buffer
	app, globalFlags, err := naistrix.NewApplication(
		"narc",
		"test app",
		"test",
		naistrix.ApplicationWithWriter(&buf),
	)
	if err != nil {
		t.Fatalf("new application: %v", err)
	}

	if err := app.AddCommand(Fasit(globalFlags)); err != nil {
		t.Fatalf("add fasit command: %v", err)
	}

	if err := app.Run(naistrix.RunWithArgs(args)); err != nil {
		t.Fatalf("run %q: %v", strings.Join(args, " "), err)
	}

	return buf.String()
}

func TestFasitSharedFlags(t *testing.T) {
	t.Run("root help shows shared flags", func(t *testing.T) {
		out := runFasitCommand(t, "fasit", "--help")

		for _, want := range []string{"--output", "login"} {
			if !strings.Contains(out, want) {
				t.Fatalf("root help missing %q\n%s", want, out)
			}
		}
	})

	t.Run("read command help shows shared flags", func(t *testing.T) {
		out := runFasitCommand(t, "fasit", "env", "feature", "audit", "--help")

		for _, want := range []string{"--output"} {
			if !strings.Contains(out, want) {
				t.Fatalf("audit help missing %q\n%s", want, out)
			}
		}
	})

	t.Run("another read command help shows shared flags", func(t *testing.T) {
		out := runFasitCommand(t, "fasit", "tenant", "list", "--help")

		for _, want := range []string{"--output"} {
			if !strings.Contains(out, want) {
				t.Fatalf("tenant list help missing %q\n%s", want, out)
			}
		}
	})

	t.Run("audit command accepts shared flags at runtime", func(t *testing.T) {
		out := runFasitCommand(
			t,
			"fasit",
			"env",
			"feature",
			"audit",
			"tenant-a",
			"prod",
			"feature-x",
			"--output",
			"json",
		)

		for _, want := range []string{
			`"available": false`,
			`"tenant": "tenant-a"`,
			`"environment": "prod"`,
			`"feature": "feature-x"`,
		} {
			if !strings.Contains(out, want) {
				t.Fatalf("audit output missing %q\n%s", want, out)
			}
		}
	})
}

func TestEnvFeatureAuditPlaceholder(t *testing.T) {
	cmd := envFeatureAuditCmd(&flag.Fasit{})
	if cmd.Name != "audit" {
		t.Fatalf("expected audit command, got %q", cmd.Name)
	}
	if cmd.Flags == nil {
		t.Fatal("expected audit command flags")
	}

	t.Run("table", func(t *testing.T) {
		var buf bytes.Buffer
		out := naistrix.NewOutputWriter(&buf, nil)

		if err := renderEnvFeatureAuditPlaceholder(out, "", "tenant-a", "prod", "feature-x"); err != nil {
			t.Fatalf("render table placeholder: %v", err)
		}

		got := buf.String()
		want := "Audit data is not available: the Fasit backend does not currently expose audit history."
		if !strings.Contains(got, want) {
			t.Fatalf("table output missing placeholder message\nwant: %s\ngot: %s", want, got)
		}
	})

	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		out := naistrix.NewOutputWriter(&buf, nil)

		if err := renderEnvFeatureAuditPlaceholder(out, "json", "tenant-a", "prod", "feature-x"); err != nil {
			t.Fatalf("render json placeholder: %v", err)
		}

		got := buf.String()
		for _, want := range []string{
			`"available": false`,
			`"message": "Audit data is not available: the Fasit backend does not currently expose audit history."`,
			`"tenant": "tenant-a"`,
			`"environment": "prod"`,
			`"feature": "feature-x"`,
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("json output missing %q\ngot: %s", want, got)
			}
		}
	})
}

func TestTenantsListFilters(t *testing.T) {
	tenants := []fasit.Tenant{
		{
			Name:         "tenant-a",
			Environments: []fasit.Environment{{Name: "dev-a", Kind: "dev"}, {Name: "prod-a", Kind: "prod"}},
		},
		{
			Name:         "tenant-b",
			Environments: []fasit.Environment{{Name: "dev-b", Kind: "dev"}},
		},
	}

	t.Run("tenant filter keeps matching tenant", func(t *testing.T) {
		got := filterTenants(tenants, "tenant-b", "")
		if len(got) != 1 || got[0].Name != "tenant-b" {
			t.Fatalf("expected only tenant-b, got %#v", got)
		}
	})

	t.Run("kind filter narrows environments and omits empty tenants", func(t *testing.T) {
		got := filterTenants(tenants, "", "prod")
		want := []fasit.Tenant{{
			Name:         "tenant-a",
			Environments: []fasit.Environment{{Name: "prod-a", Kind: "prod"}},
		}}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected filtered tenants\nwant: %#v\ngot:  %#v", want, got)
		}
	})
}

func TestFeatureStatusFilters(t *testing.T) {
	statuses := []fasit.FeatureStatus{
		{Tenant: "tenant-a", Environment: "dev-a", Kind: "dev", Enabled: true},
		{Tenant: "tenant-a", Environment: "prod-a", Kind: "prod", Enabled: false},
		{Tenant: "tenant-b", Environment: "prod-b", Kind: "prod", Enabled: true},
	}

	t.Run("combines tenant env kind and enabled filters", func(t *testing.T) {
		got, err := filterFeatureStatuses(statuses, "tenant-b", "prod-b", "prod", "true")
		if err != nil {
			t.Fatalf("filter feature statuses: %v", err)
		}

		want := []fasit.FeatureStatus{{Tenant: "tenant-b", Environment: "prod-b", Kind: "prod", Enabled: true}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected filtered statuses\nwant: %#v\ngot:  %#v", want, got)
		}
	})

	t.Run("rejects invalid enabled filter", func(t *testing.T) {
		_, err := filterFeatureStatuses(statuses, "", "", "", "maybe")
		if err == nil || !strings.Contains(err.Error(), "invalid --enabled value") {
			t.Fatalf("expected invalid enabled error, got %v", err)
		}
	})
}

func TestRolloutFilters(t *testing.T) {
	rollouts := []fasit.RolloutSummary{
		{FeatureName: "feature-a", Version: "1.0.0", Status: "CREATED"},
		{FeatureName: "feature-a", Version: "1.0.1", Status: "DEPLOYED"},
		{FeatureName: "feature-b", Version: "2.0.0", Status: "FAILED"},
	}

	t.Run("feature filter keeps exact feature matches", func(t *testing.T) {
		got := filterRolloutSummaries(rollouts, "feature-b", "")
		want := []fasit.RolloutSummary{{FeatureName: "feature-b", Version: "2.0.0", Status: "FAILED"}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected filtered rollouts\nwant: %#v\ngot:  %#v", want, got)
		}
	})

	t.Run("status filter is case insensitive", func(t *testing.T) {
		got := filterRolloutSummaries(rollouts, "", "deployed")
		want := []fasit.RolloutSummary{{FeatureName: "feature-a", Version: "1.0.1", Status: "DEPLOYED"}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected filtered rollouts\nwant: %#v\ngot:  %#v", want, got)
		}
	})
}

func TestEnvironmentFeatureRolloutsBehavior(t *testing.T) {
	rollouts := []fasit.Rollout{
		{FeatureName: "feature-a", Version: "1.0.0", DeploymentID: "dep-1", Target: "tenant=tenant-a, environment=prod"},
		{FeatureName: "feature-a", Version: "1.0.1", DeploymentID: "dep-2", Target: "tenant=tenant-a, environment=dev"},
		{FeatureName: "feature-a", Version: "1.0.2"},
	}

	t.Run("filters deployment-backed history to matching env labels", func(t *testing.T) {
		got := filterEnvironmentFeatureRollouts(rollouts, "tenant-a", "prod")
		want := []fasit.Rollout{{FeatureName: "feature-a", Version: "1.0.0", DeploymentID: "dep-1", Target: "tenant=tenant-a, environment=prod"}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected filtered rollouts\nwant: %#v\ngot:  %#v", want, got)
		}
	})

	t.Run("accepts target labels using env key", func(t *testing.T) {
		got := filterEnvironmentFeatureRollouts([]fasit.Rollout{{DeploymentID: "dep-1", Target: "tenant=tenant-a, env=prod"}}, "tenant-a", "prod")
		if len(got) != 1 {
			t.Fatalf("expected one matching rollout, got %#v", got)
		}
	})

	t.Run("returns nil when scope cannot be determined", func(t *testing.T) {
		got := filterEnvironmentFeatureRollouts(rollouts, "tenant-a", "missing")
		if len(got) != 0 {
			t.Fatalf("expected no matching rollout, got %#v", got)
		}
	})
}

func TestReadCommandFilterFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "tenants list filters",
			args: []string{"fasit", "tenant", "list", "--help"},
			want: []string{"--tenant", "--kind"},
		},
		{
			name: "features list filters",
			args: []string{"fasit", "feature", "list", "--help"},
			want: []string{"--feature", "--kind"},
		},
		{
			name: "feature status filters",
			args: []string{"fasit", "feature", "status", "my-feature", "--help"},
			want: []string{"--tenant", "--env", "--kind", "--enabled"},
		},
		{
			name: "rollouts list filters",
			args: []string{"fasit", "rollout", "list", "--help"},
			want: []string{"--feature", "--status"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := runFasitCommand(t, tt.args...)
			for _, want := range tt.want {
				if !strings.Contains(out, want) {
					t.Fatalf("help missing %q\n%s", want, out)
				}
			}
		})
	}
}

func TestDeploymentCommandRegistered(t *testing.T) {
	cmd := Fasit(&naistrix.GlobalFlags{})
	got := make([]string, 0, len(cmd.SubCommands))
	for _, sub := range cmd.SubCommands {
		got = append(got, sub.Name)
	}

	if !strings.Contains(strings.Join(got, ","), "deployment") {
		t.Fatalf("fasit root missing deployment command: %v", got)
	}
}

func TestRolloutRowsExposeDeploymentMetadata(t *testing.T) {
	rows := rolloutRows([]fasit.Rollout{{
		FeatureName:  "feature-a",
		Version:      "1.0.0",
		Status:       "DEPLOYED",
		Target:       "tenant=tenant-a, environment=prod",
		DeploymentID: "dep-1",
		Created:      "2026-01-01T10:00:00Z",
	}})

	if len(rows) != 1 {
		t.Fatalf("expected one row, got %d", len(rows))
	}
	if rows[0].Target != "tenant=tenant-a, environment=prod" {
		t.Fatalf("unexpected target %q", rows[0].Target)
	}
	if rows[0].DetailRef != "deployment:dep-1" {
		t.Fatalf("unexpected detail ref %q", rows[0].DetailRef)
	}
}

func TestRenderDeploymentDetail(t *testing.T) {
	t.Run("json preserves deployment payload", func(t *testing.T) {
		var buf bytes.Buffer
		out := naistrix.NewOutputWriter(&buf, nil)

		err := renderDeploymentDetail(out, "json", &fasit.Deployment{
			ID:          "dep-1",
			FeatureName: "feature-a",
			Version:     "1.0.0",
			Target:      "tenant=tenant-a, environment=prod",
			Statuses: []fasit.DeploymentStatus{{
				TenantName:      "tenant-a",
				EnvironmentName: "prod",
				State:           "DEPLOYED",
			}},
		})
		if err != nil {
			t.Fatalf("render deployment json: %v", err)
		}

		for _, want := range []string{`"id": "dep-1"`, `"featureName": "feature-a"`, `"state": "DEPLOYED"`} {
			if !strings.Contains(buf.String(), want) {
				t.Fatalf("deployment json missing %q\n%s", want, buf.String())
			}
		}
	})

	t.Run("table shows statuses section", func(t *testing.T) {
		var buf bytes.Buffer
		out := naistrix.NewOutputWriter(&buf, nil)

		err := renderDeploymentDetail(out, "", &fasit.Deployment{
			ID:          "dep-1",
			FeatureName: "feature-a",
			Version:     "1.0.0",
			Target:      "tenant=tenant-a, environment=prod",
			Statuses: []fasit.DeploymentStatus{{
				TenantName:      "tenant-a",
				EnvironmentName: "prod",
				State:           "DEPLOYED",
				Message:         "ok",
				LastModified:    "2026-01-01T10:05:00Z",
			}},
		})
		if err != nil {
			t.Fatalf("render deployment table: %v", err)
		}

		for _, want := range []string{"Environment statuses:", "tenant-a", "DEPLOYED"} {
			if !strings.Contains(buf.String(), want) {
				t.Fatalf("deployment table missing %q\n%s", want, buf.String())
			}
		}
	})
}

func TestFasitAutocompleteHelpers(t *testing.T) {
	t.Run("tenant names", func(t *testing.T) {
		got, hint := completeTenantNames(context.Background(), func(context.Context) ([]fasit.Tenant, error) {
			return []fasit.Tenant{{Name: "nav"}, {Name: "devnais"}}, nil
		})

		want := []string{"nav", "devnais"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected tenant completions\nwant: %#v\ngot:  %#v", want, got)
		}
		if hint != "Choose the tenant." {
			t.Fatalf("unexpected hint: %q", hint)
		}
	})

	t.Run("tenant names degrade gracefully on backend errors", func(t *testing.T) {
		got, hint := completeTenantNames(context.Background(), func(context.Context) ([]fasit.Tenant, error) {
			return nil, errors.New("iap expired")
		})

		if got != nil {
			t.Fatalf("expected nil completions, got %#v", got)
		}
		if !strings.Contains(hint, "Unable to list Fasit tenants for autocomplete") || !strings.Contains(hint, "iap expired") {
			t.Fatalf("unexpected hint: %q", hint)
		}
	})

	t.Run("environment names come from selected tenant", func(t *testing.T) {
		got, hint := completeEnvironmentNames(context.Background(), "nav", func(context.Context) ([]fasit.Tenant, error) {
			return []fasit.Tenant{{Name: "nav", Environments: []fasit.Environment{{Name: "dev"}, {Name: "prod"}}}}, nil
		})

		want := []string{"dev", "prod"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected env completions\nwant: %#v\ngot:  %#v", want, got)
		}
		if hint != "Choose an environment in nav." {
			t.Fatalf("unexpected hint: %q", hint)
		}
	})

	t.Run("environment names explain missing tenant", func(t *testing.T) {
		got, hint := completeEnvironmentNames(context.Background(), "missing", func(context.Context) ([]fasit.Tenant, error) {
			return []fasit.Tenant{{Name: "nav"}}, nil
		})

		if got != nil {
			t.Fatalf("expected nil completions, got %#v", got)
		}
		if hint != `Unknown tenant "missing".` {
			t.Fatalf("unexpected hint: %q", hint)
		}
	})

	t.Run("environment feature names come from selected environment", func(t *testing.T) {
		got, hint := completeEnvironmentFeatureNames(context.Background(), "nav", "prod", func(context.Context, string, string) (*fasit.Environment, error) {
			return &fasit.Environment{Features: []fasit.Feature{{Name: "a"}, {Name: "b"}}}, nil
		})

		want := []string{"a", "b"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected feature completions\nwant: %#v\ngot:  %#v", want, got)
		}
		if hint != "Choose a feature in nav/prod." {
			t.Fatalf("unexpected hint: %q", hint)
		}
	})

	t.Run("environment feature names explain missing environment", func(t *testing.T) {
		got, hint := completeEnvironmentFeatureNames(context.Background(), "nav", "missing", func(context.Context, string, string) (*fasit.Environment, error) {
			return nil, errors.New("not found: environment nav/missing")
		})

		if got != nil {
			t.Fatalf("expected nil completions, got %#v", got)
		}
		if hint != `Unknown environment "missing" in tenant "nav".` {
			t.Fatalf("unexpected hint: %q", hint)
		}
	})

	t.Run("feature names", func(t *testing.T) {
		got, hint := completeFeatureNames(context.Background(), func(context.Context) ([]fasit.Feature, error) {
			return []fasit.Feature{{Name: "a"}, {Name: "b"}}, nil
		})

		want := []string{"a", "b"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected feature completions\nwant: %#v\ngot:  %#v", want, got)
		}
		if hint != "Choose the feature." {
			t.Fatalf("unexpected hint: %q", hint)
		}
	})

	t.Run("rollout versions", func(t *testing.T) {
		got, hint := completeRolloutVersions(context.Background(), "feature-a", func(context.Context, string) ([]fasit.Rollout, error) {
			return []fasit.Rollout{{Version: "1.0.0"}, {Version: "1.0.1"}}, nil
		})

		want := []string{"1.0.0", "1.0.1"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected rollout completions\nwant: %#v\ngot:  %#v", want, got)
		}
		if hint != "Choose a rollout version for feature-a." {
			t.Fatalf("unexpected hint: %q", hint)
		}
	})

	t.Run("rollout versions explain unknown feature", func(t *testing.T) {
		got, hint := completeRolloutVersions(context.Background(), "missing", func(context.Context, string) ([]fasit.Rollout, error) {
			return nil, errors.New("not found: feature missing")
		})

		if got != nil {
			t.Fatalf("expected nil completions, got %#v", got)
		}
		if hint != `Unknown feature "missing".` {
			t.Fatalf("unexpected hint: %q", hint)
		}
	})
}

func TestTargetedCommandsHaveAutocomplete(t *testing.T) {
	parentFlags := &flag.Fasit{}
	commands := []*naistrix.Command{
		tenantGetCmd(parentFlags),
		envGetCmd(parentFlags),
		envFeatureGetCmd(parentFlags),
		envFeatureLogsCmd(parentFlags),
		envFeatureHelmCmd(parentFlags),
		envFeatureRolloutsCmd(parentFlags),
		envFeatureAuditCmd(parentFlags),
		featureGetCmd(parentFlags),
		featureStatusCmd(parentFlags),
		featureRolloutsCmd(parentFlags),
		rolloutGetCmd(parentFlags),
	}

	for _, cmd := range commands {
		if cmd.AutoCompleteFunc == nil {
			t.Fatalf("expected autocomplete func on %q", cmd.Name)
		}
	}
}

func TestResolveConfigMutationInputWithTerminal(t *testing.T) {
	t.Run("interactive secret path prints hidden prompt guidance", func(t *testing.T) {
		var buf bytes.Buffer
		out := naistrix.NewOutputWriter(&buf, nil)

		_, err := resolveConfigMutationInputWithTerminalPrompt(out, strings.NewReader("ignored\n"), 0, true, "", true, commandSecretPromptStub{value: []byte("secret\n")})
		if err != nil {
			t.Fatalf("resolve secret input: %v", err)
		}

		if !strings.Contains(buf.String(), "Enter new secret value (input hidden):") {
			t.Fatalf("expected hidden prompt guidance, got %q", buf.String())
		}
	})

	t.Run("interactive non-secret path stays silent", func(t *testing.T) {
		var buf bytes.Buffer
		out := naistrix.NewOutputWriter(&buf, nil)

		value, err := resolveConfigMutationInputWithTerminalPrompt(out, strings.NewReader("ignored\n"), 0, true, "plain", false, nil)
		if err != nil {
			t.Fatalf("resolve non-secret input: %v", err)
		}
		if value != "plain" {
			t.Fatalf("expected plain value, got %q", value)
		}
		if buf.Len() != 0 {
			t.Fatalf("expected no prompt output, got %q", buf.String())
		}
	})
}

type commandSecretPromptStub struct {
	value []byte
	err   error
}

func (s commandSecretPromptStub) ReadPassword(int) ([]byte, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.value, nil
}
