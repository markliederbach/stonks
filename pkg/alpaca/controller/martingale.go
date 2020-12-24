package controller

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/alpacahq/alpaca-trade-api-go/stream"
)

// MartingaleController implements the martingale system for tracking a stock
type MartingaleController struct {
	client        AlpacaClient
	tickSize      int
	tickIndex     int
	baseBet       float64
	currStreak    streak
	currOrder     string
	lastPrice     float64
	lastTradeTime time.Time
	stock         string
	position      int64
	equity        float64
	marginMult    float64
	seconds       int
}

type streak struct {
	start      float64
	count      int
	increasing bool
}

// NewMartingaleController returns a new MartingaleController after initializing with Alpaca
func NewMartingaleController(client AlpacaClient, stock string) (MartingaleController, error) {
	// Cancel any open orders so they don't interfere with this script
	if err := client.CancelAllOrders(); err != nil {
		return MartingaleController{}, err
	}

	var currentPosition int64
	position, err := client.GetPosition(stock)
	if err != nil {

		if err.Error() != "position does not exist" {
			return MartingaleController{}, err
		}

		// No position exists, set to zero
		currentPosition = 0
	} else {
		currentPosition = position.Qty.IntPart()
	}

	// Figure out how much money we have to work with, accounting for margin
	accountInfo, err := client.GetAccount()
	if err != nil {
		return MartingaleController{}, err
	}

	equity, _ := accountInfo.Equity.Float64()
	marginMult, err := strconv.ParseFloat(accountInfo.Multiplier, 64)
	if err != nil {
		return MartingaleController{}, err
	}

	totalBuyingPower := marginMult * equity
	logrus.WithFields(logrus.Fields{
		"initial_equity":       math.Round(equity*100) / 100,
		"initial_buying_power": math.Round(totalBuyingPower*100) / 100,
	}).Debugf("Calculated initial balances")

	return MartingaleController{
		client:    client,
		tickSize:  5,
		tickIndex: -1,
		baseBet:   0.1,
		currStreak: streak{
			start:      0,
			count:      0,
			increasing: true,
		},
		currOrder:     "",
		lastPrice:     0,
		lastTradeTime: time.Now().UTC(),
		stock:         stock,
		position:      currentPosition,
		equity:        equity,
		marginMult:    marginMult,
		seconds:       0, // TODO: what affect would this have?
	}, nil
}

// Run kicks off the main logic of this controller
func (c *MartingaleController) Run() error {

	// First, cancel any existing orders so they don't impact our buying power.
	status, until, limit := "open", time.Now(), 100
	orders, _ := c.client.ListOrders(&status, &until, &limit, nil)
	for _, order := range orders {
		_ = c.client.CancelOrder(order.ID)
	}

	// Register a handler for the stock stream we want to watch
	if err := stream.Register(fmt.Sprintf("T.%s", c.stock), c.handleStockEvent); err != nil {
		return err
	}

	if err := stream.Register("trade_updates", c.handleTradeEvent); err != nil {
		return err
	}

	// Sleep indefinitely while we wait for events
	select {}
}

// Listen for quote data and perform trading logic
func (c *MartingaleController) handleStockEvent(msg interface{}) {

}

// Listen for updates to our orders
func (c *MartingaleController) handleTradeEvent(msg interface{}) {

}
