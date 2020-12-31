package alpaca_adaptor

import (
	"fmt"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/stream"
	"github.com/markliederbach/stonks/pkg/platform"
	"github.com/markliederbach/stonks/pkg/platform/alpaca_adaptor/api"
	"github.com/markliederbach/stonks/pkg/strategy"
	"github.com/sirupsen/logrus"
)

var _ platform.IPlatformAdaptor = &AlpacaAdaptor{}

type AlpacaAdaptor struct {
	Client api.AlpacaClient
}

func (a *AlpacaAdaptor) GetAccount() (platform.PlatformAccount, error) {
	account, err := a.Client.GetAccount()
	if err != nil {
		return &api.AlpacaAccount{}, err
	}
	return &api.AlpacaAccount{Account: account}, nil
}

func (a *AlpacaAdaptor) GetPosition(stock string) (platform.PlatformPosition, error) {
	position, err := a.Client.GetPosition(stock)
	if err != nil {
		return &api.AlpacaPosition{}, err
	}
	return &api.AlpacaPosition{Position: position}, nil
}

func (a *AlpacaAdaptor) CancelOrder(orderID string) error {
	if err := a.Client.CancelOrder(orderID); err != nil {
		return err
	}
	return nil
}

func (a *AlpacaAdaptor) ListOrders(input platform.ListOrdersInput) ([]platform.PlatformOrder, error) {
	alpacaOrders, err := a.Client.ListOrders((*string)(&input.OrderStatus), &input.Until, &input.LimitResults, nil)
	if err != nil {
		return []platform.PlatformOrder{}, err
	}

	// https://stackoverflow.com/questions/12994679/slice-of-struct-slice-of-interface-it-implements
	orders := make([]platform.PlatformOrder, len(alpacaOrders))
	for index, order := range alpacaOrders {
		orders[index] = &api.AlpacaOrder{Order: &order}
	}

	return orders, nil
}

func (a *AlpacaAdaptor) PlaceOrder(input platform.PlaceOrderInput) (platform.PlatformOrder, error) {
	order, err := a.Client.PlaceOrder(alpaca.PlaceOrderRequest{
		AccountID:   input.AccountID,
		AssetKey:    &input.Symbol,
		Qty:         input.Quantity,
		Side:        alpaca.Side(input.Side),
		Type:        alpaca.OrderType(input.OrderType),
		LimitPrice:  &input.LimitPrice,
		TimeInForce: alpaca.TimeInForce(input.TimeInForce),
	})

	if err != nil {
		return &api.AlpacaOrder{}, err
	}

	return &api.AlpacaOrder{Order: order}, nil
}

func (a *AlpacaAdaptor) RegisterStreams(stock string, algorithm strategy.Algorithm) error {
	// Register a handler for the stock stream we want to watch
	// https://alpaca.markets/docs/api-documentation/api-v2/market-data/streaming/
	dataStreamKey := fmt.Sprintf("T.%s", stock)
	if err := stream.Register(dataStreamKey, a.streamTradeHandler(stock, algorithm)); err != nil {
		return err
	}

	// Register a handler for updates to our existing trade orders
	if err := stream.Register(alpaca.TradeUpdates, a.tradeUpdateHandler(stock)); err != nil {
		return err
	}

	return nil
}

func (a *AlpacaAdaptor) DeregisterStreams(stock string) error {

}

// Listen for quote data and perform trading logic
func (a *AlpacaAdaptor) streamTradeHandler(stock string, algorithm strategy.Algorithm) func(msg interface{}) {
	
	return func(msg interface{}) {
		data, ok := msg.(alpaca.StreamTrade)
		if !ok {
			logrus.Error("Failed to decode stream trade")
			return
		}
	
		contextLog := logrus.WithFields(logrus.Fields{
			"symbol": data.Symbol,
			"price":  data.Price,
		})
	
		contextLog.Info("Handling stream trade event")
	
		if data.Symbol != stock {
			logrus.Infof("Ignoring stream trade event for unrelated stock %s", data.Symbol)
			return
		}

		positionState, err := a.GetPosition(stock)
		if err != nil {
			logrus.Error(err)
			return
		}

		accountState, err := a.GetAccount()
		if err != nil {
			logrus.Error(err)
				return
		}
		
		algorithm.HandleStreamTrade(
			strategy.StreamTradeContext{
				Client:     a,
				Stock:      positionState,
				Account:    accountState,
				Trade:      data,
				ContextLog: contextLog,
			},
		)
	}
}

// TODO: How to handle account, position, order? These should probably be stored
// on the adaptor, and not the controller
func (a *AlpacaAdaptor) tradeUpdateHandler(stock string) func(msg interface{}) {
	return func(msg interface{}) {
		data, ok := msg.(alpaca.TradeUpdate)
		if !ok {
			logrus.Error("Failed to decode trade update")
			return
		}
	
		contextLog := logrus.WithFields(logrus.Fields{
			"event":    data.Event,
			"order_id": data.Order.ID,
		})
	
		contextLog.Info("Handling trade update")
	
		if data.Order.Symbol != stock {
			logrus.Infof("Ignoring trade update for unrelated stock %s", data.Order.Symbol)
			return
		}
	
		switch data.Event {
		case "fill", "partial_fill":
			// Our position has changed
			if err := c.UpdatePosition(); err != nil {
				logrus.Error(err)
				return
			}
			contextLog.WithFields(logrus.Fields{
				"symbol":   data.Order.Symbol,
				"position": c.Stock.Position,
			}).Info("Updated position")
	
			if data.Event == "fill" && c.Order.ID == data.Order.ID {
				// Clear out completed order
				c.Order.ID = ""
			}
		case "rejected", "canceled":
			if c.Order.ID == data.Order.ID {
				// Clear out order
				c.Order.ID = ""
			}
		case "new":
			c.Order.ID = data.Order.ID
		default:
			contextLog.Error("Unexpected order event type")
		}
		contextLog.Info("Completed trade update")
		}
}

