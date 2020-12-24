package alpaca

import "time"

type Last struct {
	AskPrice float64 `json:"askprice"`
	AskSize uint64  `json:"asksize"`
	AskExchange string `json:"askexchange"`
	Bidprice float64 `json:"bidprice"`
	BidSize uint64 `json:"bidsize"`
	BidExchange string `json:"bidexchange"`
	Timestamp time.Time `json:"timestamp"`
}

type LastQuote struct {
	Status string `json:"status"`
	Symbol string `json:"symbol"`
	Last Last
}
