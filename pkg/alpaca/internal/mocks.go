package internal

import (
	"errors"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/markliederbach/stonks/pkg/alpaca/api"
	"github.com/shopspring/decimal"
)

var (
	objChs map[string]chan interface{} = make(map[string]chan interface{})
)

type void struct{}

func init() {
	functions := []string{
		"CancelAllOrders",
		"GetAccount",
		"GetPosition",
		"CancelOrder",
		"ListOrders",
		"PlaceOrder",
	}

	for _, functionName := range functions {
		// Create a channel to pass in return objects for a function
		objChs[functionName] = make(chan interface{}, 100)
	}
}

// getObj looks up the next object return from a channel, defaults if none exists
func getObj(functionName string) interface{} {
	select {
	case obj := <-objChs[functionName]:
		return obj
	default:
		return void{}
	}
}

// AddObjReturns allows a test to add one or more object returns for a function
func AddObjReturns(functionName string, objs ...interface{}) error {
	for _, obj := range objs {
		ch, ok := objChs[functionName]
		if !ok {
			return errors.New("Function channel does not exist")
		}
		ch <- obj
	}
	return nil
}

// MockAlpacaClient mocks the Alpaca SDK client
type MockAlpacaClient struct {
	api.AlpacaClient
}

// NewMockAlpacaClient returns a new mock algorithm
func NewMockAlpacaClient() api.AlpacaClient {
	return &MockAlpacaClient{}
}

// CancelAllOrders implements the corresponding function on api.AlpacaClient
func (mc *MockAlpacaClient) CancelAllOrders() error {
	funcitonName := "CancelAllOrders"
	obj := getObj(funcitonName)
	switch obj := obj.(type) {
	case error:
		return obj
	default:
		return nil
	}
}

// CancelOrder implements the corresponding function on api.AlpacaClient
func (mc *MockAlpacaClient) CancelOrder(orderID string) error {
	funcitonName := "CancelOrder"
	obj := getObj(funcitonName)
	switch obj := obj.(type) {
	case error:
		return obj
	default:
		return nil
	}
}

// GetAccount implements the corresponding function on api.AlpacaClient
func (mc *MockAlpacaClient) GetAccount() (*alpaca.Account, error) {
	funcitonName := "GetAccount"
	obj := getObj(funcitonName)
	switch obj := obj.(type) {
	case *alpaca.Account:
		return obj, nil
	case error:
		return &alpaca.Account{}, obj
	default:
		return &alpaca.Account{
			ID:         "account123",
			Equity:     decimal.NewFromFloat(1000),
			Multiplier: "2.00",
		}, nil
	}
}

// GetPosition implements the corresponding function on api.AlpacaClient
func (mc *MockAlpacaClient) GetPosition(string) (*alpaca.Position, error) {
	funcitonName := "GetPosition"
	obj := getObj(funcitonName)
	switch obj := obj.(type) {
	case *alpaca.Position:
		return obj, nil
	case error:
		return &alpaca.Position{}, obj
	default:
		return &alpaca.Position{
			Qty: decimal.NewFromFloat(3.5),
		}, nil
	}
}

// ListOrders implements the corresponding function on api.AlpacaClient
func (mc *MockAlpacaClient) ListOrders(status *string, until *time.Time, limit *int, nested *bool) ([]alpaca.Order, error) {
	funcitonName := "ListOrders"
	obj := getObj(funcitonName)
	switch obj := obj.(type) {
	case []alpaca.Order:
		return obj, nil
	case error:
		return []alpaca.Order{}, obj
	default:
		return []alpaca.Order{
			{ID: "foobar123"},
		}, nil
	}
}

// PlaceOrder implements the corresponding function on api.AlpacaClient
func (mc *MockAlpacaClient) PlaceOrder(req alpaca.PlaceOrderRequest) (*alpaca.Order, error) {
	funcitonName := "PlaceOrder"
	obj := getObj(funcitonName)
	switch obj := obj.(type) {
	case *alpaca.Order:
		return obj, nil
	case error:
		return &alpaca.Order{ID: "this should not be read"}, obj
	default:
		return &alpaca.Order{
			ID:          "order123",
			Symbol:      *req.AssetKey,
			Side:        req.Side,
			Type:        req.Type,
			Qty:         req.Qty,
			LimitPrice:  req.LimitPrice,
			TimeInForce: req.TimeInForce,
		}, nil
	}
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
