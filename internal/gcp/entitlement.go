package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type FolderID int

// FIXME: this is a mockup; get real data from a bucket instead
func TenantNaisFolderIDMapping() map[string]FolderID {
	return map[string]FolderID{
		"dev-nais.io": 0,
	}
}

// bucket://FOO/nav.no.json
// en fil per tenant
// laget av TF
type DetSomLiggerPaaBoetta struct {
	//TenantName       string
	NaisFolderID     FolderID
	EntitlementNames []string // [nais-admin, nais-viewer]
}

func ParseEntitlementResponse(entitlementData []byte) (EntitlementsResponse, error) {
	var resp EntitlementsResponse

	err := json.Unmarshal(entitlementData, &resp)
	return resp, err
}

// Return a list of possible entitlements that can be granted.
//
// The folder ID is a reference to the `nais` folder of a specific tenant.
func ListEntitlements(ctx context.Context, folderID FolderID) (*EntitlementsResponse, error) {
	id := entitlementsID(folderID)
	url := fmt.Sprintf(
		"https://privilegedaccessmanager.googleapis.com/v1beta/%s:search?callerAccessType=GRANT_REQUESTER",
		id,
	)

	accessToken, err := GCloudAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	data, err := ParseEntitlementResponse(body)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

type Entitlement struct {
	Name               string `json:"name"`
	MaxRequestDuration string `json:"maxRequestDuration"`
	PrivilegedAccess   struct {
		GCPIAMAccess struct {
			RoleBindings []struct {
				Role string `json:"role"`
			} `json:"roleBindings"`
		} `json:"gcpIamAccess"`
	} `json:"privilegedAccess"`
}

// Convert `folders/448765591554/locations/global/entitlements/nais-admin` -> `nais-admin`
func (ent Entitlement) ShortName() string {
	parts := strings.Split(ent.Name, "/")
	return parts[len(parts)-1]
}

// Extract roles as a simple slice
func (ent Entitlement) Roles() []string {
	roles := make([]string, 0)
	for _, role := range ent.PrivilegedAccess.GCPIAMAccess.RoleBindings {
		roles = append(roles, role.Role)
	}
	return roles
}

// Parse duration to a known type
func (ent Entitlement) MaxDuration() time.Duration {
	duration, _ := time.ParseDuration(ent.MaxRequestDuration)
	return duration
}

// Actual Entitlements response from GCP
type EntitlementsResponse struct {
	Entitlements []Entitlement `json:"entitlements"`
}

func (r EntitlementsResponse) GetByName(tenantName string) *Entitlement {
	for _, entitlement := range r.Entitlements {
		if tenantName == entitlement.ShortName() {
			return &entitlement
		}
	}
	return nil
}

func entitlementsID(folderID FolderID) string {
	return fmt.Sprintf("folders/%d/locations/global/entitlements", folderID)
}
