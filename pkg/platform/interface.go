package platform

import (
	"time"

	"github.com/markliederbach/stonks/pkg/strategy"
	"github.com/shopspring/decimal"
)

type IPlatformAdaptor interface {
	GetAccount() (PlatformAccount, error)
	GetPosition(stock string) (PlatformPosition, error)
	CancelOrder(orderID string) error
	ListOrders(input ListOrdersInput) ([]PlatformOrder, error)
	PlaceOrder(input PlaceOrderInput) (PlatformOrder, error)
	RegisterStreams(stock string, algorithm strategy.Algorithm) error
	DeregisterStreams(stock string) error
}

type ListOrdersInput struct {
	OrderStatus  OrderStatus
	Until        time.Time
	LimitResults int
}

type OrderStatus string

const (
	OrderOpen OrderStatus = "open"
)

type OrderType string

const (
	Market       OrderType = "market"
	Limit        OrderType = "limit"
	Stop         OrderType = "stop"
	StopLimit    OrderType = "stop_limit"
	TrailingStop OrderType = "trailing_stop"
)

type Side string

const (
	Buy  Side = "buy"
	Sell Side = "sell"
)

type PlaceOrderInput struct {
	AccountID   string
	Symbol      string
	Side        string
	OrderType   string
	Quantity    decimal.Decimal
	TimeInForce string
	LimitPrice  decimal.Decimal
	StopPrice   decimal.Decimal
}

type PlatformOrder interface {
	GetID() string
	GetStatus() OrderStatus
	GetOrderType() OrderType
	GetSide() Side
	GetSymbol() string
	GetQuantity() decimal.Decimal
	GetLimitPrice() decimal.Decimal
	GetFilledAvgPrice() decimal.Decimal
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetSubmittedAt() time.Time
	GetFilledAt() time.Time
	GetExpiredAt() time.Time
	GetCanceledAt() time.Time
}

type PlatformPosition interface {
	GetSymbol() string
	GetQuantity() int64
	GetEntryPrice() decimal.Decimal
	GetCurrentPrice() decimal.Decimal
	GetMarketValue() decimal.Decimal
}

type PlatformAccount interface {
	GetAccountID() string
	GetEquity() float64
	GetMarginMultiplier() (float64, error)
}
