package command

import (
	"context"

	"github.com/nais/naistrix"
	"github.com/nais/narcos/internal/fasit"
	"github.com/nais/narcos/internal/fasit/command/flag"
)

func deploymentCmd(parentFlags *flag.Fasit) *naistrix.Command {
	return &naistrix.Command{
		Name:  "deployment",
		Title: "Inspect a single deployment-backed rollout.",
		SubCommands: []*naistrix.Command{
			deploymentGetCmd(parentFlags),
		},
	}
}

func deploymentGetCmd(parentFlags *flag.Fasit) *naistrix.Command {
	flags := &flag.DeploymentGet{Fasit: parentFlags}
	return &naistrix.Command{
		Name:  "get",
		Title: "Get deployment-backed rollout detail.",
		Flags: flags,
		Args:  []naistrix.Argument{{Name: "id"}},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			detail, err := fasit.GetDeployment(ctx, args.Get("id"))
			if err != nil {
				return err
			}

			return renderDeploymentDetail(out, flags.Output, detail)
		},
	}
}

type deploymentRow struct {
	ID          string `heading:"Deployment ID" json:"id" yaml:"id"`
	FeatureName string `heading:"Feature" json:"featureName" yaml:"featureName"`
	Version     string `heading:"Version" json:"version" yaml:"version"`
	Target      string `heading:"Target" json:"target" yaml:"target"`
	Created     string `heading:"Created" json:"created" yaml:"created"`
	Description string `heading:"Description" json:"description,omitempty" yaml:"description,omitempty"`
}

type deploymentStatusRow struct {
	Tenant       string `heading:"Tenant" json:"tenant" yaml:"tenant"`
	Environment  string `heading:"Environment" json:"environment" yaml:"environment"`
	State        string `heading:"State" json:"state" yaml:"state"`
	Message      string `heading:"Message" json:"message" yaml:"message"`
	LastModified string `heading:"Last modified" json:"lastModified" yaml:"lastModified"`
}

func renderDeploymentDetail(out *naistrix.OutputWriter, format string, detail *fasit.Deployment) error {
	switch fasit.NormalizeOutputFormat(format) {
	case fasit.OutputFormatJSON, fasit.OutputFormatYAML:
		return fasit.RenderDataOutput(out, format, detail)
	default:
		if err := out.Table().Render([]deploymentRow{{
			ID:          detail.ID,
			FeatureName: detail.FeatureName,
			Version:     detail.Version,
			Target:      detail.Target,
			Created:     detail.Created,
			Description: detail.Description,
		}}); err != nil {
			return err
		}

		out.Println("")
		out.Println("Environment statuses:")
		if len(detail.Statuses) == 0 {
			out.Println("No results")
			return nil
		}

		return out.Table().Render(deploymentStatusRows(detail.Statuses))
	}
}

func deploymentStatusRows(statuses []fasit.DeploymentStatus) []deploymentStatusRow {
	rows := make([]deploymentStatusRow, 0, len(statuses))
	for _, status := range statuses {
		rows = append(rows, deploymentStatusRow{
			Tenant:       status.TenantName,
			Environment:  status.EnvironmentName,
			State:        status.State,
			Message:      status.Message,
			LastModified: status.LastModified,
		})
	}

	return rows
}
