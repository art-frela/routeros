package routeros

import (
	"context"
	"net/http"
	"net/url"

	"github.com/art-frela/routeros/types"
)

type IPService struct {
	c *Client
}

func (ips *IPService) GetAddresses(ctx context.Context) (types.IPAddressList, error) {
	return makeRequest[types.IPAddressList](ctx, ips.c, types.EndpointIPAddresses, http.MethodGet, nil, nil)
}

func (ips *IPService) GetAddressByID(ctx context.Context, id string) (types.IPAddress, error) {
	var queries url.Values
	queries = make(url.Values)
	queries[".id"] = []string{id}

	res, err := makeRequest[types.IPAddressList](ctx, ips.c, types.EndpointIPAddresses, http.MethodGet, nil, queries)
	if err != nil {
		return types.IPAddress{}, err
	}

	if len(res) == 0 {
		return types.IPAddress{}, nil
	}

	return res[0], nil
}

func (ips *IPService) AddAddress(ctx context.Context, item types.IPAddressAdd) (types.IPAddress, error) {
	res, err := makeRequest[types.IPAddress](ctx, ips.c, types.EndpointIPAddresses, http.MethodPut, item, nil)
	if err != nil {
		return types.IPAddress{}, err
	}

	return res, nil
}

func (ips *IPService) RemoveAddress(ctx context.Context, id string) error {
	queries := url.Values{".id": []string{id}}
	_, err := makeRequest[any](ctx, ips.c, types.EndpointIPAddresses, http.MethodDelete, nil, queries)
	return err
}

func (ips *IPService) UpdateAddress(ctx context.Context, id string, item types.IPAddressAdd) (types.IPAddress, error) {
	queries := url.Values{".id": []string{id}}
	res, err := makeRequest[types.IPAddress](ctx, ips.c, types.EndpointIPAddresses, http.MethodPatch, item, queries)
	if err != nil {
		return types.IPAddress{}, err
	}

	return res, nil
}
