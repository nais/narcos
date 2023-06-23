package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/container/v1"
	"strings"
)

type Project struct {
	ID          string
	Tenant      string
	Environment Environment
	Kind        Kind
}

type Cluster struct {
	Name        string
	Endpoint    string
	Location    string
	CA          string
	Tenant      string
	User        *OnpremUser
	Kind        Kind
	Environment Environment
}

type OnpremUser struct {
	ServerID string `json:"serverID"`
	ClientID string `json:"clientID"`
	TenantID string `json:"tenantID"`
	UserName string `json:"userName"`
}

func GetClusters(ctx context.Context, includeManagement, includeOnprem, prefixTenant, includeKnada, skipNAVPrefix bool, tenant string) ([]Cluster, error) {
	projects, err := getProjects(ctx, includeManagement, includeOnprem, includeKnada, tenant)
	if err != nil {
		return nil, err
	}

	clusters, err := getClusters(ctx, projects)
	if err != nil {
		return nil, err
	}
	if prefixTenant {
		for i, cluster := range clusters {
			if skipNAVPrefix && cluster.Tenant == "nav" {
				continue
			}

			cluster.Name = cluster.Tenant + "-" + strings.TrimPrefix(cluster.Name, "nais-")
			clusters[i] = cluster
		}
	}

	return clusters, nil
}

func getProjects(ctx context.Context, includeManagement, includeOnprem, includeKnada bool, filterTenant string) ([]Project, error) {
	var projects []Project

	svc, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return nil, err
	}

	filter := "(labels.naiscluster=true OR labels.kind=legacy"
	if includeOnprem {
		filter += " OR labels.kind=onprem"
	}
	if includeKnada {
		filter += " OR labels.kind=knada"
	}
	filter += ")"

	if !includeManagement {
		filter += " labels.environment:*"
	}
	if filterTenant != "" {
		filter += " labels.tenant=" + filterTenant
	}

	call := svc.Projects.Search().Query(filter)
	for {
		response, err := call.Do()
		if err != nil {
			if strings.Contains(err.Error(), "invalid_grant") {
				return nil, fmt.Errorf("%v\nlooks like you are missing Application Default Credentials, run `gcloud auth application-default login` first\n", err)
			}

			return nil, err
		}

		for _, project := range response.Projects {
			projects = append(projects, Project{
				ID:          project.ProjectId,
				Tenant:      project.Labels["tenant"],
				Environment: ParseEnvironment(project.Labels["environment"]),
				Kind:        ParseKind(project.Labels["kind"]),
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
		case KindOnprem:
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
		name := cluster.Name
		if cluster.Name == "knada-gke" {
			name = "knada"
		}

		clusters = append(clusters, Cluster{
			Name:        name,
			Endpoint:    "https://" + cluster.Endpoint,
			Location:    cluster.Location,
			CA:          cluster.MasterAuth.ClusterCaCertificate,
			Tenant:      project.Tenant,
			Kind:        project.Kind,
			Environment: project.Environment,
		})
	}
	return clusters, nil
}

func getOnpremClusters(ctx context.Context, project Project) ([]Cluster, error) {
	if project.Kind != KindOnprem {
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
			Name:     project.Environment.String(),
			Endpoint: config.URL,
			Tenant:   "nav",
			Kind:     KindOnprem,
			User: &OnpremUser{
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
