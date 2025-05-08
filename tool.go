package routeros

import (
	"context"
	"net/http"

	"github.com/art-frela/routeros/types"
)

type ToolService struct {
	c *Client
}

func (ts *ToolService) Ping(ctx context.Context, req types.EchoRequest) (types.EchoResponse, error) {
	res, err := makeRequest[types.EchoResponse](ctx, ts.c, types.EndpointToolPing, http.MethodPost, req, nil)
	if err != nil {
		return nil, err
	}

	return res, nil
}
