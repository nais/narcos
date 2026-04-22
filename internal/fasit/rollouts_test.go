package fasit

import (
	"context"
	"encoding/json"
	"testing"

	fasitgraphql "github.com/nais/narcos/internal/fasit/graphql"
	"github.com/stretchr/testify/require"
)

func TestListRollouts(t *testing.T) {
	client := &mockGraphQLClient{responses: map[string]any{
		"GetRollouts": &fasitgraphql.GetRolloutsResponse{Rollouts: []*fasitgraphql.GetRolloutsRolloutsRollout{{
			FeatureName: "my-feature",
			Version:     "1.2.3",
			Status:      fasitgraphql.RolloutStatusPending,
			Created:     "2026-01-01T10:00:00Z",
			Completed:   new("2026-01-01T11:00:00Z"),
			Events: []*fasitgraphql.GetRolloutsRolloutsRolloutEventsRolloutEvent{{
				Message: "started",
				Failure: false,
			}},
		}}},
	}}

	rollouts, err := (&store{client: client}).listRollouts(context.Background(), "my-feature")
	require.NoError(t, err)
	require.Len(t, rollouts, 1)
	require.Equal(t, "my-feature", rollouts[0].FeatureName)
	require.Equal(t, "1.2.3", rollouts[0].Version)
	require.Equal(t, "PENDING", rollouts[0].Status)
	require.Equal(t, "started", rollouts[0].Events[0].Message)
}

func TestListFeatureRolloutsIncludesDeployments(t *testing.T) {
	client := &mockGraphQLClient{responses: map[string]any{
		"GetRollouts": &fasitgraphql.GetRolloutsResponse{Rollouts: []*fasitgraphql.GetRolloutsRolloutsRollout{{
			FeatureName: "my-feature",
			Version:     "1.2.2",
			Status:      fasitgraphql.RolloutStatusDeployed,
			Created:     "2026-01-01T10:00:00Z",
		}}},
		"GetDeployments": &fasitgraphql.GetDeploymentsResponse{Deployments: []*fasitgraphql.GetDeploymentsDeploymentsDeployment{{
			Id:      "deployment-1",
			Created: "2026-01-02T10:00:00Z",
			Ci:      false,
			Feature: &fasitgraphql.GetDeploymentsDeploymentsDeploymentFeature{Name: "my-feature", Version: "1.2.3"},
			Target: []*fasitgraphql.GetDeploymentsDeploymentsDeploymentTargetEnvironmentLabel{{
				Key:   "tenant",
				Value: "tenant-a",
			}, {
				Key:   "environment",
				Value: "prod",
			}},
			Statuses: []*fasitgraphql.GetDeploymentsDeploymentsDeploymentStatusesDeploymentStatus{{
				State: fasitgraphql.DeploymentStatusStateDeployed,
			}},
		}}},
	}}

	history, err := (&store{client: client}).listFeatureRollouts(context.Background(), "my-feature")
	require.NoError(t, err)
	require.Len(t, history, 2)
	require.Equal(t, "deployment-1", history[0].DeploymentID)
	require.Equal(t, "tenant=tenant-a, environment=prod", history[0].Target)
	require.Equal(t, "DEPLOYED", history[0].Status)
	require.Equal(t, "1.2.2", history[1].Version)
}

func TestGetRollout(t *testing.T) {
	client := &mockGraphQLClient{responses: map[string]any{
		"GetRollout": &fasitgraphql.GetRolloutResponse{Rollout: &fasitgraphql.GetRolloutRollout{
			Id:          "rollout-1",
			FeatureName: "my-feature",
			Version:     "1.2.3",
			Status:      fasitgraphql.RolloutStatusDeployed,
			Created:     "2026-01-01T10:00:00Z",
			Completed:   new("2026-01-01T11:00:00Z"),
			Events: []*fasitgraphql.GetRolloutRolloutEventsRolloutEvent{{
				Message: "done",
				Failure: false,
				Created: "2026-01-01T10:30:00Z",
			}},
			Logs: []*fasitgraphql.GetRolloutRolloutLogsRolloutLog{{
				TenantName:  "tenant-a",
				Environment: "dev",
				Lines: []*fasitgraphql.GetRolloutRolloutLogsRolloutLogLinesLogLine{{
					Timestamp: "2026-01-01T10:35:00Z",
					Message:   "healthy",
				}},
			}},
		}},
	}}

	detail, err := (&store{client: client}).getRollout(context.Background(), "my-feature", "1.2.3")
	require.NoError(t, err)
	require.Equal(t, "rollout-1", detail.ID)
	require.Equal(t, "DEPLOYED", detail.Status)
	require.Len(t, detail.Events, 1)
	require.Len(t, detail.Logs, 1)
	require.Equal(t, "healthy", detail.Logs[0].Lines[0].Message)
}

func TestListAllRollouts(t *testing.T) {
	client := &mockGraphQLClient{responses: map[string]any{
		"GetRolloutsList": &fasitgraphql.GetRolloutsListResponse{Rollouts: []*fasitgraphql.GetRolloutsListRolloutsRollout{{
			FeatureName: "my-feature",
			Version:     "1.2.3",
			Status:      fasitgraphql.RolloutStatusCreated,
			Created:     "2026-01-01T10:00:00Z",
		}}},
		"GetDeploymentsList": &fasitgraphql.GetDeploymentsListResponse{Deployments: []*fasitgraphql.GetDeploymentsListDeploymentsDeployment{{
			Id:      "deployment-1",
			Created: "2026-01-02T10:00:00Z",
			Ci:      true,
			Feature: &fasitgraphql.GetDeploymentsListDeploymentsDeploymentFeature{Name: "my-feature", Version: "1.2.4"},
			Statuses: []*fasitgraphql.GetDeploymentsListDeploymentsDeploymentStatusesDeploymentStatus{{
				State: fasitgraphql.DeploymentStatusStatePending,
			}},
		}}},
	}}

	rollouts, err := (&store{client: client}).listAllRollouts(context.Background())
	require.NoError(t, err)
	require.Len(t, rollouts, 2)
	require.Equal(t, "deployment-1", rollouts[0].DeploymentID)
	require.Equal(t, "CI", rollouts[0].Target)
	require.Equal(t, "PENDING", rollouts[0].Status)
	require.Equal(t, "CREATED", rollouts[1].Status)
}

