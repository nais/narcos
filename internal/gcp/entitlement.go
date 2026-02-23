package gcp

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"time"

	privilegedaccessmanager "cloud.google.com/go/privilegedaccessmanager/apiv1"
	pb "cloud.google.com/go/privilegedaccessmanager/apiv1/privilegedaccessmanagerpb"
	"google.golang.org/protobuf/types/known/durationpb"
)

type FolderID string

func (folderID FolderID) entitlementsParent() string {
	return fmt.Sprintf("%s/locations/global", folderID)
}

// nais-terraform-modules exports tenant metadata through a public Google storage bucket.
//
// Each tenant corresponds to a single file on this bucket.
// The file has the same name as the tenant domain, suffixed with .json.
type TenantMetadata struct {
	NaisFolderID FolderID `json:"folderId"`
}

type tenant struct {
	Name string `xml:"Key"`
}

type bucket struct {
	Name        string   `xml:"Name"`
	Prefix      string   `xml:"Prefix"`
	Marker      string   `xml:"Marker"`
	IsTruncated bool     `xml:"IsTruncated"`
	Contents    []tenant `xml:"Contents"`
}

// FetchAllTenantNames returns a list of known tenant names by listing files in a public Google storage bucket.
func FetchAllTenantNames(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://storage.googleapis.com/nais-tenant-data", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %q", resp.Status)
	}

	decoder := xml.NewDecoder(resp.Body)
	var bucket bucket
	err = decoder.Decode(&bucket)

	tenants := make([]string, 0)
	for _, tenant := range bucket.Contents {
		if !strings.HasSuffix(tenant.Name, ".json") {
			continue
		}
		tenants = append(tenants, strings.TrimSuffix(tenant.Name, ".json"))
	}
	return tenants, err
}

// FetchTenantMetadata returns metadata for a given tenant.
func FetchTenantMetadata(ctx context.Context, tenantName string) (*TenantMetadata, error) {
	u := fmt.Sprintf("https://storage.googleapis.com/nais-tenant-data/%s.json", tenantName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("unknown tenant %q", tenantName)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %q", resp.Status)
	}

	metadata := &TenantMetadata{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(metadata)

	return metadata, err
}

// Entitlement wraps the protobuf Entitlement type with convenience methods.
type Entitlement struct {
	*pb.Entitlement
}

// Convert `folders/448765591554/locations/global/entitlements/nais-admin` -> `nais-admin`
func (ent Entitlement) ShortName() string {
	parts := strings.Split(ent.GetName(), "/")
	return parts[len(parts)-1]
}

// Extract roles as a simple slice
func (ent Entitlement) Roles() []string {
	roles := make([]string, 0)
	pa := ent.GetPrivilegedAccess()
	if pa == nil {
		return roles
	}
	gcpIAM := pa.GetGcpIamAccess()
	if gcpIAM == nil {
		return roles
	}
	for _, rb := range gcpIAM.GetRoleBindings() {
		roles = append(roles, rb.GetRole())
	}
	return roles
}

// Parse duration to a known type
func (ent Entitlement) MaxDuration() time.Duration {
	d := ent.GetMaxRequestDuration()
	if d == nil {
		return 0
	}
	return d.AsDuration()
}

// Grant wraps the protobuf Grant type with convenience methods.
type Grant struct {
	*pb.Grant
}

func (grant Grant) Duration() time.Duration {
	d := grant.GetRequestedDuration()
	if d == nil {
		return 0
	}
	return d.AsDuration()
}

func (grant Grant) TimeRemaining() time.Duration {
	ct := grant.GetCreateTime()
	if ct == nil {
		return 0
	}
	grantTime := ct.AsTime()
	expires := grantTime.Add(grant.Duration())
	return time.Until(expires).Truncate(time.Second)
}

// NewPAMClient creates a new Privileged Access Manager client.
// It uses Application Default Credentials for authentication.
func NewPAMClient(ctx context.Context) (*privilegedaccessmanager.Client, error) {
	return privilegedaccessmanager.NewRESTClient(ctx)
}

// Return a list of possible entitlements that can be granted.
//
// The folder ID is a reference to the `nais` folder of a specific tenant.
func ListEntitlements(ctx context.Context, folderID FolderID) ([]Entitlement, error) {
	client, err := NewPAMClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating PAM client: %w", err)
	}
	defer func() { _ = client.Close() }()

	req := &pb.SearchEntitlementsRequest{
		Parent:           folderID.entitlementsParent(),
		CallerAccessType: pb.SearchEntitlementsRequest_GRANT_REQUESTER,
	}

	var entitlements []Entitlement
	it := client.SearchEntitlements(ctx, req)
	for {
		ent, err := it.Next()
		if err != nil {
			// iterator.Done
			break
		}
		entitlements = append(entitlements, Entitlement{ent})
	}

	if len(entitlements) == 0 {
		return nil, fmt.Errorf("no entitlements found: ensure you are logged into gcloud")
	}

	return entitlements, nil
}

// GetEntitlementByName finds an entitlement by its short name.
func GetEntitlementByName(entitlements []Entitlement, name string) *Entitlement {
	for _, ent := range entitlements {
		if name == ent.ShortName() {
			return &ent
		}
	}
	return nil
}

// ListActiveGrants lists all active grants for a given entitlement.
func ListActiveGrants(ctx context.Context, entitlementName string, userName string) ([]Grant, error) {
	client, err := NewPAMClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating PAM client: %w", err)
	}
	defer func() { _ = client.Close() }()

	req := &pb.ListGrantsRequest{
		Parent:   entitlementName,
		Filter:   fmt.Sprintf(`state = "ACTIVE" AND requester = "%s"`, userName),
		PageSize: 500,
	}

	var grants []Grant
	it := client.ListGrants(ctx, req)
	for {
		g, err := it.Next()
		if err != nil {
			// iterator.Done
			break
		}
		grants = append(grants, Grant{g})
	}

	return grants, nil
}

// RevokeGrant revokes an active grant, ending the privilege elevation early.
func RevokeGrant(ctx context.Context, grantName string, reason string) error {
	client, err := NewPAMClient(ctx)
	if err != nil {
		return fmt.Errorf("creating PAM client: %w", err)
	}
	defer func() { _ = client.Close() }()

	req := &pb.RevokeGrantRequest{
		Name:   grantName,
		Reason: reason,
	}

	op, err := client.RevokeGrant(ctx, req)
	if err != nil {
		return err
	}

	_, err = op.Wait(ctx)
	return err
}

// ElevatePrivileges requests a "grant" for the "entitlement" at Google APIs.
func ElevatePrivileges(ctx context.Context, ent Entitlement, duration time.Duration, justification string) error {
	client, err := NewPAMClient(ctx)
	if err != nil {
		return fmt.Errorf("creating PAM client: %w", err)
	}
	defer func() { _ = client.Close() }()

	req := &pb.CreateGrantRequest{
		Parent: ent.GetName(),
		Grant: &pb.Grant{
			RequestedDuration: durationpb.New(duration),
			Justification: &pb.Justification{
				Justification: &pb.Justification_UnstructuredJustification{
					UnstructuredJustification: justification,
				},
			},
		},
	}

	_, err = client.CreateGrant(ctx, req)
	if err != nil {
		return err
	}

	return nil
}
