package internal

import (
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/markliederbach/stonks/pkg/alpaca/api"
)

// MockAlpacaClient mocks the Alpaca SDK client
type MockAlpacaClient struct {
	Errors         []error
	Account        alpaca.Account
	Position       alpaca.Position
	ListOrderStack [][]alpaca.Order
	OrderStack     []alpaca.Order
}

// MockAlpacaClientInput provides optional fields to set on the mock client
type MockAlpacaClientInput struct {
	Errors         []error
	Account        alpaca.Account
	Position       alpaca.Position
	ListOrderStack [][]alpaca.Order
	OrderStack     []alpaca.Order
}

// NewMockAlpacaClient returns a new mock algorithm
func NewMockAlpacaClient(input MockAlpacaClientInput) api.AlpacaClient {
	return &MockAlpacaClient{
		Errors:         input.Errors,
		Account:        input.Account,
		Position:       input.Position,
		ListOrderStack: input.ListOrderStack,
		OrderStack:     input.OrderStack,
	}
}

// CancelAllOrders implements the corresponding function on api.AlpacaClient
func (mc *MockAlpacaClient) CancelAllOrders() error {
	if err := mc.popError(); err != nil {
		return err
	}
	return nil
}

// CancelOrder implements the corresponding function on api.AlpacaClient
func (mc *MockAlpacaClient) CancelOrder(orderID string) error {
	if err := mc.popError(); err != nil {
		return err
	}
	return nil
}

// GetAccount implements the corresponding function on api.AlpacaClient
func (mc *MockAlpacaClient) GetAccount() (*alpaca.Account, error) {
	if err := mc.popError(); err != nil {
		return &alpaca.Account{}, err
	}
	return &mc.Account, nil
}

// GetPosition implements the corresponding function on api.AlpacaClient
func (mc *MockAlpacaClient) GetPosition(string) (*alpaca.Position, error) {
	if err := mc.popError(); err != nil {
		return &alpaca.Position{}, err
	}
	return &mc.Position, nil
}

// ListOrders implements the corresponding function on api.AlpacaClient
func (mc *MockAlpacaClient) ListOrders(status *string, until *time.Time, limit *int, nested *bool) ([]alpaca.Order, error) {
	if err := mc.popError(); err != nil {
		return []alpaca.Order{}, err
	}
	return mc.popListOrders(), nil
}

// PlaceOrder implements the corresponding function on api.AlpacaClient
func (mc *MockAlpacaClient) PlaceOrder(req alpaca.PlaceOrderRequest) (*alpaca.Order, error) {
	if err := mc.popError(); err != nil {
		return &alpaca.Order{}, err
	}
	return mc.popOrder(), nil
}

func (mc *MockAlpacaClient) popError() error {
	if mc.Errors == nil || len(mc.Errors) == 0 {
		return nil
	}

	nextError := mc.Errors[0]
	mc.Errors = mc.Errors[1:]

	return nextError
}

func (mc *MockAlpacaClient) popOrder() *alpaca.Order {
	if mc.OrderStack == nil || len(mc.OrderStack) == 0 {
		return &alpaca.Order{}
	}

	nextOrder := mc.OrderStack[0]
	mc.OrderStack = mc.OrderStack[1:]

	return &nextOrder
}

func (mc *MockAlpacaClient) popListOrders() []alpaca.Order {
	if mc.ListOrderStack == nil || len(mc.ListOrderStack) == 0 {
		return []alpaca.Order{}
	}

	nextListOrder := mc.ListOrderStack[0]
	mc.ListOrderStack = mc.ListOrderStack[1:]

	return nextListOrder
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
