package internal

import (
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/markliederbach/stonks/pkg/alpaca/api"
)

// MockAlpacaClient mocks the Alpaca SDK client
type MockAlpacaClient struct {
	Error    error
	Account  alpaca.Account
	Position alpaca.Position
	Orders   []alpaca.Order
	Order    alpaca.Order
}

// MockAlpacaClientInput provides optional fields to set on the mock client
type MockAlpacaClientInput struct {
	Error    error
	Account  alpaca.Account
	Position alpaca.Position
	Orders   []alpaca.Order
	Order    alpaca.Order
}

// NewMockAlpacaClient returns a new mock algorithm
func NewMockAlpacaClient(input MockAlpacaClientInput) api.AlpacaClient {
	if input.Orders == nil {
		input.Orders = []alpaca.Order{}
	}
	return &MockAlpacaClient{
		Error:   input.Error,
		Account: input.Account,
		Orders:  input.Orders,
		Order:   input.Order,
	}
}

// CancelAllOrders implements the corresponding function on api.AlpacaClient
func (mc *MockAlpacaClient) CancelAllOrders() error {
	if mc.Error != nil {
		return mc.Error
	}
	return nil
}

// CancelOrder implements the corresponding function on api.AlpacaClient
func (mc *MockAlpacaClient) CancelOrder(orderID string) error {
	if mc.Error != nil {
		return mc.Error
	}
	return nil
}

// GetAccount implements the corresponding function on api.AlpacaClient
func (mc *MockAlpacaClient) GetAccount() (*alpaca.Account, error) {
	if mc.Error != nil {
		return &alpaca.Account{}, mc.Error
	}
	return &mc.Account, nil
}

// GetPosition implements the corresponding function on api.AlpacaClient
func (mc *MockAlpacaClient) GetPosition(string) (*alpaca.Position, error) {
	if mc.Error != nil {
		return &alpaca.Position{}, mc.Error
	}
	return &mc.Position, nil
}

// ListOrders implements the corresponding function on api.AlpacaClient
func (mc *MockAlpacaClient) ListOrders(status *string, until *time.Time, limit *int, nested *bool) ([]alpaca.Order, error) {
	if mc.Error != nil {
		return []alpaca.Order{}, mc.Error
	}
	return mc.Orders, nil
}

// PlaceOrder implements the corresponding function on api.AlpacaClient
func (mc *MockAlpacaClient) PlaceOrder(req alpaca.PlaceOrderRequest) (*alpaca.Order, error) {
	if mc.Error != nil {
		return &alpaca.Order{}, mc.Error
	}
	return &mc.Order, nil
}

// MockAlgorithm mocks an algorithm handler and tracks
// how many times it was called
type MockAlgorithm struct {
	HandleStreamTradeCalled int
}

// NewMockAlgorithm returns a new mock algorithm
func NewMockAlgorithm() api.AlpacaAlgorithm {
	return &MockAlgorithm{HandleStreamTradeCalled: 0}
}

// HandleStreamTrade implements the function
func (ma *MockAlgorithm) HandleStreamTrade(context api.StreamTradeContext) {
	ma.HandleStreamTradeCalled++
}
