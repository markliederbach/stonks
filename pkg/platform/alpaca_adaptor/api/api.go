package api

import (
	"strconv"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/markliederbach/stonks/pkg/platform"
	"github.com/shopspring/decimal"
)

var _ platform.PlatformAccount = &AlpacaAccount{}
var _ platform.PlatformPosition = &AlpacaPosition{}
var _ platform.PlatformOrder = &AlpacaOrder{}

// AlpacaClient wraps alpaca.Client to allow easy swap-out (such as for testing)
type AlpacaClient interface {
	GetAccount() (*alpaca.Account, error)
	GetPosition(string) (*alpaca.Position, error)
	CancelOrder(orderID string) error
	ListOrders(status *string, until *time.Time, limit *int, nested *bool) ([]alpaca.Order, error)
	PlaceOrder(req alpaca.PlaceOrderRequest) (*alpaca.Order, error)
}

type AlpacaAccount struct {
	Account *alpaca.Account
}

func (a *AlpacaAccount) GetAccountID() string {
	return a.Account.ID
}

func (a *AlpacaAccount) GetEquity() float64 {
	equity, _ := a.Account.Equity.Float64()
	return equity
}

func (a *AlpacaAccount) GetMarginMultiplier() (float64, error) {
	marginMultiplier, err := strconv.ParseFloat(a.Account.Multiplier, 64)
	if err != nil {
		return 0, err
	}
	return marginMultiplier, nil
}

type AlpacaPosition struct {
	Position *alpaca.Position
}

func (p *AlpacaPosition) GetSymbol() string {
	return p.Position.Symbol
}

func (p *AlpacaPosition) GetQuantity() int64 {
	return p.Position.Qty.IntPart()
}

func (p *AlpacaPosition) GetEntryPrice() decimal.Decimal {
	return p.Position.EntryPrice
}

func (p *AlpacaPosition) GetCurrentPrice() decimal.Decimal {
	return p.Position.CurrentPrice
}

func (p *AlpacaPosition) GetMarketValue() decimal.Decimal {
	return p.Position.MarketValue
}

type AlpacaOrder struct {
	Order *alpaca.Order
}

func (o *AlpacaOrder) GetID() string {
	return o.Order.ID
}

func (o *AlpacaOrder) GetStatus() platform.OrderStatus {
	return platform.OrderStatus(o.Order.Status)
}

func (o *AlpacaOrder) GetOrderType() platform.OrderType {
	return platform.OrderType(o.Order.Type)
}

func (o *AlpacaOrder) GetSide() platform.Side {
	return platform.Side(o.Order.Side)
}

func (o *AlpacaOrder) GetSymbol() string {
	return o.Order.Symbol
}

func (o *AlpacaOrder) GetQuantity() decimal.Decimal {
	return o.Order.Qty
}

func (o *AlpacaOrder) GetLimitPrice() decimal.Decimal {
	return *o.Order.LimitPrice
}

func (o *AlpacaOrder) GetFilledAvgPrice() decimal.Decimal {
	return *o.Order.FilledAvgPrice
}

func (o *AlpacaOrder) GetCreatedAt() time.Time {
	return o.Order.CreatedAt
}

func (o *AlpacaOrder) GetUpdatedAt() time.Time {
	return o.Order.UpdatedAt
}

func (o *AlpacaOrder) GetSubmittedAt() time.Time {
	return o.Order.SubmittedAt
}

func (o *AlpacaOrder) GetFilledAt() time.Time {
	return *o.Order.FilledAt
}

func (o *AlpacaOrder) GetExpiredAt() time.Time {
	return *o.Order.ExpiredAt
}

func (o *AlpacaOrder) GetCanceledAt() time.Time {
	return *o.Order.CanceledAt
}
