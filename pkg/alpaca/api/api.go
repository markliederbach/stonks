package api

import (
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/sirupsen/logrus"
)

// AlpacaClient wraps alpaca.Client to allow easy swap-out (such as for testing)
type AlpacaClient interface {
	CancelAllOrders() error
	GetAccount() (*alpaca.Account, error)
	GetPosition(string) (*alpaca.Position, error)
	CancelOrder(orderID string) error
	ListOrders(status *string, until *time.Time, limit *int, nested *bool) ([]alpaca.Order, error)
	PlaceOrder(req alpaca.PlaceOrderRequest) (*alpaca.Order, error)
}

// AlpacaAlgorithm defines a contract for any implementing
// algorithm strategy to use with our Alpaca controller.
// The underlying assumption is that all algorithms will base
// their actions on a set of stream trades.
type AlpacaAlgorithm interface {
	// Given a stream trade, perform some action based on the data.
	HandleStreamTrade(context StreamTradeContext)
}

// StreamTradeContext encapsulates context that is passed from
// a controller to the implementing algorithm
type StreamTradeContext struct {
	Client     AlpacaClient
	Stock      StockInfo
	Account    AccountInfo
	Trade      alpaca.StreamTrade
	ContextLog *logrus.Entry
}

// AccountInfo stores latest data about our alpaca account
type AccountInfo struct {
	ID               string
	Equity           float64
	MarginMultiplier float64
}

// StockInfo tracks our position in the stock we are watching
type StockInfo struct {
	Symbol   string
	Position int64
}

// OrderInfo tracks our current order status
type OrderInfo struct {
	ID string
}
