package routeros

import (
	"context"
	"net/http"
	"net/url"

	"github.com/art-frela/routeros/types"
)

// IPRouteService handles communication with the IP route methods
// of the RouterOS REST API.
//
// RouterOS REST API: https://manual.mikrotik.com/docs/Developer+Guides/rest-api
type IPRouteService struct {
	c *Client
}

// GetRoutes returns a list of all IP routes configured on the device.
//
// RouterOS REST API: https://manual.mikrotik.com/docs/Developer+Guides/rest-api
func (s *IPRouteService) GetRoutes(ctx context.Context) (types.IPRouteList, error) {
	return makeRequest[types.IPRouteList](ctx, s.c, types.EndpointIPRoutes, http.MethodGet, nil, nil)
}

// GetRouteByID returns a specific IP route by its ID.
//
// RouterOS API docs: https://manual.mikrotik.com/docs/Developer+Guides/rest-api
func (s *IPRouteService) GetRouteByID(ctx context.Context, id string) (types.IPRoute, error) {
	if id == "" {
		return types.IPRoute{}, nil
	}

	queries := url.Values{".id": []string{id}}

	res, err := makeRequest[types.IPRouteList](ctx, s.c, types.EndpointIPRoutes, http.MethodGet, nil, queries)
	if err != nil {
		return types.IPRoute{}, err
	}

	if len(res) == 0 {
		return types.IPRoute{}, nil
	}

	return res[0], nil
}

// AddRoute adds a new static route.
//
// RouterOS REST API: https://manual.mikrotik.com/docs/Developer+Guides/rest-api
func (s *IPRouteService) AddRoute(ctx context.Context, item types.IPRouteAdd) (types.IPRoute, error) {
	res, err := makeRequest[types.IPRoute](ctx, s.c, types.EndpointIPRoutes, http.MethodPut, item, nil)
	if err != nil {
		return types.IPRoute{}, err
	}

	return res, nil
}

// RemoveRoute removes an IP route by its ID.
//
// RouterOS REST API: https://manual.mikrotik.com/docs/Developer+Guides/rest-api
func (s *IPRouteService) RemoveRoute(ctx context.Context, id string) error {
	queries := url.Values{".id": []string{id}}
	_, err := makeRequest[any](ctx, s.c, types.EndpointIPRoutes, http.MethodDelete, nil, queries)

	return err
}

// UpdateRoute updates an existing IP route identified by its ID.
//
// RouterOS REST API: https://manual.mikrotik.com/docs/Developer+Guides/rest-api
func (s *IPRouteService) UpdateRoute(ctx context.Context, id string, item types.IPRouteAdd) (types.IPRoute, error) {
	queries := url.Values{".id": []string{id}}
	res, err := makeRequest[types.IPRoute](ctx, s.c, types.EndpointIPRoutes, http.MethodPatch, item, queries)
	if err != nil {
		return types.IPRoute{}, err
	}

	return res, nil
}
