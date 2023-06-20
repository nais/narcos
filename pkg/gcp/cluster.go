package gcp

import (
	"context"
	"encoding/json"
	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/container/v1"
)

type Project struct {
	ID          string
	Tenant      string
	Environment string
	Kind        string
}

type Cluster struct {
	Name     string
	Endpoint string
	Location string
	CA       string
	Tenant   string
	User     *User
}

type User struct {
	ServerID string `json:"serverID"`
	ClientID string `json:"clientID"`
	TenantID string `json:"tenantID"`
	UserName string `json:"userName"`
}

func GetClusters(ctx context.Context, includeManagement, includeOnprem bool, tenant string) ([]Cluster, error) {
	projects, err := getProjects(ctx, includeManagement, includeOnprem, tenant)
	if err != nil {
		return nil, err
	}

	return getClusters(ctx, projects)
}

func getProjects(ctx context.Context, includeManagement, includeOnprem bool, filterTenant string) ([]Project, error) {
	var projects []Project

	svc, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return nil, err
	}

	filter := "labels.naiscluster:true"
	if includeOnprem {
		filter = "(labels.naiscluster:true OR labels.kind:onprem)"
	}
	if !includeManagement {
		filter += " labels.environment:*"
	}
	if filterTenant != "" {
		filter += " labels.tenant:" + filterTenant
	}

	call := svc.Projects.Search().Query(filter)
	for {
		response, err := call.Do()
		if err != nil {
			return nil, err
		}

		for _, project := range response.Projects {
			projects = append(projects, Project{
				ID:          project.ProjectId,
				Tenant:      project.Labels["tenant"],
				Environment: project.Labels["environment"],
				Kind:        project.Labels["kind"],
			})
		}
		if response.NextPageToken == "" {
			break
		}
		call.PageToken(response.NextPageToken)
	}

	return projects, nil
}

func getClusters(ctx context.Context, projects []Project) ([]Cluster, error) {
	var clusters []Cluster
	for _, project := range projects {
		var cluster []Cluster
		var err error

		switch project.Kind {
		case "onprem":
			cluster, err = getOnpremClusters(ctx, project)
		default:
			cluster, err = getGCPClusters(ctx, project)
		}

		if err != nil {
			return nil, err
		}
		clusters = append(clusters, cluster...)
	}

	return clusters, nil
}

func getGCPClusters(ctx context.Context, project Project) ([]Cluster, error) {
	svc, err := container.NewService(ctx)
	if err != nil {
		return nil, err
	}

	call := svc.Projects.Locations.Clusters.List("projects/" + project.ID + "/locations/-")
	response, err := call.Do()
	if err != nil {
		return nil, err
	}

	var clusters []Cluster
	for _, cluster := range response.Clusters {
		clusters = append(clusters, Cluster{
			Name:     cluster.Name,
			Endpoint: "https://" + cluster.Endpoint,
			Location: cluster.Location,
			CA:       cluster.MasterAuth.ClusterCaCertificate,
			Tenant:   project.Tenant,
		})
	}
	return clusters, nil
}

func getOnpremClusters(ctx context.Context, project Project) ([]Cluster, error) {
	if project.Kind != "onprem" {
		return nil, nil
	}

	svc, err := compute.NewService(ctx)
	if err != nil {
		return nil, err
	}
	proj, err := svc.Projects.Get(project.ID).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	var clusters []Cluster
	for _, meta := range proj.CommonInstanceMetadata.Items {
		if meta.Key != "kubeconfig" || meta.Value == nil {
			continue
		}

		config := &struct {
			ServerID string `json:"serverID"`
			ClientID string `json:"clientID"`
			TenantID string `json:"tenantID"`
			URL      string `json:"url"`
			UserName string `json:"userName"`
		}{}
		if err := json.Unmarshal([]byte(*meta.Value), &config); err != nil {
			return nil, err
		}

		clusters = append(clusters, Cluster{
			Name:     project.Environment,
			Endpoint: config.URL,
			User: &User{
				ServerID: config.ServerID,
				ClientID: config.ClientID,
				TenantID: config.TenantID,
				UserName: config.UserName,
			},
		})

		return clusters, nil

	}

	return clusters, nil
}
