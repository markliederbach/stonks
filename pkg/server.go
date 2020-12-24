package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/markliederbach/stonks/pkg/clients/alpaca"
	"github.com/markliederbach/stonks/pkg/config"
	"github.com/sirupsen/logrus"
)

const (
	StockEndpoint string = "/stock/"
)

type Stock struct {
	URL string `json:"url"`
}

type StockHandler struct {
	endpoint string
	config   config.Config
}

func (h *StockHandler) handle(w http.ResponseWriter, r *http.Request) {
	logrus.Infof("Handling stock request")

	client, err := alpaca.NewAlpaca(h.config)
	if err != nil {
		logrus.Error(err, "Failed to create Alpaca client")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	symbol := strings.TrimPrefix(r.URL.Path, h.endpoint)
	if symbol == "" {
		logrus.Warn("No stock symbol provided for request")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	result, statusCode, err := client.LastQuote(symbol)
	if err != nil {
		logrus.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if statusCode != http.StatusOK {
		logrus.WithField("alpaca_status", statusCode).Error(err)
		http.Error(w, http.StatusText(statusCode), statusCode)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func main() {
	stockHandler := StockHandler{endpoint: StockEndpoint, config: config.Load()}

	portNumber := "9000"

	http.HandleFunc(stockHandler.endpoint, stockHandler.handle)

	logrus.WithFields(logrus.Fields{"port": portNumber}).Infof("Server Listening")
	logrus.Fatal(http.ListenAndServe(":"+portNumber, nil))
}
