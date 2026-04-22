package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/naistrix"
	"github.com/nais/narcos/internal/fasit"
	"github.com/nais/narcos/internal/fasit/command/flag"
)

func rolloutsCmd(parentFlags *flag.Fasit) *naistrix.Command {
	return &naistrix.Command{
		Name:  "rollouts",
		Title: "Inspect rollout history.",
		SubCommands: []*naistrix.Command{
			rolloutsListCmd(parentFlags),
		},
	}
}

func rolloutCmd(parentFlags *flag.Fasit) *naistrix.Command {
	return &naistrix.Command{
		Name:  "rollout",
		Title: "Inspect a single rollout.",
		SubCommands: []*naistrix.Command{
			rolloutGetCmd(parentFlags),
		},
	}
}

func rolloutsListCmd(parentFlags *flag.Fasit) *naistrix.Command {
	flags := &flag.RolloutsList{Fasit: parentFlags}
	return &naistrix.Command{
		Name:  "list",
		Title: "List all rollouts.",
		Flags: flags,
		RunFunc: func(ctx context.Context, _ *naistrix.Arguments, out *naistrix.OutputWriter) error {
			rollouts, err := fasit.ListAllRollouts(ctx)
			if err != nil {
				return err
			}

			rollouts = filterRolloutSummaries(rollouts, flags.Feature, flags.Status)

			return fasit.RenderStructuredOutput(out, flags.Output, rolloutSummaryRows(rollouts), rollouts)
		},
	}
}

func rolloutGetCmd(parentFlags *flag.Fasit) *naistrix.Command {
	flags := &flag.RolloutGet{Fasit: parentFlags}
	return &naistrix.Command{
		Name:             "get",
		Title:            "Get rollout detail.",
		Flags:            flags,
		AutoCompleteFunc: completeFasitRolloutArgs(flags.Fasit),
		Args:             []naistrix.Argument{{Name: "feature"}, {Name: "version"}},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			detail, err := fasit.GetRollout(ctx, args.Get("feature"), args.Get("version"))
			if err != nil {
				return err
			}

			return renderRolloutDetail(out, flags.Output, detail)
		},
	}
}

func filterRolloutSummaries(rollouts []fasit.RolloutSummary, featureName, status string) []fasit.RolloutSummary {
	filtered := make([]fasit.RolloutSummary, 0, len(rollouts))
	for _, rollout := range rollouts {
		if featureName != "" && rollout.FeatureName != featureName {
			continue
		}
		if status != "" && !strings.EqualFold(rollout.Status, status) {
			continue
		}

		filtered = append(filtered, rollout)
	}

	return filtered
}

type rolloutRow struct {
	Feature   string `heading:"Feature" json:"feature" yaml:"feature"`
	Version   string `heading:"Version" json:"version" yaml:"version"`
	Status    string `heading:"Status" json:"status" yaml:"status"`
	Target    string `heading:"Target" json:"target,omitempty" yaml:"target,omitempty"`
	DetailRef string `heading:"Detail" json:"detailRef,omitempty" yaml:"detailRef,omitempty"`
	Created   string `heading:"Created" json:"created" yaml:"created"`
	Completed string `heading:"Completed" json:"completed" yaml:"completed"`
}

type rolloutEventRow struct {
	Created string `heading:"Created" json:"created" yaml:"created"`
	Failure bool   `heading:"Failure" json:"failure" yaml:"failure"`
	Message string `heading:"Message" json:"message" yaml:"message"`
}

type logLineRow struct {
	Timestamp string `heading:"Timestamp" json:"timestamp" yaml:"timestamp"`
	Message   string `heading:"Message" json:"message" yaml:"message"`
}

func rolloutRows(rollouts []fasit.Rollout) []rolloutRow {
	rows := make([]rolloutRow, 0, len(rollouts))
	for _, rollout := range rollouts {
		rows = append(rows, rolloutRow{
			Feature:   rollout.FeatureName,
			Version:   rollout.Version,
			Status:    rollout.Status,
			Target:    rollout.Target,
			DetailRef: rolloutDetailRef(rollout.DeploymentID),
			Created:   rollout.Created,
			Completed: rollout.Completed,
		})
	}

	return rows
}

func rolloutSummaryRows(rollouts []fasit.RolloutSummary) []rolloutRow {
	rows := make([]rolloutRow, 0, len(rollouts))
	for _, rollout := range rollouts {
		rows = append(rows, rolloutRow{
			Feature:   rollout.FeatureName,
			Version:   rollout.Version,
			Status:    rollout.Status,
			Target:    rollout.Target,
			DetailRef: rolloutDetailRef(rollout.DeploymentID),
			Created:   rollout.Created,
			Completed: rollout.Completed,
		})
	}

	return rows
}

