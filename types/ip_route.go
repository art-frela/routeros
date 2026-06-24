package types

// IPRoute represents a single IP route entry from the RouterOS routing table.
//
// RouterOS API docs: https://manual.mikrotik.com/docs/Developer+Guides/rest-api
type IPRoute struct {
	ID            string `json:".id"`
	DstAddress    string `json:"dst-address"`
	Gateway       string `json:"gateway"`
	Distance      string `json:"distance"`
	Scope         string `json:"scope"`
	TargetScope   string `json:"target-scope"`
	RoutingTable  string `json:"routing-table,omitempty"`
	PrefSrc       string `json:"pref-src,omitempty"`
	GatewayStatus string `json:"gateway-status,omitempty"`
	Active        string `json:"active,omitempty"`
	Static        string `json:"static,omitempty"`
	Dynamic       string `json:"dynamic,omitempty"`
	Connect       string `json:"connect,omitempty"`
	Disabled      string `json:"disabled,omitempty"`
	Comment       string `json:"comment,omitempty"`
}

// IPRouteList is a slice of IPRoute entries.
//
// RouterOS API docs: https://manual.mikrotik.com/docs/Developer+Guides/rest-api
type IPRouteList []IPRoute

// IPRouteAdd represents the fields required to add a new static route.
//
// RouterOS API docs: https://manual.mikrotik.com/docs/Developer+Guides/rest-api
type IPRouteAdd struct {
	DstAddress   string `json:"dst-address"`
	Gateway      string `json:"gateway"`
	Distance     string `json:"distance,omitempty"`
	Scope        string `json:"scope,omitempty"`
	TargetScope  string `json:"target-scope,omitempty"`
	RoutingTable string `json:"routing-table,omitempty"`
	Disabled     string `json:"disabled,omitempty"`
	Comment      string `json:"comment,omitempty"`
}
