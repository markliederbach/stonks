package main

import (
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/common"
	"github.com/markliederbach/stonks/pkg/alpaca/config"
	"github.com/sirupsen/logrus"
)

var (
	appConfig config.Config
)

func init() {
	appConfig = config.Load()
	alpaca.SetBaseUrl(appConfig.AlpacaAPIBaseURL)
}

func main() {
	logrus.Info("Alpaca trader is starting")
	client := alpaca.NewClient(&common.APIKey{
		ID:     appConfig.AlpacaAPIKeyID,
		Secret: appConfig.AlpacaAPISecretKey,
	})
}
