package stream_controller

import (
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/markliederbach/stonks/pkg/controller"
	"github.com/markliederbach/stonks/pkg/platform"
	"github.com/markliederbach/stonks/pkg/strategy"
	"github.com/sirupsen/logrus"
)

type StreamController struct {
	Adaptor   platform.IPlatformAdaptor
	Algorithm strategy.Algorithm
	Stock     controller.StockInfo
	Account   controller.AccountInfo
	Order     controller.OrderInfo
}

func NewStreamController(platformAdaptor platform.IPlatformAdaptor, algorithm strategy.Algorithm, symbol string) (StreamController, error) {
	// Cancel existing orders for the account so they don't impact buying power.
	streamController := StreamController{
		Adaptor:   platformAdaptor,
		Algorithm: algorithm,
		Stock:     controller.StockInfo{Symbol: symbol, Position: 0},
		Account:   controller.AccountInfo{},
	}

	if err := streamController.CancelAllOrders(); err != nil {
		return StreamController{}, err
	}

	if err := streamController.UpdatePosition(); err != nil {
		return StreamController{}, err
	}

	if err := streamController.UpdateAccount(); err != nil {
		return StreamController{}, err
	}

	logrus.WithFields(logrus.Fields{
		"stock":        streamController.Stock.Symbol,
		"position":     streamController.Stock.Position,
		"equity":       math.Round(streamController.Account.Equity*100) / 100,
		"buying_power": math.Round(streamController.Account.MarginMultiplier*streamController.Account.Equity*100) / 100,
	}).Info("Initial stream controller state")

	return streamController, nil
}

// UpdateAccount refreshes our available equity and margin
func (c *StreamController) UpdateAccount() error {

	accountState, err := c.Adaptor.GetAccount()
	if err != nil {
		return err
	}

	c.Account = controller.AccountInfo{
		ID:               accountState.GetAccountID(),
		Equity:           accountState.GetEquity(),
		MarginMultiplier: accountState.GetMarginMultiplier(),
	}

	return nil
}

// UpdatePosition refreshes our current position for a stock
func (c *StreamController) UpdatePosition() error {

	positionState, err := c.Adaptor.GetPosition(c.Stock.Symbol)
	if err != nil {
		return err
	}

	c.Stock = controller.StockInfo{
		Symbol:   positionState.GetSymbol(),
		Position: positionState.GetQuantity(),
	}

	return nil
}

func (c *StreamController) CancelAllOrders() error {
	orders, err := c.Adaptor.ListOrders(platform.ListOrdersInput{
		OrderStatus:  platform.OrderOpen,
		Until:        time.Now(),
		LimitResults: 100,
	})
	if err != nil {
		return err
	}

	for _, order := range orders {
		logrus.Debugf("Canceling order %s", order.GetID())
		if err := c.Adaptor.CancelOrder(order.GetID()); err != nil {
			return err
		}
	}
	return nil
}

func (c *StreamController) Run() error {
	if err := c.CancelAllOrders(); err != nil {
		return err
	}

	// Register handlers for streams
	if err := c.Adaptor.RegisterStreams(c.Stock.Symbol); err != nil {
		return err
	}

	// Setup handlers to close streams
	defer c.deferDeregister()
	c.setupInterruptHandler()

	// Sleep indefinitely while we wait for events
	select {}
}

// setupInterruptHandler catches CTRL-C interrupts and close streams
func (c *StreamController) setupInterruptHandler() {
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		if err := c.Adaptor.DeregisterStreams(c.Stock.Symbol); err != nil {
			logrus.Error(err)
		}
		os.Exit(1)
	}()
}

func (c *StreamController) deferDeregister() {
	if err := c.Adaptor.DeregisterStreams(c.Stock.Symbol); err != nil {
		logrus.Error(err)
	}
}
