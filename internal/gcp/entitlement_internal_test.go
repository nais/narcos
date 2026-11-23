package gcp

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
