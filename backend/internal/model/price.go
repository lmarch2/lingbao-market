package model

type PriceItem struct {
	Code      string  `json:"code"`
	Price     float64 `json:"price"`
	Server    string  `json:"server,omitempty"`
	Timestamp int64   `json:"ts"`
}

type SubmitRequest struct {
	Code   string  `json:"code"`
	Price  float64 `json:"price"`
	Server string  `json:"server"`
}
