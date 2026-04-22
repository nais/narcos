package fasit

import (
	"context"
	"encoding/json"

	genqlientgraphql "github.com/Khan/genqlient/graphql"
)

type mockGraphQLClient struct {
	responses map[string]any
	errors    map[string]error
	lastReq   *genqlientgraphql.Request
}

func (m *mockGraphQLClient) MakeRequest(_ context.Context, req *genqlientgraphql.Request, resp *genqlientgraphql.Response) error {
	m.lastReq = req
	if err := m.errors[req.OpName]; err != nil {
		return err
	}

	data := m.responses[req.OpName]
	if data == nil {
		return nil
	}

	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, resp.Data)
}
