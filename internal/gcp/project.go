package gcp

import (
	"context"
	"fmt"

	"google.golang.org/api/cloudresourcemanager/v3"
)

func FindProjectByTeamEnv(ctx context.Context, tenant, team, env string) (*cloudresourcemanager.Project, error) {
	crm, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("create cloud resource manager client: %w", err)
	}

	query := fmt.Sprintf("(labels.tenant:%v AND labels.team:%v AND labels.environment:%v)", tenant, team, env)
	queryResponse, err := crm.Projects.Search().Query(query).Do()
	if err != nil {
		return nil, fmt.Errorf("search for projects: %w, query was: %v", err, query)
	}

	projects := queryResponse.Projects
	if len(projects) == 0 {
		return nil, fmt.Errorf("no projects matched the query, query was: %v", query)
	}

	if len(projects) > 1 {
		projectIds := []string{}
		for _, project := range projects {
			projectIds = append(projectIds, project.ProjectId)
		}
		return nil, fmt.Errorf("multiple projects matched the query: %v, query was: %v", projectIds, query)
	}

	return projects[0], nil
}
