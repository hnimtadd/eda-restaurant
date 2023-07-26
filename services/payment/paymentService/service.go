package paymentservice

type PaymentService interface {
	MakePaymentHandler(msg any) error
	CheckPaymentHandler(msg any) error
	InitBackground()
	ListenAndServePaymentOrder()
	HandleCompletedPayment(paymentId string) error
}
