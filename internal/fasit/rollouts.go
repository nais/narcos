package fasit

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	fasitgraphql "github.com/nais/narcos/internal/fasit/graphql"
)

func ListRollouts(ctx context.Context, feature string) ([]Rollout, error) {
	store, err := newStore(ctx)
	if err != nil {
		return nil, err
	}

	return store.listRollouts(ctx, feature)
}

func ListFeatureRollouts(ctx context.Context, feature string) ([]Rollout, error) {
	store, err := newStore(ctx)
	if err != nil {
		return nil, err
	}

	return store.listFeatureRollouts(ctx, feature)
}

func GetRollout(ctx context.Context, feature, version string) (*RolloutDetail, error) {
	store, err := newStore(ctx)
	if err != nil {
		return nil, err
	}

	return store.getRollout(ctx, feature, version)
}

func ListAllRollouts(ctx context.Context) ([]RolloutSummary, error) {
	store, err := newStore(ctx)
	if err != nil {
		return nil, err
	}

	return store.listAllRollouts(ctx)
}

func GetDeployment(ctx context.Context, id string) (*Deployment, error) {
	store, err := newStore(ctx)
	if err != nil {
		return nil, err
	}

	return store.getDeployment(ctx, id)
}

func GetFeatureLog(ctx context.Context, tenant, env, feature string) (*FeatureLog, error) {
	store, err := newStore(ctx)
	if err != nil {
		return nil, err
	}

	return store.getFeatureLog(ctx, tenant, env, feature)
}

func GetHelmValues(ctx context.Context, feature, tenant, env string) (string, error) {
	store, err := newStore(ctx)
	if err != nil {
		return "", err
	}

	return store.getHelmValues(ctx, feature, tenant, env)
}

func (s *store) listRollouts(ctx context.Context, feature string) ([]Rollout, error) {
	resp, err := fasitgraphql.GetRollouts(ctx, s.client, feature)
	if err != nil {
		return nil, fmt.Errorf("get rollouts: %w", err)
	}

	rollouts := make([]Rollout, 0, len(resp.Rollouts))
	for _, rollout := range resp.Rollouts {
		if rollout == nil {
			continue
		}

		rollouts = append(rollouts, Rollout{
			FeatureName: rollout.FeatureName,
			Version:     rollout.Version,
			Status:      string(rollout.Status),
			Created:     rollout.Created,
			Completed:   stringValue(rollout.Completed),
			Events:      convertRolloutEvents(rollout.Events),
		})
	}

	return rollouts, nil
}

func (s *store) listFeatureRollouts(ctx context.Context, feature string) ([]Rollout, error) {
	rollouts, err := s.listRollouts(ctx, feature)
	if err != nil {
		return nil, err
	}

	deployments, err := s.listDeployments(ctx, feature)
	if err != nil {
		return nil, err
	}

	history := append(rollouts, deployments...)
	sort.Slice(history, func(i, j int) bool {
		return history[i].Created > history[j].Created
	})

	return history, nil
}

func (s *store) getRollout(ctx context.Context, feature, version string) (*RolloutDetail, error) {
	resp, err := fasitgraphql.GetRollout(ctx, s.client, feature, version)
	if err != nil {
		return nil, fmt.Errorf("get rollout: %w", err)
	}

	if resp.Rollout == nil {
		return nil, fmt.Errorf("not found: rollout %s/%s", feature, version)
	}

	return &RolloutDetail{
		ID:          resp.Rollout.Id,
		FeatureName: resp.Rollout.FeatureName,
		Version:     resp.Rollout.Version,
		Status:      string(resp.Rollout.Status),
		Created:     resp.Rollout.Created,
		Completed:   stringValue(resp.Rollout.Completed),
		Events:      convertRolloutDetailEvents(resp.Rollout.Events),
		Logs:        convertRolloutLogs(resp.Rollout.Logs),
	}, nil
}

