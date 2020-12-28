package controller

import (
	"errors"
	"fmt"
	"math"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/stream"
	"github.com/markliederbach/stonks/pkg/alpaca/api"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// AlpacaController is the backbone of the system, which supports pluggable
// underlying algorithms.
type AlpacaController struct {
	Client    api.AlpacaClient
	Algorithm api.AlpacaAlgorithm
	Stock     api.StockInfo
	Account   api.AccountInfo
	Order     api.OrderInfo
}

// NewAlpacaController returns an new controller.
func NewAlpacaController(client api.AlpacaClient, algorithm api.AlpacaAlgorithm, stock string) (AlpacaController, error) {
	// Cancel any open orders so they don't interfere with this script
	if err := client.CancelAllOrders(); err != nil {
		return AlpacaController{}, err
	}

	alpacaController := AlpacaController{
		Client:    client,
		Algorithm: algorithm,
		Stock: api.StockInfo{
			Symbol:   stock,
			Position: 0,
		},
		Account: api.AccountInfo{},
	}

	if err := alpacaController.UpdatePosition(); err != nil {
		return AlpacaController{}, err
	}

	if err := alpacaController.UpdateAccount(); err != nil {
		return AlpacaController{}, err
	}

	logrus.WithFields(logrus.Fields{
		"stock":        alpacaController.Stock.Symbol,
		"position":     alpacaController.Stock.Position,
		"equity":       math.Round(alpacaController.Account.Equity*100) / 100,
		"buying_power": math.Round(alpacaController.Account.MarginMultiplier*alpacaController.Account.Equity*100) / 100,
	}).Debugf("Loaded initial state")

	return alpacaController, nil
}

// UpdatePosition refreshes our current position for a stock
func (c *AlpacaController) UpdatePosition() error {
	var position int64 = 0
	stockPosition, err := c.Client.GetPosition(c.Stock.Symbol)
	if err != nil {
		if err.Error() != "position does not exist" {
			return err
		}
	} else {
		position = stockPosition.Qty.IntPart()
	}

	c.Stock.Position = position

	return nil
}

// UpdateAccount refreshes our available equity and margin from Alpaca
func (c *AlpacaController) UpdateAccount() error {
	// Figure out how much money we have to work with, accounting for margin
	accountState, err := c.Client.GetAccount()
	if err != nil {
		return err
	}

	equity, _ := accountState.Equity.Float64()
	marginMultiplier, err := strconv.ParseFloat(accountState.Multiplier, 64)
	if err != nil {
		return err
	}

	c.Account.ID = accountState.ID
	c.Account.Equity = equity
	c.Account.MarginMultiplier = marginMultiplier

	return nil
}

// Run kicks off the main logic of this controller
func (c *AlpacaController) Run() error {
	// Cancel any existing orders so they don't impact our buying power.
	status, until, limit := "open", time.Now(), 100
	orders, _ := c.Client.ListOrders(&status, &until, &limit, nil)
	for _, order := range orders {
		logrus.Debugf("Cancelling pre-existing order %s", order.ID)
		if err := c.Client.CancelOrder(order.ID); err != nil {
			return err
		}
	}

	// Register a handler for the stock stream we want to watch
	// https://alpaca.markets/docs/api-documentation/api-v2/market-data/streaming/
	dataStreamKey := fmt.Sprintf("T.%s", c.Stock.Symbol)
	if err := stream.Register(dataStreamKey, c.handleStreamTrade); err != nil {
		return err
	}

	// Runs if this function ever returns
	defer stream.Deregister(dataStreamKey)

	// Register a handler for updates to our existing trade orders
	if err := stream.Register(alpaca.TradeUpdates, c.handleTradeUpdate); err != nil {
		return err
	}

	// Runs if this function ever returns
	defer stream.Deregister(alpaca.TradeUpdates)

	// Add SIGTERM handler
	c.setupInterruptHandler([]string{dataStreamKey, alpaca.TradeUpdates})

	// TODO: Uncomment to send a test order
	// time.Sleep(time.Second * 5)
	// orderID, err := c.SendLimitOrder(1, 192.82)
	// if err != nil {
	// 	return err
	// }
	// logrus.Infof("Created dummy order %s", orderID)

	// Sleep indefinitely while we wait for events
	select {}
}

// setupInterruptHandler catches CTRL-C interrupts and close streams
func (c *AlpacaController) setupInterruptHandler(streamKeys []string) {
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		logrus.WithFields(logrus.Fields{"stream_keys": streamKeys}).Info("Closing Alpaca data streams")
		for _, streamKey := range streamKeys {
			if err := stream.Deregister(streamKey); err != nil {
				logrus.WithFields(logrus.Fields{"stream_key": streamKey}).Warnf("Failed to deregister stream: %v", err)
			}
		}
		os.Exit(1)
	}()
}

// SendLimitOrder takes a position at which we want to have in the stock and makes it so,
// either by selling or buying shares.
func (c *AlpacaController) SendLimitOrder(targetPosition int, targetPrice float64) (string, error) {
	delta := math.Max(float64(targetPosition), 0) - math.Max(float64(c.Stock.Position), 0)

	var (
		side     alpaca.Side
		quantity float64 = math.Abs(delta)
	)

	if delta == 0 {
		// We are already at our target position
		return "", errors.New("no-op order requested")
	}

	if delta > 0 {
		// We need to buy more shares to reach our target position
		side = alpaca.Buy
	} else {
		// We need to sell shares to reach our target position
		side = alpaca.Sell
	}

	limitPrice := decimal.NewFromFloat(targetPrice)

	order, err := c.Client.PlaceOrder(alpaca.PlaceOrderRequest{
		AccountID:   c.Account.ID,
		AssetKey:    &c.Stock.Symbol,
		Qty:         decimal.NewFromFloat(quantity),
		Side:        side,
		Type:        alpaca.Limit,
		LimitPrice:  &limitPrice,
		TimeInForce: alpaca.Day,
	})

	if err != nil {
		return "", err
	}

	return order.ID, nil
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

	if data.Symbol != c.Stock.Symbol {
		logrus.Infof("Ignoring stream trade event for unrelated stock %s", data.Symbol)
		return
	}

	c.Algorithm.HandleStreamTrade(
		api.StreamTradeContext{
			Client:     c.Client,
			Stock:      c.Stock,
			Account:    c.Account,
			Trade:      data,
			ContextLog: contextLog,
		},
	)

	if err := c.UpdateAccount(); err != nil {
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

	if data.Order.Symbol != c.Stock.Symbol {
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
