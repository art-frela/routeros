package routeros

import (
	"context"
	"net/http"

	"github.com/art-frela/routeros/types"
)

// ToolService handles communication with the tool methods
// of the RouterOS API docs.
//
// RouterOS API docs: https://manual.mikrotik.com/docs/Developer+Guides/rest-api
type ToolService struct {
	c *Client
}

// Ping performs a ping test to a specified host address.
//
// RouterOS API docs: https://manual.mikrotik.com/docs/Developer+Guides/rest-api
func (ts *ToolService) Ping(ctx context.Context, req types.EchoRequest) (types.EchoResponse, error) {
	res, err := makeRequest[types.EchoResponse](ctx, ts.c, types.EndpointToolPing, http.MethodPost, req, nil)
	if err != nil {
		return nil, err
	}

	return res, nil
}
