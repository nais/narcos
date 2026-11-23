package gcp

import (
	"strings"
	"testing"

	pb "cloud.google.com/go/privilegedaccessmanager/apiv1/privilegedaccessmanagerpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)
	assert.Equal(t, []string{"nav.no", "test-nais.no"}, tenants)
}

func TestParseTenantNamesInvalidXML(t *testing.T) {
	_, err := parseTenantNames(strings.NewReader("not xml"))
	assert.Error(t, err)
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
	assert.Equal(t, "One or more users named in the policy do not belong to a permitted customer.", grantActivationError(grant))
}

func TestGrantActivationErrorNoFailure(t *testing.T) {
	assert.Equal(t, "unknown reason", grantActivationError(&pb.Grant{}))

	grant := &pb.Grant{
		Timeline: &pb.Grant_Timeline{
			Events: []*pb.Grant_Timeline_Event{
				{Event: &pb.Grant_Timeline_Event_Requested_{}},
				{Event: &pb.Grant_Timeline_Event_ActivationFailed_{ActivationFailed: &pb.Grant_Timeline_Event_ActivationFailed{}}},
			},
		},
	}
	assert.Equal(t, "unknown reason", grantActivationError(grant))
}
