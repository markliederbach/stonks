package strategy

import (
	"github.com/markliederbach/stonks/pkg/platform"
	"github.com/sirupsen/logrus"
)

// Algorithm defines a contract for any implementing
// algorithm strategy to use with our Alpaca controller.
// The underlying assumption is that all algorithms will base
// their actions on a set of stream trades.
type Algorithm interface {
	// Given a stream trade, perform some action based on the data.
	HandleStreamTrade(context StreamTradeContext)
}

type StreamTradeContext struct {
	Client     platform.IPlatformAdaptor
	Stock      platform.PlatformPosition
	Account    platform.PlatformAccount
	Trade      interface{}
	ContextLog *logrus.Entry
}
