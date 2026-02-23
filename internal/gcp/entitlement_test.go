package gcp_test

import (
	"testing"
	"time"

	pb "cloud.google.com/go/privilegedaccessmanager/apiv1/privilegedaccessmanagerpb"
	"github.com/nais/narcos/internal/gcp"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestEntitlementShortName(t *testing.T) {
	ent := gcp.Entitlement{
		Entitlement: &pb.Entitlement{
			Name: "folders/448765591554/locations/global/entitlements/nais-admin",
		},
	}
	assert.Equal(t, "nais-admin", ent.ShortName())
}

func TestEntitlementRoles(t *testing.T) {
	ent := gcp.Entitlement{
		Entitlement: &pb.Entitlement{
			Name: "folders/448765591554/locations/global/entitlements/nais-admin",
			PrivilegedAccess: &pb.PrivilegedAccess{
				AccessType: &pb.PrivilegedAccess_GcpIamAccess_{
					GcpIamAccess: &pb.PrivilegedAccess_GcpIamAccess{
						RoleBindings: []*pb.PrivilegedAccess_GcpIamAccess_RoleBinding{
							{Role: "roles/storage.admin"},
							{Role: "roles/compute.admin"},
						},
					},
				},
			},
		},
	}
	assert.Equal(t, []string{"roles/storage.admin", "roles/compute.admin"}, ent.Roles())
}

func TestEntitlementMaxDuration(t *testing.T) {
	ent := gcp.Entitlement{
		Entitlement: &pb.Entitlement{
			Name:               "folders/448765591554/locations/global/entitlements/nais-admin",
			MaxRequestDuration: durationpb.New(4 * time.Hour),
		},
	}
	assert.Equal(t, 4*time.Hour, ent.MaxDuration())
}

func TestGrantDuration(t *testing.T) {
	grant := gcp.Grant{
		Grant: &pb.Grant{
			RequestedDuration: durationpb.New(24 * time.Hour),
		},
	}
	assert.Equal(t, 24*time.Hour, grant.Duration())
}

func TestGrantTimeRemaining(t *testing.T) {
	now := time.Now()
	grant := gcp.Grant{
		Grant: &pb.Grant{
			CreateTime:        timestamppb.New(now),
			RequestedDuration: durationpb.New(1 * time.Hour),
		},
	}
	remaining := grant.TimeRemaining()
	// Should be close to 1 hour
	assert.InDelta(t, time.Hour.Seconds(), remaining.Seconds(), 2)
}

func TestGetEntitlementByName(t *testing.T) {
	entitlements := []gcp.Entitlement{
		{
			Entitlement: &pb.Entitlement{
				Name: "folders/448765591554/locations/global/entitlements/nais-admin",
				PrivilegedAccess: &pb.PrivilegedAccess{
					AccessType: &pb.PrivilegedAccess_GcpIamAccess_{
						GcpIamAccess: &pb.PrivilegedAccess_GcpIamAccess{
							RoleBindings: []*pb.PrivilegedAccess_GcpIamAccess_RoleBinding{
								{Role: "roles/storage.admin"},
								{Role: "roles/compute.admin"},
							},
						},
					},
				},
				MaxRequestDuration: durationpb.New(4 * time.Hour),
			},
		},
		{
			Entitlement: &pb.Entitlement{
				Name: "folders/448765591554/locations/global/entitlements/nais-privileged",
				PrivilegedAccess: &pb.PrivilegedAccess{
					AccessType: &pb.PrivilegedAccess_GcpIamAccess_{
						GcpIamAccess: &pb.PrivilegedAccess_GcpIamAccess{
							RoleBindings: []*pb.PrivilegedAccess_GcpIamAccess_RoleBinding{
								{Role: "roles/container.clusterAdmin"},
							},
						},
					},
				},
				MaxRequestDuration: durationpb.New(8 * time.Hour),
			},
		},
	}

	assert.Len(t, entitlements, 2)

	assert.Equal(t, "nais-admin", entitlements[0].ShortName())
	assert.Equal(t, []string{"roles/storage.admin", "roles/compute.admin"}, entitlements[0].Roles())
	assert.Equal(t, 4*time.Hour, entitlements[0].MaxDuration())

	assert.Equal(t, "nais-privileged", entitlements[1].ShortName())
	assert.Equal(t, []string{"roles/container.clusterAdmin"}, entitlements[1].Roles())
	assert.Equal(t, 8*time.Hour, entitlements[1].MaxDuration())

	found := gcp.GetEntitlementByName(entitlements, "nais-admin")
	assert.NotNil(t, found)
	assert.Equal(t, "nais-admin", found.ShortName())

	notFound := gcp.GetEntitlementByName(entitlements, "does-not-exist")
	assert.Nil(t, notFound)
}
