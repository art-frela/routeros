package routeros

import (
	"context"
	"net/http"
	"net/url"

	"github.com/art-frela/routeros/types"
)

type IPFirewallAddressListService struct {
	c *Client
}

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

func (ipfwls *IPFirewallAddressListService) Add(ctx context.Context, item types.FirewallAddressListNewItem) (*types.FirewallAddressListItem, error) {
	res, err := makeRequest[types.FirewallAddressListItem](ctx, ipfwls.c, types.EndpointIPFirewallAddresList, http.MethodPut, item, nil)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
