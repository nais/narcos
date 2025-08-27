package gcp

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type FolderID string

func (folderID FolderID) entitlementsID() string {
	return fmt.Sprintf("%s/locations/global/entitlements", folderID)
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

// From Google API.
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

// List all grants for a given entitlement, looping through pagination as needed.
func (ent Entitlement) ListActiveGrants(ctx context.Context, userName string) ([]Grant, error) {
	const urlTemplate = "https://privilegedaccessmanager.googleapis.com/v1beta/%s/grants"

	urlBase := fmt.Sprintf(urlTemplate, ent.Name)

	accessToken, err := GCloudAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	grants := make([]Grant, 0)

	urlValues := url.Values{}
	urlValues.Set("filter", fmt.Sprintf(`state = "ACTIVE" AND requester = "%s"`, userName))
	urlValues.Set("pageSize", "500")
	requestURL := urlBase + "?" + urlValues.Encode()

	for len(requestURL) > 0 {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Accept", "application/json")

		body, err := func(req *http.Request) ([]byte, error) {
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return nil, err
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("server returned %q", resp.Status)
			}

			return io.ReadAll(resp.Body)
		}(req)
		if err != nil {
			return nil, err
		}

		grantsResponse, err := ParseGrantsResponse(body)
		if err != nil {
			return nil, err
		}

		grants = append(grants, grantsResponse.Grants...)
		if len(grantsResponse.NextPageToken) == 0 {
			break
		}

		urlValues.Set("pageToken", grantsResponse.NextPageToken)
		requestURL = urlBase + "?" + urlValues.Encode()
	}

	return grants, nil
}

// From Google API.
type Justification struct {
	Text string `json:"unstructuredJustification"`
}

type Grant struct {
	// Name              string `json:"name"`
	CreateTime        string        `json:"createTime,omitempty"`
	Requester         string        `json:"requester,omitempty"`
	RequestedDuration string        `json:"requestedDuration"`
	Justification     Justification `json:"justification"`
}

func (grant Grant) Duration() time.Duration {
	requestedDuration, _ := time.ParseDuration(grant.RequestedDuration)
	return requestedDuration
}

func (grant Grant) TimeRemaining() time.Duration {
	grantTime, err := time.Parse(time.RFC3339, grant.CreateTime)
	if err != nil {
		return 0
	}

	expires := grantTime.Add(grant.Duration())

	return time.Until(expires).Truncate(time.Second)
}

// https://cloud.google.com/iam/docs/reference/pam/rest/v1beta/ListGrantsResponse
type ListGrantsResponse struct {
	Grants        []Grant `json:"grants"`
	NextPageToken string  `json:"nextPageToken"`
}

// https://cloud.google.com/iam/docs/reference/pam/rest/v1beta/ListGrantsResponse
func ParseGrantsResponse(grantsData []byte) (*ListGrantsResponse, error) {
	var resp ListGrantsResponse
	err := json.Unmarshal(grantsData, &resp)
	if err != nil {
		return nil, fmt.Errorf("json decode error: %w", err)
	}
	return &resp, nil
}

// Create a Grant object needed to elevate privileges.
//
// https://cloud.google.com/iam/docs/reference/pam/rest/v1beta/folders.locations.entitlements.grants#Grant.Justification
// https://cloud.google.com/iam/docs/pam-request-temporary-elevated-access#iam-pam-request-grants-search-rest
// https://protobuf.dev/reference/protobuf/google.protobuf/#duration
func NewGrant(duration time.Duration, justification string) Grant {
	return Grant{
		RequestedDuration: fmt.Sprintf("%.0fs", duration.Seconds()),
		Justification: Justification{
			Text: justification,
		},
	}
}

// Request a "grant" for the "entitlement" at Google APIs
//
// https://cloud.google.com/iam/docs/reference/pam/rest/v1beta/folders.locations.entitlements.grants/create
func ElevatePrivileges(ctx context.Context, ent Entitlement, grant Grant) error {
	accessToken, err := GCloudAccessToken(ctx)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(grant)
	if err != nil {
		return err
	}

	u := fmt.Sprintf("https://privilegedaccessmanager.googleapis.com/v1beta/%s/grants", ent.Name)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusBadRequest {
		errorMessage, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %q: %q", resp.Status, errorMessage)
	}

	return nil
}

// Return a list of possible entitlements that can be granted.
//
// The folder ID is a reference to the `nais` folder of a specific tenant.
func ListEntitlements(ctx context.Context, folderID FolderID) (*EntitlementsResponse, error) {
	accessToken, err := GCloudAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	u := fmt.Sprintf(
		"https://privilegedaccessmanager.googleapis.com/v1beta/%s:search?callerAccessType=GRANT_REQUESTER",
		folderID.entitlementsID(),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	data, err := ParseEntitlementResponse(body)
	if err != nil {
		return nil, err
	}
	if len(data.Entitlements) == 0 {
		return nil, fmt.Errorf("no entitlements found: ensure you are logged into gcloud")
	}

	return &data, nil
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

func ParseEntitlementResponse(entitlementData []byte) (EntitlementsResponse, error) {
	var resp EntitlementsResponse
	err := json.Unmarshal(entitlementData, &resp)
	return resp, err
}
