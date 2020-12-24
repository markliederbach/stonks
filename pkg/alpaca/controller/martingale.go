package controller

import (
	"strconv"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/sirupsen/logrus"
)

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

func NewMartingaleController(client *alpaca.Client, stock string) (MartingaleController, error) {
	// Cancel any open orders so they don't interfere with this script
	if err := client.CancelAllOrders(); err != nil {
		return MartingaleController{}, err
	}

	var currentPosition int64
	position, err := client.GetPosition(stock)
	if err != nil {
		// No position exists
		// currentPosition = 0
		return MartingaleController{}, err
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
	logrus.Infof("Initial total buying power is %.2f", totalBuyingPower)
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
