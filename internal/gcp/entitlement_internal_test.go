package gcp

import (
	"slices"
	"strings"
	"testing"

	pb "cloud.google.com/go/privilegedaccessmanager/apiv1/privilegedaccessmanagerpb"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func TestParseTenantNames(t *testing.T) {
	bucketXML := `<?xml version="1.0" encoding="UTF-8"?>
<ListBucketResult xmlns="http://doc.s3.amazonaws.com/2006-03-01">
  <Contents><Key>nav.no.json</Key></Contents>
  <Contents><Key>docs/nav/prod.json</Key></Contents>
  <Contents><Key>docs/nav/tenant.json</Key></Contents>
  <Contents><Key>test-nais.no.json</Key></Contents>
  <Contents><Key>README.txt</Key></Contents>
</ListBucketResult>`

	tenants, err := parseTenantNames(strings.NewReader(bucketXML))
	if err != nil {
		t.Fatalf("parseTenantNames() returned an error: %v", err)
	}
	if want := []string{"nav.no", "test-nais.no"}; !slices.Equal(tenants, want) {
		t.Errorf("parseTenantNames() = %q, want %q", tenants, want)
	}
}

func TestParseTenantNamesInvalidXML(t *testing.T) {
	if _, err := parseTenantNames(strings.NewReader("not xml")); err == nil {
		t.Error("parseTenantNames() returned no error for invalid XML")
	}
}

func TestGrantActivationError(t *testing.T) {
	grant := &pb.Grant{
		Timeline: &pb.Grant_Timeline{
			Events: []*pb.Grant_Timeline_Event{
				{Event: &pb.Grant_Timeline_Event_Requested_{}},
				{Event: &pb.Grant_Timeline_Event_ActivationFailed_{
					ActivationFailed: &pb.Grant_Timeline_Event_ActivationFailed{
						Error: &status.Status{Code: 400, Message: "One or more users named in the policy do not belong to a permitted customer."},
					},
				}},
			},
		},
	}
	if got, want := grantActivationError(grant), "One or more users named in the policy do not belong to a permitted customer."; got != want {
		t.Errorf("grantActivationError() = %q, want %q", got, want)
	}
}

func TestGrantActivationErrorNoFailure(t *testing.T) {
	const want = "unknown reason"
	if got := grantActivationError(&pb.Grant{}); got != want {
		t.Errorf("grantActivationError() for a grant without a timeline = %q, want %q", got, want)
	}

	grant := &pb.Grant{
		Timeline: &pb.Grant_Timeline{
			Events: []*pb.Grant_Timeline_Event{
				{Event: &pb.Grant_Timeline_Event_Requested_{}},
				{Event: &pb.Grant_Timeline_Event_ActivationFailed_{ActivationFailed: &pb.Grant_Timeline_Event_ActivationFailed{}}},
			},
		},
	}
	if got := grantActivationError(grant); got != want {
		t.Errorf("grantActivationError() for a failure without an error message = %q, want %q", got, want)
	}
}
