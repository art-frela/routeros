package types

type EchoRequest struct {
	Address  string  `json:"address"`
	Count    int64   `json:"count"`
	Interval float64 `json:"interval"`
}

type EchoResponse []EchoResponseElement

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
