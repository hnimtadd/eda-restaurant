package paymentrepository

import payment "edaRestaurant/services/payment/type"

type PaymentRepository interface {
	GetWalletInformation(supplier string) (*payment.WalletInformation, error)
	GetBankInformation(supplier string) (*payment.BankInformation, error)
}
