package controller

// https://www.investopedia.com/articles/forex/06/martingale.asp

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/stream"
)

// MartingaleController implements the martingale system for tracking a stock
type MartingaleController struct {
	// client performs Alpaca actions.
	client AlpacaClient
	// tickSize is how many ticks to wait between transactions.
	tickSize int
	// tickIndex tracks where in the cycle we are between transactions.
	tickIndex int
	// baseBet is a percentage (0 - 1.0) representing
	// what percentage of our buying power to risk.
	baseBet float64
	// currStreak tracks the current streak from the latest tick.
	currStreak streak
	// currOrder is a unique string ID for the last order placed.
	currOrder string
	// lastPrice is the most recent stock price before the current tick price.
	lastPrice float64
	// lastTradeTime is the last time we transacted.
	lastTradeTime time.Time
	// stock is the symbol of the stock we are tracking.
	stock string
	// position is the number of shares we currently own of the stock.
	position int64
	// equity is the cash available in our account (before margin).
	equity float64
	// marginMult is the margin multiplier available for our account.
	marginMult float64
	// seconds TODO: what does this even do?
	seconds int
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
	// https://alpaca.markets/docs/api-documentation/api-v2/market-data/streaming/
	dataStreamKey := fmt.Sprintf("T.%s", c.stock)
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
func (c *MartingaleController) handleStreamTrade(msg interface{}) {
	data, ok := msg.(alpaca.StreamTrade)
	if !ok {
		logrus.Error("Failed to decode stream trade")
		return
	}

	if data.Symbol != c.stock {
		logrus.Debugf("Ignoring stream trade event for unrelated stock %s", data.Symbol)
		return
	}

	now := time.Now().UTC()
	if now.Sub(c.lastTradeTime) < time.Second {
		// don't react every tick unless at least 1 second past
		return
	}

	c.lastTradeTime = now

	// Increment or cycle the tick index
	c.tickIndex = (c.tickIndex + 1) % c.tickSize

	// Only process every n ticks
	if c.tickIndex == 0 {
		// It's time to update

		// Update price info
		previousPrice := c.lastPrice
		newPrice := float64(data.Price)
		c.lastPrice = newPrice

		c.processTick(previousPrice, newPrice)
	}
}

// Listen for updates to our orders
func (c *MartingaleController) handleTradeUpdate(msg interface{}) {

}

// processTick implements the algorithm
func (c *MartingaleController) processTick(previousPrice float64, newPrice float64) {
	logrus.WithFields(logrus.Fields{
		"logger":         "processTick",
		"previous_price": math.Round(previousPrice*100) / 100,
		"tick_price":     math.Round(newPrice*100) / 100,
	}).Info("Handling tick")

	// Only act on meaningful changes
	if math.Abs(newPrice-previousPrice) >= 0.01 {
		increasing := newPrice > previousPrice
		if c.currStreak.increasing != increasing {
			// It moved in the opposite direction of the tracked streak.
			// Therefore, the tracked streak is over, and we should reset.
			if c.position != 0 {
				// Empty out our position
				if _, err := c.sendOrder(0); err != nil {
					logrus.Panic(err)
				}
			}

			// Reset streak
			c.currStreak.increasing = increasing
			c.currStreak.start = newPrice
			c.currStreak.count = 0
		} else {
			// Our streak continues.
			c.currStreak.count++

			// Calculate our current buying power
			totalBuyingPower := c.equity * c.marginMult

			// Use our streak, buying power, and how much we originally bet to calculate
			// how much value we want to have in the stock.
			targetValue := math.Pow(2, float64(c.currStreak.count)) * c.baseBet * totalBuyingPower

			// Limit ourselves to roughly one share less than our total buying power.
			targetValue = math.Min(targetValue, totalBuyingPower-newPrice)

			// Calculate how many shares we need to transact
			targetQty := int(targetValue / newPrice)

			if increasing {
				// If the streak is increasing, that means we want to
				// sell our position
				targetQty = -targetQty
			}

		}
	}

	return
}
