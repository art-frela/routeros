package types

// EchoRequest represents the request parameters for the /tool/ping endpoint.
//
// RouterOS API docs: https://manual.mikrotik.com/docs/Developer+Guides/rest-api
type EchoRequest struct {
	Address  string  `json:"address"`
	Count    int64   `json:"count"`
	Interval float64 `json:"interval"`
}

// EchoResponse is a slice of EchoResponseElement representing ping results.
type EchoResponse []EchoResponseElement

// EchoResponseElement represents a single ping response element.
//
// RouterOS API docs: https://manual.mikrotik.com/docs/Developer+Guides/rest-api
type EchoResponseElement struct {
	Host       string  `json:"host"`
	PacketLoss string  `json:"packet-loss"`
	Received   string  `json:"received"`
	Sent       string  `json:"sent"`
	Seq        string  `json:"seq"`
	AvgRtt     *string `json:"avg-rtt,omitempty"`
	MaxRtt     *string `json:"max-rtt,omitempty"`
	MinRtt     *string `json:"min-rtt,omitempty"`
	Size       *string `json:"size,omitempty"`
	Time       *string `json:"time,omitempty"`
	TTL        *string `json:"ttl,omitempty"`
	Status     *string `json:"status,omitempty"`
}
