package bsv20

// Bsv20 represents a BSV20 token
type Bsv20 struct {
	Id       string  `json:"id,omitempty"`
	Op       string  `json:"op"`
	Ticker   string  `json:"tick,omitempty"`
	Decimals uint8   `json:"dec"`
	Icon     *string `json:"icon,omitempty"`
	Amt      uint64  `json:"amt"`
}
