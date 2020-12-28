package controller_test

import (
	"errors"
	"fmt"

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
		var (
			equity       = float64(1000)
			multiplier   = float64(2.00)
			mockSettings = internal.MockAlpacaClientInput{
				Account: alpaca.Account{
					ID:         "foobar",
					Equity:     decimal.NewFromFloat(equity),
					Multiplier: fmt.Sprintf("%.2f", multiplier),
				},
				Position: alpaca.Position{
					Qty: decimal.NewFromFloat(3.5),
				},
			}
		)

		JustBeforeEach(func() {
			mockClient = internal.NewMockAlpacaClient(mockSettings)
			mockAlgorithm = internal.NewMockAlgorithm()
			alpacaController, err = controller.NewAlpacaController(mockClient, mockAlgorithm, stock)
		})

		It("should not have failed", func() {
			Expect(err).ToNot(HaveOccurred())
		})
		It("should have initial account and position stats", func() {
			Expect(alpacaController.Account).To(Equal(api.AccountInfo{
				ID:               mockSettings.Account.ID,
				Equity:           equity,
				MarginMultiplier: multiplier,
			}))
			Expect(alpacaController.Stock).To(Equal(api.StockInfo{
				Symbol:   stock,
				Position: 3,
			}))
		})

		Context("when no existing positions are found", func() {
			BeforeEach(func() {
				mockSettings.Errors = []error{
					nil,                                   // CancelAllOrders
					errors.New("position does not exist"), // GetPosition
				}
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
			orderID      string
			equity       = float64(1000)
			multiplier   = float64(2.00)
			mockSettings = internal.MockAlpacaClientInput{
				Account: alpaca.Account{
					ID:         "foobar",
					Equity:     decimal.NewFromFloat(equity),
					Multiplier: fmt.Sprintf("%.2f", multiplier),
				},
				Position: alpaca.Position{
					Qty: decimal.NewFromFloat(3.5),
				},
				OrderStack: []alpaca.Order{
					{ID: "foobar123"}, // PlaceOrder
				},
			}
		)

		JustBeforeEach(func() {
			mockClient = internal.NewMockAlpacaClient(mockSettings)
			mockAlgorithm = internal.NewMockAlgorithm()
			alpacaController, err = controller.NewAlpacaController(mockClient, mockAlgorithm, stock)
		})

		Context("when target position is greater than current position", func() {
			JustBeforeEach(func() {
				orderID, err = alpacaController.SendLimitOrder(5, 1.25)
			})
			It("should submit a BUY order for 2 shares and return the order ID", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(orderID).To(Equal("foobar123"))
				// TODO: test order parameters
			})
		})
	})

})
