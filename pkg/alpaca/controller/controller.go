package controller

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/stream"
	"github.com/markliederbach/stonks/pkg/alpaca/api"
	"github.com/sirupsen/logrus"
)

// AlpacaController is the backbone of the system, which supports pluggable
// underlying algorithms.
type AlpacaController struct {
	client    api.AlpacaClient
	algorithm api.AlpacaAlgorithm
	stock     api.StockInfo
	account   api.AccountInfo
	order     api.OrderInfo
}

// NewAlpacaController returns an new controller.
func NewAlpacaController(client api.AlpacaClient, algorithm api.AlpacaAlgorithm, stock string) (AlpacaController, error) {
	// Cancel any open orders so they don't interfere with this script
	if err := client.CancelAllOrders(); err != nil {
		return AlpacaController{}, err
	}

	alpacaController := AlpacaController{
		client:    client,
		algorithm: algorithm,
		stock: api.StockInfo{
			Symbol:   stock,
			Position: 0,
		},
		account: api.AccountInfo{},
	}

	if err := alpacaController.UpdatePosition(); err != nil {
		return AlpacaController{}, err
	}

	if err := alpacaController.UpdateAccountBalance(); err != nil {
		return AlpacaController{}, err
	}

	logrus.WithFields(logrus.Fields{
		"stock":        alpacaController.stock.Symbol,
		"position":     alpacaController.stock.Position,
		"equity":       math.Round(alpacaController.account.Equity*100) / 100,
		"buying_power": math.Round(alpacaController.account.MarginMultiplier*alpacaController.account.Equity*100) / 100,
	}).Debugf("Loaded initial state")

	return alpacaController, nil
}

// UpdatePosition refreshes our current position for a stock
func (c *AlpacaController) UpdatePosition() error {
	var position int64 = 0
	stockPosition, err := c.client.GetPosition(c.stock.Symbol)
	if err != nil {
		if err.Error() != "position does not exist" {
			return err
		}
	} else {
		position = stockPosition.Qty.IntPart()
	}

	c.stock.Position = position

	return nil
}

// UpdateAccountBalance refreshes our available equity and margin from Alpaca
func (c *AlpacaController) UpdateAccountBalance() error {
	// Figure out how much money we have to work with, accounting for margin
	accountState, err := c.client.GetAccount()
	if err != nil {
		return err
	}

	equity, _ := accountState.Equity.Float64()
	marginMultiplier, err := strconv.ParseFloat(accountState.Multiplier, 64)
	if err != nil {
		return err
	}

	c.account.Equity = equity
	c.account.MarginMultiplier = marginMultiplier

	return nil
}

// Run kicks off the main logic of this controller
func (c *AlpacaController) Run() error {
	// Cancel any existing orders so they don't impact our buying power.
	status, until, limit := "open", time.Now(), 100
	orders, _ := c.client.ListOrders(&status, &until, &limit, nil)
	for _, order := range orders {
		logrus.Debugf("Cancelling pre-existing order %s", order.ID)
		if err := c.client.CancelOrder(order.ID); err != nil {
			return err
		}
	}

	// Register a handler for the stock stream we want to watch
	// https://alpaca.markets/docs/api-documentation/api-v2/market-data/streaming/
	dataStreamKey := fmt.Sprintf("T.%s", c.stock.Symbol)
	if err := stream.Register(dataStreamKey, c.handleStreamTrade); err != nil {
		return err
	}

	defer stream.Deregister(dataStreamKey)

	// Register a handler for updates to our existing trade orders
	if err := stream.Register(alpaca.TradeUpdates, c.handleTradeUpdate); err != nil {
		return err
	}

	defer stream.Deregister(alpaca.TradeUpdates)

	// Sleep indefinitely while we wait for events
	select {}
}

// Listen for quote data and perform trading logic
func (c *AlpacaController) handleStreamTrade(msg interface{}) {
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

	if data.Symbol != c.stock.Symbol {
		logrus.Infof("Ignoring stream trade event for unrelated stock %s", data.Symbol)
		return
	}

	c.algorithm.HandleStreamTrade(
		api.StreamTradeContext{
			Client:     c.client,
			Stock:      c.stock,
			Account:    c.account,
			Trade:      data,
			ContextLog: contextLog,
		},
	)

	if err := c.UpdateAccountBalance(); err != nil {
		logrus.Error(err)
		return
	}
}

// Listen for updates to our orders
func (c *AlpacaController) handleTradeUpdate(msg interface{}) {
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

	if data.Order.Symbol != c.stock.Symbol {
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
			"position": c.stock.Position,
		}).Info("Updated position")

		if data.Event == "fill" && c.order.ID == data.Order.ID {
			// Clear out completed order
			c.order.ID = ""
		}
	case "rejected", "canceled":
		if c.order.ID == data.Order.ID {
			// Clear out order
			c.order.ID = ""
		}
	case "new":
		c.order.ID = data.Order.ID
	default:
		contextLog.Error("Unexpected order event type")
	}
}
