package main

import (
	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/common"
	"github.com/markliederbach/stonks/pkg/alpaca/algorithm"
	"github.com/markliederbach/stonks/pkg/alpaca/config"
	"github.com/markliederbach/stonks/pkg/alpaca/controller"
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

	stock := "VTI"

	client := alpaca.NewClient(&common.APIKey{
		ID:     appConfig.AlpacaAPIKeyID,
		Secret: appConfig.AlpacaAPISecretKey,
	})

	martingale, err := algorithm.NewMartingale()
	if err != nil {
		logrus.Panic(err)
	}

	alpacaController, err := controller.NewAlpacaController(client, martingale, stock)
	if err != nil {
		logrus.Panic(err)
	}

	// Does not return unless an error occurred
	if err := alpacaController.Run(); err != nil {
		logrus.Panic(err)
	}

}
