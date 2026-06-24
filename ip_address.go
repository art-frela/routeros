package routeros

import (
	"context"
	"net/http"
	"net/url"

	"github.com/art-frela/routeros/types"
)

// IPService handles communication with the IP address methods
// of the RouterOS REST API.
//
// RouterOS API docs: https://manual.mikrotik.com/docs/Developer+Guides/rest-api
type IPService struct {
	c *Client
}

// GetAddresses returns a list of all IP addresses configured on the device.
//
// RouterOS API docs: https://manual.mikrotik.com/docs/Developer+Guides/rest-api
func (ips *IPService) GetAddresses(ctx context.Context) (types.IPAddressList, error) {
	return makeRequest[types.IPAddressList](ctx, ips.c, types.EndpointIPAddresses, http.MethodGet, nil, nil)
}

// GetAddressByID returns a specific IP address by its ID.
//
// RouterOS API docs: https://manual.mikrotik.com/docs/Developer+Guides/rest-api
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

// AddAddress adds a new IP address.
//
// RouterOS API docs: https://manual.mikrotik.com/docs/Developer+Guides/rest-api
func (ips *IPService) AddAddress(ctx context.Context, item types.IPAddressAdd) (types.IPAddress, error) {
	res, err := makeRequest[types.IPAddress](ctx, ips.c, types.EndpointIPAddresses, http.MethodPut, item, nil)
	if err != nil {
		return types.IPAddress{}, err
	}

	return res, nil
}

// RemoveAddress removes an IP address by its ID.
//
// RouterOS API docs: https://manual.mikrotik.com/docs/Developer+Guides/rest-api
func (ips *IPService) RemoveAddress(ctx context.Context, id string) error {
	queries := url.Values{".id": []string{id}}
	_, err := makeRequest[any](ctx, ips.c, types.EndpointIPAddresses, http.MethodDelete, nil, queries)
	return err
}

// UpdateAddress updates an existing IP address identified by its ID.
//
// RouterOS API docs: https://manual.mikrotik.com/docs/Developer+Guides/rest-api
func (ips *IPService) UpdateAddress(ctx context.Context, id string, item types.IPAddressAdd) (types.IPAddress, error) {
	queries := url.Values{".id": []string{id}}
	res, err := makeRequest[types.IPAddress](ctx, ips.c, types.EndpointIPAddresses, http.MethodPatch, item, queries)
	if err != nil {
		return types.IPAddress{}, err
	}

	return res, nil
}
