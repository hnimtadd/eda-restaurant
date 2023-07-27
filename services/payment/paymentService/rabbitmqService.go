package paymentservice

import (
	"edaRestaurant/services/config"
	"edaRestaurant/services/payment/paymentrepository"
	stripeengine "edaRestaurant/services/payment/stripeEngine"
	payment "edaRestaurant/services/payment/type"
	queueagent "edaRestaurant/services/queueAgent"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type paymentService struct {
	conn          *amqp.Connection
	config        config.RabbitmqConfig
	repo          paymentrepository.PaymentRepository
	publisher     queueagent.Publisher
	paymentEngine stripeengine.StripeEngine
}

func NewPaymentService(repo paymentrepository.PaymentRepository, engine stripeengine.StripeEngine, publisher queueagent.Publisher, config config.RabbitmqConfig) (PaymentService, error) {
	service := &paymentService{
		repo:          repo,
		config:        config,
		publisher:     publisher,
		paymentEngine: engine,
	}
	if err := service.initConnection(); err != nil {
		return nil, err
	}
	return service, nil
}

func (s *paymentService) initConnection() error {
	conn, err := amqp.Dial(s.config.Source)
	if err != nil {
		return err
	}
	s.conn = conn
	return nil
}
func (s *paymentService) MakePaymentHandler(msg any) error {
	d := msg.(amqp.Delivery)
	var (
		req = payment.PaymentRequest{}
		rsp any
		err error
	)
	if err = json.Unmarshal(d.Body, &req); err != nil {
		return err
	}
	switch paymentType := req.PaymentType; paymentType {
	case "bank":
		{
			// TODO with bank processor
			rsp, err = s.ProcessBankPaymentRequest(&req)
		}
	case "cash":
		{
			// TODO with cash processor
			rsp, err = s.ProcessCashPaymentRequest(&req)
		}
	case "wallet":
		{
			// TODO with wallet processor
			rsp, err = s.ProcessWalletPaymentRequest(&req)
		}
	default:
		{
			err = errors.New("No payment method found")
		}
	}
	if err != nil {
		return err
	}
	replyMsg, err := s.publisher.MakeMessageWithValue("payment", d.ReplyTo, "reply", d.CorrelationId, rsp)
	if err != nil {
		return err
	}
	if err := s.publisher.PublishWithMessage(&replyMsg); err != nil {
		return err
	}
	log.Println("[makepayment] REPLIED")
	return nil
}

func (s *paymentService) GetDishMoney(dishid ...string) (float64, error) {
	money := float64(0)
	for _, id := range dishid {
		dish, err := s.repo.GetDishInformatio(id)
		if err != nil {
			return 0, err
		}
		money += dish.Price
	}
	return money, nil
}
func (s *paymentService) CheckPaymentHandler(msg any) error {
	d := msg.(amqp.Delivery)
	var req payment.CheckPaymentRequest
	if err := json.Unmarshal(d.Body, &req); err != nil {
		log.Printf("[ERROR]: %v", err)
		return err
	}

	money, err := s.GetDishMoney(req.DishId...)
	if err != nil {
		log.Printf("[ERROR]: %v", err)
		return err
	}
	rsp := payment.CheckPaymentResponse{
		TableId:      req.TableId,
		OrderId:      req.OrderId,
		CurrentMoney: money,
	}
	// Reply to askqueue
	replyMsg, err := s.publisher.MakeMessageWithValue("payment", d.ReplyTo, "reply", d.CorrelationId, rsp)
	if err != nil {
		return err
	}
	if err := s.publisher.PublishWithMessage(&replyMsg); err != nil {
		log.Printf("[ERROR]: %v", err)
		return err
	}
	return nil
}

func (s *paymentService) InitBackground() {
	s.ListenAndServePaymentOrder()
}
func (s *paymentService) ListenAndServePaymentOrder() {
	ch, err := s.conn.Channel()
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	queue, err := ch.QueueDeclare(
		"payment",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	err = ch.QueueBind(
		queue.Name,
		queue.Name,
		"restaurant",
		false,
		nil,
	)

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	ds, err := ch.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for d := range ds {
			log.Println("[Payment] Received message on queue: ", d.Type)
			var err error
			if d.Type == "check" {
				err = s.CheckPaymentHandler(d)
				// if err != nil {
				// 	log.Printf("[ERROR] %v", err)
				// }
			} else if d.Type == "make" {
				err = s.MakePaymentHandler(d)
				// if err != nil {
				// 	log.Printf("[ERROR] %v", err)
				// }
			} else {
				err = errors.New("[ERROR] no handler available")
			}
			if err != nil {
				log.Printf("[ERROR]: %v", err)
				if d.Redelivered {
					log.Println("reject")
					d.Nack(false, false)
				} else {
					log.Println("redelivery")
					d.Nack(false, true)
				}
				continue
			} else {
				d.Ack(false)
			}
		}
	}()
	log.Printf("[*] Listening to queue: %s\n", queue.Name)
	wg.Wait()
	log.Println("exited")
}

func (s *paymentService) ProcessBankPaymentRequest(req *payment.PaymentRequest) (*payment.Payment, error) {
	log.Printf("[Payment]: Processing Payment request with bank %v", req)

	// time.Sleep(time.Second * 5)
	money, err := s.GetDishMoney(req.DishId...)
	if err != nil {
		return nil, err
	}

	supplier := req.Supplier
	if supplier == "" {
		return nil, errors.New(fmt.Sprintf("bad payment header: %v", req.PaymentType))
	}

	bankInfoEntity, err := s.repo.GetBankInformation(supplier)
	if err != nil {
		return nil, err
	}

	checkOutSession, err := s.paymentEngine.CreateCheckOutSession(money)
	if err != nil {
		log.Printf("[Payment] Error: %v", err)
		return nil, err
	}
	metadata := payment.PaymentMetadata{
		SupplierId: bankInfoEntity.BankId,
		Supplier:   bankInfoEntity.BankSupplier,
		Endpoint:   checkOutSession.URL,
		ExpiredAt:  checkOutSession.ExpiresAt,
	}

	payment := payment.Payment{
		PaymentEntity: payment.PaymentEntity{
			PaymentId: checkOutSession.ID,
			TableId:   req.TableId,
			OrderId:   req.OrderId,
			Price:     money,
		},
		Metadata: metadata,
	}

	if err := s.repo.CreatePaymentHistory(payment); err != nil {
		log.Printf("[Payment]: Error: %v", err)
		return &payment, err
	}

	log.Println("[Payment]: DONE")
	return &payment, nil
}

func (s *paymentService) ProcessCashPaymentRequest(req *payment.PaymentRequest) (*payment.Payment, error) {
	log.Printf("[Payment]: Processing Payment request with cash %v", req)
	// time.Sleep(time.Second * 5)
	money, err := s.GetDishMoney(req.DishId...)
	if err != nil {
		return nil, err
	}

	payment := payment.Payment{
		PaymentEntity: payment.PaymentEntity{
			PaymentId: uuid.New().String(),
			TableId:   req.TableId,
			OrderId:   req.OrderId,
			Price:     money,
			Type:      "cash",
		},
	}

	if err := s.repo.CreatePaymentHistory(payment); err != nil {
		log.Printf("[Payment]: Error: %v", err)
		return &payment, err
	}

	log.Println("[Payment]: DONE")
	return &payment, nil

}

func (s *paymentService) ProcessWalletPaymentRequest(req *payment.PaymentRequest) (*payment.Payment, error) {
	log.Printf("[Payment]: Processing Payment request with wallet %v", req)
	// time.Sleep(time.Second * 5)
	money, err := s.GetDishMoney(req.DishId...)
	// time.Sleep(time.Second * 5)
	if err != nil {
		return nil, err
	}

	supplier := req.Supplier
	if supplier == "" {
		return nil, errors.New(fmt.Sprintf("bad payment header: %v", req.PaymentType))
	}

	walletInfo, err := s.repo.GetWalletInformation(supplier)
	if err != nil {
		return nil, err
	}

	checkOutSession, err := s.paymentEngine.CreateCheckOutSession(money)
	if err != nil {
		log.Printf("[Payment] Error: %v", err)
		return nil, err
	}
	metadata := payment.PaymentMetadata{
		SupplierId: walletInfo.WalletId,
		Supplier:   walletInfo.WalletSupplier,
		Endpoint:   checkOutSession.URL,
		ExpiredAt:  checkOutSession.ExpiresAt,
	}

	payment := payment.Payment{
		PaymentEntity: payment.PaymentEntity{
			PaymentId: checkOutSession.ID,
			TableId:   req.TableId,
			OrderId:   req.OrderId,
			Price:     money,
			Type:      "wallet",
		},
		Metadata: metadata,
	}

	if err := s.repo.CreatePaymentHistory(payment); err != nil {
		log.Printf("[Payment]: Error: %v", err)
		return &payment, err
	}

	log.Println("[Payment]: DONE")
	return &payment, nil
}

func (s *paymentService) HandleCompletedPayment(paymentId string) error {
	if err := s.repo.MarkDonePayment(paymentId); err != nil {
		return err
	}
	return nil
}
