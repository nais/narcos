package gcp_test

import (
	"testing"
	"time"

	"github.com/nais/narcos/internal/gcp"
	"github.com/stretchr/testify/assert"
)

func TestGrantFormat(t *testing.T) {
	grant := gcp.NewGrant(time.Hour*24, "foo")
	assert.Equal(t, "86400s", grant.RequestedDuration)
	assert.Equal(t, "foo", grant.Justification.Text)

	grant = gcp.NewGrant(time.Millisecond*1600, "foobar")
	assert.Equal(t, "2s", grant.RequestedDuration)
	assert.Equal(t, "foobar", grant.Justification.Text)
}

func TestEntitlementParsing(t *testing.T) {
	resp, err := gcp.ParseEntitlementResponse([]byte(entitlementData))
	assert.NoError(t, err)

	assert.Len(t, resp.Entitlements, 2)

	assert.Equal(t, "nais-admin", resp.Entitlements[0].ShortName())
	assert.Equal(t, []string{"roles/storage.admin", "roles/compute.admin"}, resp.Entitlements[0].Roles())
	assert.Equal(t, time.Hour*4, resp.Entitlements[0].MaxDuration())

	assert.Equal(t, "nais-privileged", resp.Entitlements[1].ShortName())
	assert.Equal(t, []string{"roles/container.clusterAdmin"}, resp.Entitlements[1].Roles())
	assert.Equal(t, time.Hour*8, resp.Entitlements[1].MaxDuration())
}

var entitlementData = `
{
  "entitlements": [
    {
      "name": "folders/448765591554/locations/global/entitlements/nais-admin",
      "createTime": "2025-01-29T08:02:44.299558404Z",
      "updateTime": "2025-01-29T08:02:47.861819618Z",
      "privilegedAccess": {
        "gcpIamAccess": {
          "resourceType": "cloudresourcemanager.googleapis.com/Folder",
          "resource": "//cloudresourcemanager.googleapis.com/folders/448765591554",
          "roleBindings": [
            {
              "role": "roles/storage.admin"
            },
            {
              "role": "roles/compute.admin"
            }
          ]
        }
      },
      "maxRequestDuration": "14400s",
      "state": "AVAILABLE",
      "requesterJustificationConfig": {
        "unstructured": {}
      },
      "etag": "\"NjFhZTY2MTgtNGI3Ni00M2ExLTk0NGYtNjkxZmIxYzUzZDI2BwYs07wfkH0=\""
    },
    {
      "name": "folders/448765591554/locations/global/entitlements/nais-privileged",
      "createTime": "2025-01-16T09:32:10.580175887Z",
      "updateTime": "2025-01-16T09:32:14.555020366Z",
      "privilegedAccess": {
        "gcpIamAccess": {
          "resourceType": "cloudresourcemanager.googleapis.com/Folder",
          "resource": "//cloudresourcemanager.googleapis.com/folders/448765591554",
          "roleBindings": [
            {
              "role": "roles/container.clusterAdmin"
            }
          ]
        }
      },
      "maxRequestDuration": "28800s",
      "state": "AVAILABLE",
      "requesterJustificationConfig": {
        "unstructured": {}
      },
      "additionalNotificationTargets": {
        "adminEmailRecipients": [
          "foo.bar@nav.no"
        ]
      },
      "etag": "\"MmViZTQ3Y2QtNDBjNi00YTJmLTg3MTEtYjI5NjBmMmIwNzE3BwYrz3gMWkY=\""
    }
  ]
}
`