func TestGetDeployment(t *testing.T) {
	client := &mockGraphQLClient{responses: map[string]any{
		"GetDeploymentDetail": &fasitgraphql.GetDeploymentDetailResponse{Deployment: &fasitgraphql.GetDeploymentDetailDeployment{
			Id:          "deployment-1",
			Created:     "2026-01-02T10:00:00Z",
			Ci:          false,
			Description: new("manual deploy"),
			Feature:     &fasitgraphql.GetDeploymentDetailDeploymentFeature{Name: "my-feature", Version: "1.2.3"},
			Target: []*fasitgraphql.GetDeploymentDetailDeploymentTargetEnvironmentLabel{{
				Key:   "tenant",
				Value: "tenant-a",
			}, {
				Key:   "environment",
				Value: "prod",
			}},
			Statuses: []*fasitgraphql.GetDeploymentDetailDeploymentStatusesDeploymentStatus{{
				State:        fasitgraphql.DeploymentStatusStateFailed,
				Message:      "boom",
				LastModified: "2026-01-02T10:05:00Z",
				Environment: &fasitgraphql.GetDeploymentDetailDeploymentStatusesDeploymentStatusEnvironment{
					Name: "prod",
					Tenant: &fasitgraphql.GetDeploymentDetailDeploymentStatusesDeploymentStatusEnvironmentTenant{
						Name: "tenant-a",
					},
				},
			}},
		}},
	}}

	deployment, err := (&store{client: client}).getDeployment(context.Background(), "deployment-1")
	require.NoError(t, err)
	require.Equal(t, "deployment-1", deployment.ID)
	require.Equal(t, "my-feature", deployment.FeatureName)
	require.Equal(t, "tenant=tenant-a, environment=prod", deployment.Target)
	require.Equal(t, "manual deploy", deployment.Description)
	require.Len(t, deployment.Statuses, 1)
	require.Equal(t, "FAILED", deployment.Statuses[0].State)
	require.Equal(t, "tenant-a", deployment.Statuses[0].TenantName)
}

func TestDeploymentHelpers(t *testing.T) {
	t.Run("derive status ignores nil entries", func(t *testing.T) {
		statuses := []*fasitgraphql.GetDeploymentsDeploymentsDeploymentStatusesDeploymentStatus{
			nil,
			{State: fasitgraphql.DeploymentStatusStatePending},
		}
		require.Equal(t, "PENDING", deriveDeploymentStatus(statuses))
	})

	t.Run("format target ignores nil entries and falls back when empty", func(t *testing.T) {
		targets := []*fasitgraphql.GetDeploymentsDeploymentsDeploymentTargetEnvironmentLabel{
			nil,
			{Key: "tenant", Value: "tenant-a"},
		}
		require.Equal(t, "tenant=tenant-a", formatDeploymentTarget(false, targets))
		require.Equal(t, "All environments", formatDeploymentTarget(false, []*fasitgraphql.GetDeploymentsDeploymentsDeploymentTargetEnvironmentLabel{nil}))
	})
}

func TestGetFeatureLog(t *testing.T) {
	client := &mockGraphQLClient{responses: map[string]any{
		"GetFeatureLog": &fasitgraphql.GetFeatureLogResponse{Tenant: &fasitgraphql.GetFeatureLogTenant{
			Environment: &fasitgraphql.GetFeatureLogTenantEnvironment{
				Feature: &fasitgraphql.GetFeatureLogTenantEnvironmentFeature{
					Name: "my-feature",
					Status: &fasitgraphql.GetFeatureLogTenantEnvironmentFeatureStatus{
						Version:      "1.2.3",
						Status:       fasitgraphql.RolloutStatusPending,
						LastModified: "2026-01-01T10:00:00Z",
						Log: []*fasitgraphql.GetFeatureLogTenantEnvironmentFeatureStatusLogLogLine{{
							Timestamp: "2026-01-01T10:01:00Z",
							Message:   "syncing",
						}},
					},
					HelmValueDiff: &fasitgraphql.GetFeatureLogTenantEnvironmentFeatureHelmValueDiff{
						Difference: fasitgraphql.HelmValueDifferenceNoMatch,
						Diff:       "- old\n+ new",
					},
				},
			},
		}},
	}}

	featureLog, err := (&store{client: client}).getFeatureLog(context.Background(), "tenant-a", "dev", "my-feature")
	require.NoError(t, err)
	require.Equal(t, "1.2.3", featureLog.CurrentVersion)
	require.Equal(t, "syncing", featureLog.CurrentLog[0].Message)
	require.Equal(t, "NO_MATCH", featureLog.HelmDiff.Difference)
}

func TestGetHelmValues(t *testing.T) {
	client := &mockGraphQLClient{responses: map[string]any{
		"GetHelmValues": &fasitgraphql.GetHelmValuesResponse{HelmValues: json.RawMessage(`"key: value\n"`)},
	}}

	values, err := (&store{client: client}).getHelmValues(context.Background(), "my-feature", "tenant-a", "dev")
	require.NoError(t, err)
	require.Equal(t, "key: value\n", values)
}