func renderFeatureLog(out *naistrix.OutputWriter, format string, featureLog *fasit.FeatureLog) error {
	switch fasit.NormalizeOutputFormat(format) {
	case fasit.OutputFormatJSON, fasit.OutputFormatYAML:
		return fasit.RenderDataOutput(out, format, featureLog)
	default:
		summary := []struct {
			Version      string `heading:"Version"`
			Status       string `heading:"Status"`
			LastModified string `heading:"Last modified"`
		}{{
			Version:      featureLog.CurrentVersion,
			Status:       featureLog.CurrentStatus,
			LastModified: featureLog.LastModified,
		}}

		if err := out.Table().Render(summary); err != nil {
			return err
		}

		if len(featureLog.CurrentLog) > 0 {
			out.Println("")
			if err := out.Table().Render(logLineRows(featureLog.CurrentLog)); err != nil {
				return err
			}
		}

		if featureLog.HelmDiff.Diff != "" {
			out.Println("")
			out.Println("Helm diff:")
			out.Println(featureLog.HelmDiff.Diff)
		}

		return nil
	}
}

func renderHelmValues(out *naistrix.OutputWriter, format, values string) error {
	switch fasit.NormalizeOutputFormat(format) {
	case fasit.OutputFormatJSON, fasit.OutputFormatYAML:
		return fasit.RenderDataOutput(out, format, map[string]string{"helmValues": values})
	default:
		out.Println(values)
		return nil
	}
}

func renderRolloutDetail(out *naistrix.OutputWriter, format string, detail *fasit.RolloutDetail) error {
	switch fasit.NormalizeOutputFormat(format) {
	case fasit.OutputFormatJSON, fasit.OutputFormatYAML:
		return fasit.RenderDataOutput(out, format, detail)
	default:
		if err := out.Table().Render([]rolloutRow{{
			Feature:   detail.FeatureName,
			Version:   detail.Version,
			Status:    detail.Status,
			Target:    "",
			DetailRef: "",
			Created:   detail.Created,
			Completed: detail.Completed,
		}}); err != nil {
			return err
		}

		out.Println("")
		out.Println("Events:")
		if err := out.Table().Render(rolloutEventRows(detail.Events)); err != nil {
			return err
		}

		for _, log := range detail.Logs {
			out.Println("")
			out.Println(fmt.Sprintf("Logs for %s/%s:", log.TenantName, log.Environment))
			if err := out.Table().Render(logLineRows(log.Lines)); err != nil {
				return err
			}
		}

		return nil
	}
}

func rolloutEventRows(events []fasit.RolloutEvent) []rolloutEventRow {
	rows := make([]rolloutEventRow, 0, len(events))
	for _, event := range events {
		rows = append(rows, rolloutEventRow{Created: event.Created, Failure: event.Failure, Message: event.Message})
	}

	return rows
}

func logLineRows(lines []fasit.LogLine) []logLineRow {
	rows := make([]logLineRow, 0, len(lines))
	for _, line := range lines {
		rows = append(rows, logLineRow{Timestamp: line.Timestamp, Message: line.Message})
	}

	return rows
}

func rolloutDetailRef(deploymentID string) string {
	if deploymentID == "" {
		return "rollout"
	}

	return "deployment:" + deploymentID
}

func filterEnvironmentFeatureRollouts(rollouts []fasit.Rollout, tenant, env string) []fasit.Rollout {
	if tenant == "" || env == "" {
		return nil
	}

	filtered := make([]fasit.Rollout, 0, len(rollouts))
	for _, rollout := range rollouts {
		if rollout.DeploymentID == "" {
			continue
		}

		labels := map[string]string{}
		for target := range strings.SplitSeq(rollout.Target, ",") {
			parts := strings.SplitN(strings.TrimSpace(target), "=", 2)
			if len(parts) != 2 {
				continue
			}
			labels[parts[0]] = parts[1]
		}

		tenantLabel, tenantOK := labels["tenant"]
		envLabel, envOK := labels["environment"]
		if !envOK {
			envLabel, envOK = labels["env"]
		}
		if tenantOK && envOK && tenantLabel == tenant && envLabel == env {
			filtered = append(filtered, rollout)
			continue
		}
		if combined, ok := labels[tenant]; ok && combined == env {
			filtered = append(filtered, rollout)
			continue
		}
		if len(labels) == 0 && strings.TrimSpace(rollout.Target) == tenant+"="+env {
			filtered = append(filtered, rollout)
			continue
		}
	}

	return filtered
}
