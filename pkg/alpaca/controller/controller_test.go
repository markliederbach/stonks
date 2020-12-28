package controller_test

import (
	"errors"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/markliederbach/stonks/pkg/alpaca/api"
	"github.com/markliederbach/stonks/pkg/alpaca/controller"
	"github.com/markliederbach/stonks/pkg/alpaca/internal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"
)

var _ = Describe("Controller", func() {
	var (
		alpacaController controller.AlpacaController
		mockClient       api.AlpacaClient
		mockAlgorithm    api.AlpacaAlgorithm
		stock            string = "MKL"
		err              error
	)

	Context("when creating a default controller", func() {
		JustBeforeEach(func() {
			mockClient = internal.NewMockAlpacaClient()
			mockAlgorithm = internal.NewMockAlgorithm()
			alpacaController, err = controller.NewAlpacaController(mockClient, mockAlgorithm, stock)
		})

		It("should not have failed", func() {
			Expect(err).ToNot(HaveOccurred())
		})
		It("should have initial account and position stats", func() {
			Expect(alpacaController.Account).To(Equal(api.AccountInfo{
				ID:               "account123",
				Equity:           float64(1000),
				MarginMultiplier: float64(2.00),
			}))
			Expect(alpacaController.Stock).To(Equal(api.StockInfo{
				Symbol:   stock,
				Position: 3,
			}))
		})

		Context("when no existing positions are found", func() {
			BeforeEach(func() {
				err := internal.AddObjReturns("GetPosition", errors.New("position does not exist"))
				Expect(err).ToNot(HaveOccurred())
			})
			It("should report zero shares", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(alpacaController.Stock).To(Equal(api.StockInfo{
					Symbol:   stock,
					Position: 0,
				}))
			})
		})

	})

	Context("when placing an order", func() {
		var (
			order *alpaca.Order
		)

		JustBeforeEach(func() {
			mockClient = internal.NewMockAlpacaClient()
			mockAlgorithm = internal.NewMockAlgorithm()
			alpacaController, err = controller.NewAlpacaController(mockClient, mockAlgorithm, stock)
		})

		Context("when target position is greater than current position", func() {
			JustBeforeEach(func() {
				order, err = alpacaController.SendLimitOrder(5, 1.25)
			})
			It("should submit a BUY order for 2 shares and return the order", func() {
				Expect(err).ToNot(HaveOccurred())
				limitPrice := decimal.NewFromFloat(1.25)
				Expect(order).To(Equal(&alpaca.Order{
					ID:          "order123",
					Symbol:      stock,
					Side:        alpaca.Buy,
					Type:        alpaca.Limit,
					Qty:         decimal.NewFromFloat(2),
					LimitPrice:  &limitPrice,
					TimeInForce: alpaca.Day,
				}))
			})
		})

		Context("when target position is less than current position", func() {
			JustBeforeEach(func() {
				order, err = alpacaController.SendLimitOrder(2, 1.25)
			})
			It("should submit a SELL order for 1 share and return the order", func() {
				Expect(err).ToNot(HaveOccurred())
				limitPrice := decimal.NewFromFloat(1.25)
				Expect(order).To(Equal(&alpaca.Order{
					ID:          "order123",
					Symbol:      stock,
					Side:        alpaca.Sell,
					Type:        alpaca.Limit,
					Qty:         decimal.NewFromFloat(1),
					LimitPrice:  &limitPrice,
					TimeInForce: alpaca.Day,
				}))
			})
		})

		Context("when target position is equal to the current position", func() {
			JustBeforeEach(func() {
				order, err = alpacaController.SendLimitOrder(3, 1.25)
			})
			It("should return an error for no-op", func() {
				Expect(err).To(MatchError("no-op order requested"))
				Expect(order).To(Equal(&alpaca.Order{}))
			})
		})
	})

})
