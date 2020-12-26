package controller

import (
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
)

// AlpacaClient wraps alpaca.Client to allow easy swap-out (such as for testing)
type AlpacaClient interface {
	CancelAllOrders() error
	GetAccount() (*alpaca.Account, error)
	GetPosition(string) (*alpaca.Position, error)
	CancelOrder(orderID string) error
	ListOrders(status *string, until *time.Time, limit *int, nested *bool) ([]alpaca.Order, error)
}

// AlpacaAlgorithm defines a contract for any implementing
// algorithm strategy to use with out Alpaca controller.
// The underlying assumption is that all algorithms will base
// their actions on a set of stream trades.
type AlpacaAlgorithm interface {
	// Given a stream trade, perform some action based on the data.
	HandleStreamTrade(trade alpaca.StreamTrade)
}
