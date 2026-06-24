package routeros

import (
	"context"
	"net/http"
	"net/url"

	"github.com/art-frela/routeros/types"
)

// IPFirewallAddressListService handles communication with the IP firewall address-list
// methods of the RouterOS REST API.
//
// RouterOS API docs: https://manual.mikrotik.com/docs/Developer+Guides/rest-api
type IPFirewallAddressListService struct {
	c *Client
}

// Find retrieves firewall address list entries optionally filtered by list name and/or address.
//
// RouterOS API docs: https://manual.mikrotik.com/docs/Developer+Guides/rest-api
func (ipfwls *IPFirewallAddressListService) Find(ctx context.Context, list string, ips ...string) (types.FirewallAddressList, error) {
	var queries url.Values

	if len(list) > 0 {
		queries = make(url.Values)

		queries["list"] = []string{list}
	}

	if len(ips) > 0 {
		if queries == nil {
			queries = make(url.Values)
		}

		queries["address"] = ips
	}

	return makeRequest[types.FirewallAddressList](ctx, ipfwls.c, types.EndpointIPFirewallAddresList, http.MethodGet, nil, queries)
}

// Add adds a new entry to a firewall address list.
//
// RouterOS API docs: https://manual.mikrotik.com/docs/Developer+Guides/rest-api
func (ipfwls *IPFirewallAddressListService) Add(ctx context.Context, item types.FirewallAddressListNewItem) (*types.FirewallAddressListItem, error) {
	res, err := makeRequest[types.FirewallAddressListItem](ctx, ipfwls.c, types.EndpointIPFirewallAddresList, http.MethodPut, item, nil)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
