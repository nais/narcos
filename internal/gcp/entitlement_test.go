package gcp_test

import (
	"slices"
	"testing"
	"time"

	pb "cloud.google.com/go/privilegedaccessmanager/apiv1/privilegedaccessmanagerpb"
	"github.com/nais/narcos/internal/gcp"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestEntitlementShortName(t *testing.T) {
	ent := gcp.Entitlement{
		Entitlement: &pb.Entitlement{
			Name: "folders/448765591554/locations/global/entitlements/nais-admin",
		},
	}
	if got, want := ent.ShortName(), "nais-admin"; got != want {
		t.Errorf("ShortName() = %q, want %q", got, want)
	}
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
	if got, want := ent.Roles(), []string{"roles/storage.admin", "roles/compute.admin"}; !slices.Equal(got, want) {
		t.Errorf("Roles() = %q, want %q", got, want)
	}
}

func TestEntitlementMaxDuration(t *testing.T) {
	ent := gcp.Entitlement{
		Entitlement: &pb.Entitlement{
			Name:               "folders/448765591554/locations/global/entitlements/nais-admin",
			MaxRequestDuration: durationpb.New(4 * time.Hour),
		},
	}
	if got, want := ent.MaxDuration(), 4*time.Hour; got != want {
		t.Errorf("MaxDuration() = %v, want %v", got, want)
	}
}

func TestGrantDuration(t *testing.T) {
	grant := gcp.Grant{
		Grant: &pb.Grant{
			RequestedDuration: durationpb.New(24 * time.Hour),
		},
	}
	if got, want := grant.Duration(), 24*time.Hour; got != want {
		t.Errorf("Duration() = %v, want %v", got, want)
	}
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
	const tolerance = 2 * time.Second
	if remaining < time.Hour-tolerance || remaining > time.Hour+tolerance {
		t.Errorf("TimeRemaining() = %v, want %v within %v", remaining, time.Hour, tolerance)
	}
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

	if got := len(entitlements); got != 2 {
		t.Fatalf("got %d entitlements, want 2", got)
	}
	if got := entitlements[0].ShortName(); got != "nais-admin" {
		t.Errorf("entitlements[0].ShortName() = %q, want %q", got, "nais-admin")
	}
	if got, want := entitlements[0].Roles(), []string{"roles/storage.admin", "roles/compute.admin"}; !slices.Equal(got, want) {
		t.Errorf("entitlements[0].Roles() = %q, want %q", got, want)
	}
	if got, want := entitlements[0].MaxDuration(), 4*time.Hour; got != want {
		t.Errorf("entitlements[0].MaxDuration() = %v, want %v", got, want)
	}
	if got := entitlements[1].ShortName(); got != "nais-privileged" {
		t.Errorf("entitlements[1].ShortName() = %q, want %q", got, "nais-privileged")
	}
	if got, want := entitlements[1].Roles(), []string{"roles/container.clusterAdmin"}; !slices.Equal(got, want) {
		t.Errorf("entitlements[1].Roles() = %q, want %q", got, want)
	}
	if got, want := entitlements[1].MaxDuration(), 8*time.Hour; got != want {
		t.Errorf("entitlements[1].MaxDuration() = %v, want %v", got, want)
	}

	const name = "nais-admin"
	found := gcp.GetEntitlementByName(entitlements, name)
	if found == nil {
		t.Fatalf("GetEntitlementByName(%q) = nil, want an entitlement", name)
	}
	if got := found.ShortName(); got != name {
		t.Errorf("GetEntitlementByName(%q).ShortName() = %q, want %q", name, got, name)
	}

	const missingName = "does-not-exist"
	if got := gcp.GetEntitlementByName(entitlements, missingName); got != nil {
		t.Errorf("GetEntitlementByName(%q) returned entitlement %q, want nil", missingName, got.ShortName())
	}
}
