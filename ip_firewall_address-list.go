package routeros

import (
	"context"
	"net/http"
	"net/url"
)

type FirewallAddressListItem struct {
	ID           string   `json:".id"`
	Address      string   `json:"address"`
	CreationTime DateTime `json:"creation-time"`
	Disabled     string   `json:"disabled"`
	Dynamic      string   `json:"dynamic"`
	List         string   `json:"list"`
}

type FirewallAddressList []FirewallAddressListItem

type IPFirewallAddressListService struct {
	c *Client
}

func (ipfwls *IPFirewallAddressListService) Find(ctx context.Context, list string, ips ...string) (FirewallAddressList, error) {
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

	return makeRequest[FirewallAddressList](ctx, ipfwls.c, endpointIPFirewallAddresList, http.MethodGet, nil, queries)
}

type FirewallAddressListNewItem struct {
	Address  string `json:"address"`
	Comment  string `json:"comment"`
	Disabled string `json:"disabled"`
	Dynamic  string `json:"dynamic"`
	List     string `json:"list"`
}

func (ipfwls *IPFirewallAddressListService) Add(ctx context.Context, item FirewallAddressListNewItem) error {
	_, err := makeRequest[any](ctx, ipfwls.c, endpointIPFirewallAddresList, http.MethodPut, item, nil)

	return err
}
