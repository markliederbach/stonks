package controller

// StockInfo tracks our position in the stock we are watching
type StockInfo struct {
	Symbol   string
	Position int64
}

// AccountInfo stores latest data about our alpaca account
type AccountInfo struct {
	ID               string
	Equity           float64
	MarginMultiplier float64
}

// OrderInfo tracks our current order status
type OrderInfo struct {
	ID string
}
