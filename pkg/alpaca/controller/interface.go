package controller

import "github.com/alpacahq/alpaca-trade-api-go/alpaca"

type AlpacaClient interface {
	CancelAllOrders() error
	GetPosition(string) (*alpaca.Position, error)
}
