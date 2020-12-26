package controller

import (
	"math"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// AlpacaController is the backbone of the system, which supports pluggable
// underlying algorithms.
type AlpacaController struct {
	client    AlpacaClient
	algorithm AlpacaAlgorithm
	stock     StockInfo
	account   AccountInfo
}

// AccountInfo stores latest data about our alpaca account
type AccountInfo struct {
	equity           float64
	marginMultiplier float64
}

// StockInfo tracks our position in the stock we are watching
type StockInfo struct {
	symbol        string
	position      int64
	lastPrice     float64
	lastTradeTime time.Time
}

// NewAlpacaController returns an new controller.
func NewAlpacaController(client AlpacaClient, algorithm AlpacaAlgorithm, stock string) (AlpacaController, error) {
	// Cancel any open orders so they don't interfere with this script
	if err := client.CancelAllOrders(); err != nil {
		return AlpacaController{}, err
	}

	// Get our current position for the stock
	var position int64 = 0
	stockPosition, err := client.GetPosition(stock)
	if err != nil {
		if err.Error() != "position does not exist" {
			return AlpacaController{}, err
		}
	} else {
		position = stockPosition.Qty.IntPart()
	}

	// Figure out how much money we have to work with, accounting for margin
	accountState, err := client.GetAccount()
	if err != nil {
		return AlpacaController{}, err
	}

	equity, _ := accountState.Equity.Float64()
	marginMultiplier, err := strconv.ParseFloat(accountState.Multiplier, 64)
	if err != nil {
		return AlpacaController{}, err
	}

	logrus.WithFields(logrus.Fields{
		"stock":        stock,
		"position":     position,
		"equity":       math.Round(equity*100) / 100,
		"buying_power": math.Round(marginMultiplier*equity*100) / 100,
	}).Debugf("Loaded initial state")

	return AlpacaController{
		client:    client,
		algorithm: algorithm,
		stock: StockInfo{
			symbol:        stock,
			position:      position,
			lastPrice:     -1,
			lastTradeTime: time.Now().UTC(),
		},
		account: AccountInfo{
			equity:           equity,
			marginMultiplier: marginMultiplier,
		},
	}, nil
}

// Run kicks off the main logic of this controller
func (c *AlpacaController) Run() error {
		// First, cancel any existing orders so they don't impact our buying power.
		status, until, limit := "open", time.Now(), 100
		orders, _ := c.client.ListOrders(&status, &until, &limit, nil)
		for _, order := range orders {
			_ = c.client.CancelOrder(order.ID)
		}
		
}