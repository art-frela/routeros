package types

// IPAddress represents a single IP address entry from RouterOS.
//
// RouterOS API docs: https://manual.mikrotik.com/docs/Developer+Guides/rest-api
type IPAddress struct {
	ID              string `json:".id"`
	Address         string `json:"address"`
	Network         string `json:"network"`
	Interface       string `json:"interface"`
	ActualInterface string `json:"actual-interface"`
	Dynamic         string `json:"dynamic"`
	Disabled        string `json:"disabled"`
}

// IPAddressList is a slice of IPAddress entries.
//
// RouterOS API docs: https://manual.mikrotik.com/docs/Developer+Guides/rest-api
type IPAddressList []IPAddress

// IPAddressAdd represents the fields required to add a new IP address.
//
// RouterOS API docs: https://manual.mikrotik.com/docs/Developer+Guides/rest-api
type IPAddressAdd struct {
	Address   string `json:"address"`
	Interface string `json:"interface"`
	Network   string `json:"network,omitempty"`
	Comment   string `json:"comment,omitempty"`
	Disabled  string `json:"disabled,omitempty"`
}
