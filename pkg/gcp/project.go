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

	filter := "labels.naiscluster=true OR labels.kind=legacy OR labels.kind=onprem OR labels.kind=knada"

	call := svc.Projects.Search().Query(filter)
	for {
		response, err := call.Do()
		if err != nil {
			var retrieve *oauth2.RetrieveError
			if errors.As(err, &retrieve) {
				if retrieve.ErrorCode == "invalid_grant" {
					return nil, fmt.Errorf("looks like you are missing Application Default Credentials, run `gcloud auth application-default login` first\n")
				}
			}

			return nil, err
		}

		for _, project := range response.Projects {
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
