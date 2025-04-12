package types

import "slices"

type FirewallAddressListItem struct {
	ID           string   `json:".id"`
	Address      string   `json:"address"`
	CreationTime DateTime `json:"creation-time"`
	Disabled     string   `json:"disabled"`
	Dynamic      string   `json:"dynamic"`
	List         string   `json:"list"`
}

type FirewallAddressList []FirewallAddressListItem

func (list FirewallAddressList) Find(address string) FirewallAddressList {
	if address == "" || address == "*" {
		return list
	}

	ix := slices.IndexFunc(list, func(item FirewallAddressListItem) bool {
		return item.Address == address
	})
	if ix < 0 {
		return FirewallAddressList{}
	}

	return FirewallAddressList{list[ix]}
}

type FirewallAddressListNewItem struct {
	Address  string `json:"address"`
	Comment  string `json:"comment"`
	Disabled string `json:"disabled"`
	Dynamic  string `json:"dynamic"`
	List     string `json:"list"`
}
