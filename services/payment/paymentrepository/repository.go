package paymentrepository

import (
	"edaRestaurant/services/entities"
	payment "edaRestaurant/services/payment/type"
)

type PaymentRepository interface {
	GetWalletInformation(supplier string) (*entities.WalletInformation, error)
	GetBankInformation(supplier string) (*entities.BankInformation, error)
	CreatePaymentHistory(payment.Payment) error
	GetPaymentWithOrderId(orderId string) (*entities.Payment, error)
	MarkDonePayment(paymentId string) error
	GetPaymentsHistory() ([]entities.Payment, error)
	GetDishInformatio(dishId string) (*entities.Dish, error)
}
