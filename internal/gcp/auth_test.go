package gcp_test

import (
	"context"
	"testing"

	"github.com/nais/narcos/internal/gcp"
	"github.com/stretchr/testify/assert"
)

func TestGcloudAccessToken(t *testing.T) {
	t.SkipNow() // enable for debugging if needed

	token, err := gcp.GCloudAccessToken(context.Background())
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}