func (s *store) listAllRollouts(ctx context.Context) ([]RolloutSummary, error) {
	resp, err := fasitgraphql.GetRolloutsList(ctx, s.client)
	if err != nil {
		return nil, fmt.Errorf("get rollouts list: %w", err)
	}

	rollouts := make([]RolloutSummary, 0, len(resp.Rollouts))
	for _, rollout := range resp.Rollouts {
		if rollout == nil {
			continue
		}

		rollouts = append(rollouts, RolloutSummary{
			FeatureName: rollout.FeatureName,
			Version:     rollout.Version,
			Status:      string(rollout.Status),
			Created:     rollout.Created,
			Completed:   stringValue(rollout.Completed),
		})
	}

	deployments, err := s.listDeploymentsList(ctx)
	if err != nil {
		return nil, err
	}

	rollouts = append(rollouts, deployments...)
	sort.Slice(rollouts, func(i, j int) bool {
		return rollouts[i].Created > rollouts[j].Created
	})

	return rollouts, nil
}

func (s *store) listDeployments(ctx context.Context, feature string) ([]Rollout, error) {
	resp, err := fasitgraphql.GetDeployments(ctx, s.client, feature)
	if err != nil {
		return nil, fmt.Errorf("get deployments: %w", err)
	}

	deployments := make([]Rollout, 0, len(resp.Deployments))
	for _, deployment := range resp.Deployments {
		if deployment == nil || deployment.Feature == nil {
			continue
		}

		deployments = append(deployments, Rollout{
			FeatureName:  deployment.Feature.Name,
			Version:      deployment.Feature.Version,
			Status:       deriveDeploymentStatus(deployment.Statuses),
			Created:      deployment.Created,
			Target:       formatDeploymentTarget(deployment.Ci, deployment.Target),
			DeploymentID: deployment.Id,
		})
	}

	return deployments, nil
}

func (s *store) listDeploymentsList(ctx context.Context) ([]RolloutSummary, error) {
	resp, err := fasitgraphql.GetDeploymentsList(ctx, s.client)
	if err != nil {
		return nil, fmt.Errorf("get deployments list: %w", err)
	}

	deployments := make([]RolloutSummary, 0, len(resp.Deployments))
	for _, deployment := range resp.Deployments {
		if deployment == nil || deployment.Feature == nil {
			continue
		}

		deployments = append(deployments, RolloutSummary{
			FeatureName:  deployment.Feature.Name,
			Version:      deployment.Feature.Version,
			Status:       deriveDeploymentStatus(deployment.Statuses),
			Created:      deployment.Created,
			Target:       formatDeploymentTarget(deployment.Ci, deployment.Target),
			DeploymentID: deployment.Id,
		})
	}

	return deployments, nil
}

func (s *store) getDeployment(ctx context.Context, id string) (*Deployment, error) {
	resp, err := fasitgraphql.GetDeploymentDetail(ctx, s.client, id)
	if err != nil {
		return nil, fmt.Errorf("get deployment: %w", err)
	}

	if resp.Deployment == nil || resp.Deployment.Feature == nil {
		return nil, fmt.Errorf("deployment not found: %s", id)
	}

	statuses := make([]DeploymentStatus, 0, len(resp.Deployment.Statuses))
	for _, status := range resp.Deployment.Statuses {
		if status == nil {
			continue
		}

		deploymentStatus := DeploymentStatus{
			State:        string(status.State),
			Message:      status.Message,
			LastModified: status.LastModified,
		}
		if status.Environment != nil {
			deploymentStatus.EnvironmentName = status.Environment.Name
			if status.Environment.Tenant != nil {
				deploymentStatus.TenantName = status.Environment.Tenant.Name
			}
		}

		statuses = append(statuses, deploymentStatus)
	}

	return &Deployment{
		ID:          resp.Deployment.Id,
		FeatureName: resp.Deployment.Feature.Name,
		Version:     resp.Deployment.Feature.Version,
		Description: stringValue(resp.Deployment.Description),
		Created:     resp.Deployment.Created,
		Target:      formatDeploymentTarget(resp.Deployment.Ci, resp.Deployment.Target),
		Statuses:    statuses,
	}, nil
}

