package algorithm

// https://www.investopedia.com/articles/forex/06/martingale.asp

import (
	"math"
	"time"

	"github.com/markliederbach/stonks/pkg/alpaca/interfaces"
	"github.com/sirupsen/logrus"
)

var _ interfaces.AlpacaAlgorithm = &Martingale{}

// Martingale implements the martingale system for tracking a stock
type Martingale struct {
	// tickSize is how many ticks to wait between transactions.
	tickSize int
	// tickIndex tracks where in the cycle we are between transactions.
	tickIndex int
	// lastPrice is the most recent stock price before the current tick price.
	lastPrice float64
	// lastTradeTime is the last time we transacted.
	lastTradeTime time.Time
}

// NewMartingale returns a new Martingale algorithm
func NewMartingale() (*Martingale, error) {
	return &Martingale{
		tickSize:      5,
		tickIndex:     -1,
		lastPrice:     0,
		lastTradeTime: time.Now().UTC(),
	}, nil
}

// HandleStreamTrade implements the function on the AlpacaAlgorithm interface
func (c *Martingale) HandleStreamTrade(context interfaces.StreamTradeContext) {
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
		newPrice := float64(context.Trade.Price)
		c.lastPrice = newPrice

		context.ContextLog.WithFields(logrus.Fields{
			"logger":         "algorithm_martingale",
			"previous_price": math.Round(previousPrice*100) / 100,
			"tick_price":     math.Round(newPrice*100) / 100,
		}).Info("Handling stream trade")
	}
}
