package gcp

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/oauth2"
	"google.golang.org/api/cloudresourcemanager/v3"
)

type Project struct {
	ID     string
	Tenant string
	Name   string
	Kind   Kind
}

func getProjects(ctx context.Context) ([]Project, error) {
	var projects []Project

	svc, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return nil, err
	}

	filter := "labels.naiscluster=true OR labels.kind=knada"

	call := svc.Projects.Search().Query(filter)
	for {
		response, err := call.Do()
		if err != nil {
			var retrieve *oauth2.RetrieveError
			if errors.As(err, &retrieve) {
				if retrieve.ErrorCode == "invalid_grant" {
					return nil, fmt.Errorf("looks like you are missing Application Default Credentials, run `gcloud auth application-default login` first")
				}
			}

			return nil, err
		}

		for _, project := range response.Projects {
			if project.State != "ACTIVE" {
				// Only check active projects. When a project is deleted,
				// it is marked as DELETING for a while before it is removed.
				// This results in a 403 when trying to list clusters.
				continue
			}

			projects = append(projects, Project{
				ID:     project.ProjectId,
				Tenant: project.Labels["tenant"],
				Name:   project.Labels["environment"],
				Kind:   ParseKind(project.Labels["kind"]),
			})
		}
		if response.NextPageToken == "" {
			break
		}
		call.PageToken(response.NextPageToken)
	}

	return projects, nil
}