func (s *store) getFeatureLog(ctx context.Context, tenant, env, feature string) (*FeatureLog, error) {
	resp, err := fasitgraphql.GetFeatureLog(ctx, s.client, tenant, env, feature)
	if err != nil {
		return nil, fmt.Errorf("get feature log: %w", err)
	}

	if resp.Tenant == nil || resp.Tenant.Environment == nil || resp.Tenant.Environment.Feature == nil {
		return nil, fmt.Errorf("not found: feature %s in environment %s/%s", feature, tenant, env)
	}

	featureResp := resp.Tenant.Environment.Feature
	log := &FeatureLog{}
	if featureResp.Status != nil {
		log.CurrentVersion = featureResp.Status.Version
		log.CurrentStatus = string(featureResp.Status.Status)
		log.LastModified = featureResp.Status.LastModified
		log.CurrentLog = convertFeatureLogLines(featureResp.Status.Log)
	}

	if featureResp.HelmValueDiff != nil {
		log.HelmDiff = HelmValueDiff{
			Difference: string(featureResp.HelmValueDiff.Difference),
			Diff:       featureResp.HelmValueDiff.Diff,
		}
	}

	return log, nil
}

func (s *store) getHelmValues(ctx context.Context, feature, tenant, env string) (string, error) {
	resp, err := fasitgraphql.GetHelmValues(ctx, s.client, feature, nil, &tenant, &env)
	if err != nil {
		return "", fmt.Errorf("get helm values: %w", err)
	}

	return decodeHelmValues(resp.HelmValues), nil
}

func convertRolloutEvents(events []*fasitgraphql.GetRolloutsRolloutsRolloutEventsRolloutEvent) []RolloutEvent {
	out := make([]RolloutEvent, 0, len(events))
	for _, event := range events {
		if event == nil {
			continue
		}

		out = append(out, RolloutEvent{Message: event.Message, Failure: event.Failure})
	}

	return out
}

func convertRolloutDetailEvents(events []*fasitgraphql.GetRolloutRolloutEventsRolloutEvent) []RolloutEvent {
	out := make([]RolloutEvent, 0, len(events))
	for _, event := range events {
		if event == nil {
			continue
		}

		out = append(out, RolloutEvent{
			Message: event.Message,
			Failure: event.Failure,
			Created: event.Created,
		})
	}

	return out
}

func convertRolloutLogs(logs []*fasitgraphql.GetRolloutRolloutLogsRolloutLog) []RolloutLog {
	out := make([]RolloutLog, 0, len(logs))
	for _, log := range logs {
		if log == nil {
			continue
		}

		out = append(out, RolloutLog{
			TenantName:  log.TenantName,
			Environment: log.Environment,
			Lines:       convertRolloutLogLines(log.Lines),
		})
	}

	return out
}

func convertRolloutLogLines(lines []*fasitgraphql.GetRolloutRolloutLogsRolloutLogLinesLogLine) []LogLine {
	out := make([]LogLine, 0, len(lines))
	for _, line := range lines {
		if line == nil {
			continue
		}

		out = append(out, LogLine{Timestamp: line.Timestamp, Message: line.Message})
	}

	return out
}

func convertFeatureLogLines(lines []*fasitgraphql.GetFeatureLogTenantEnvironmentFeatureStatusLogLogLine) []LogLine {
	out := make([]LogLine, 0, len(lines))
	for _, line := range lines {
		if line == nil {
			continue
		}

		out = append(out, LogLine{Timestamp: line.Timestamp, Message: line.Message})
	}

	return out
}

func decodeHelmValues(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		return asString
	}

	return string(raw)
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

func deriveDeploymentStatus[T interface {
	comparable
	GetState() fasitgraphql.DeploymentStatusState
}](statuses []T) string {
	if len(statuses) == 0 {
		return "UNKNOWN"
	}

	var zero T
	for _, status := range statuses {
		if status == zero {
			continue
		}
		if status.GetState() == fasitgraphql.DeploymentStatusStateFailed {
			return "FAILED"
		}
	}

	for _, status := range statuses {
		if status == zero {
			continue
		}
		if status.GetState() != fasitgraphql.DeploymentStatusStateDeployed {
			return "PENDING"
		}
	}

	return "DEPLOYED"
}

type deploymentTargetLabel interface {
	comparable
	GetKey() string
	GetValue() string
}

func formatDeploymentTarget[T deploymentTargetLabel](ci bool, targets []T) string {
	if ci {
		return "CI"
	}
	if len(targets) == 0 {
		return "All environments"
	}

	var zero T
	parts := make([]string, 0, len(targets))
	for _, target := range targets {
		if target == zero {
			continue
		}
		parts = append(parts, target.GetKey()+"="+target.GetValue())
	}
	if len(parts) == 0 {
		return "All environments"
	}

	return strings.Join(parts, ", ")
}
