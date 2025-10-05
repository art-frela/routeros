package types

type IPAddress struct {
	ID              string `json:".id"`
	Address         string `json:"address"`
	Network         string `json:"network"`
	Interface       string `json:"interface"`
	ActualInterface string `json:"actual-interface"`
	Dynamic         string `json:"dynamic"`
	Disabled        string `json:"disabled"`
}

type IPAddressList []IPAddress

type IPAddressAdd struct {
	Address   string `json:"address"`
	Interface string `json:"interface"`
	Network   string `json:"network,omitempty"`
	Comment   string `json:"comment,omitempty"`
	Disabled  string `json:"disabled,omitempty"`
}
