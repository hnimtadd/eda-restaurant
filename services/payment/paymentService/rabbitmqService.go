package paymentservice

import (
	"edaRestaurant/services/config"
	"edaRestaurant/services/payment/paymentrepository"
	payment "edaRestaurant/services/payment/type"
	queueagent "edaRestaurant/services/queueAgent"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type paymentService struct {
	conn      *amqp.Connection
	config    config.RabbitmqConfig
	repo      paymentrepository.PaymentRepository
	publisher queueagent.Publisher
}

func NewPaymentService(repo paymentrepository.PaymentRepository, publisher queueagent.Publisher, config config.RabbitmqConfig) (PaymentService, error) {
	service := &paymentService{
		repo:      repo,
		config:    config,
		publisher: publisher,
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
		log.Printf("[ERROR]: %v", err)
		return err
	}
	if err := s.publisher.ReplyWithValue(d, "", rsp); err != nil {
		return err
	}
	log.Println("[makepayment] REPLIED")
	return nil
}

func (s *paymentService) GetDishMoney(dishid ...string) (float64, error) {
	return 100, nil
}
func (s *paymentService) CheckPaymentHandler(msg any) error {
	d := msg.(amqp.Delivery)
	log.Println("Received from: ", d.ReplyTo)
	var req payment.CheckPaymentRequest
	if err := json.Unmarshal(d.Body, &req); err != nil {
		log.Printf("[ERROR]: %v", err)
		d.Nack(false, false)
	}

	money, err := s.GetDishMoney(req.DishId...)
	if err != nil {
		log.Printf("[ERROR]: %v", err)
		d.Nack(false, false)
	}
	rsp := payment.CheckPaymentResponse{
		TableId:      req.TableId,
		OrderId:      req.OrderId,
		CurrentMoney: money,
	}
	if err := s.publisher.ReplyWithValue(d, "", rsp); err != nil {
		log.Printf("[ERROR]: %v", err)
		d.Nack(false, false)
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
			log.Println("[Payment] Received message on queue")
			var err error
			if d.Type == "check" {
				err = s.CheckPaymentHandler(d)
				if err != nil {
					log.Printf("[ERROR] %v", err)
				}
			} else if d.Type == "make" {
				err = s.MakePaymentHandler(d)
				if err != nil {
					log.Printf("[ERROR] %v", err)
				}
			} else {
				log.Printf("[ERROR] no handler available")
			}
			if err != nil {
				if d.Redelivered {
					d.Nack(false, false)
				} else {
					d.Nack(false, true)
				}
			} else {
				d.Ack(false)
			}
		}
	}()
	log.Printf("[*] Listening to queue: %s\n", queue.Name)
	wg.Wait()
	log.Println("exited")
}

func (s *paymentService) ProcessBankPaymentRequest(req *payment.PaymentRequest) (*payment.PaymentWithBankRsp, error) {
	log.Printf("[Payment]: Processing Payment request with bank %v", req)
	time.Sleep(time.Second * 5)
	money, err := s.GetDishMoney(req.DishId...)
	if err != nil {
		return nil, err
	}
	supplier := req.Supplier
	if supplier == "" {
		return nil, errors.New(fmt.Sprintf("bad payment header: %v", req.PaymentType))
	}
	bankInfo, err := s.repo.GetBankInformation(supplier)
	if err != nil {
		return nil, err
	}
	BankingUrl := "TODO: use engine to generate payment url for bank and bill"
	expiredAt := time.Now().Add(time.Minute * 5).Unix()
	rsp := payment.PaymentWithBankRsp{
		PaymentId:  uuid.New().String(),
		OrderId:    req.OrderId,
		Price:      money,
		BankingUrl: BankingUrl,
		Metadata:   *bankInfo,
		ExpiredAt:  expiredAt,
	}
	log.Println("[Payment]: DONE")
	return &rsp, nil
}

func (s *paymentService) ProcessCashPaymentRequest(req *payment.PaymentRequest) (*payment.PaymentWithCashRsp, error) {
	log.Printf("[Payment]: Processing Payment request with cash %v", req)
	time.Sleep(time.Second * 5)
	money, err := s.GetDishMoney(req.DishId...)
	if err != nil {
		return nil, err
	}

	rsp := payment.PaymentWithCashRsp{
		PaymentId: uuid.New().String(),
		OrderId:   req.OrderId,
		TableId:   req.TableId,
		Price:     money,
	}

	log.Println("[Payment]: DONE")
	return &rsp, nil
}

func (s *paymentService) ProcessWalletPaymentRequest(req *payment.PaymentRequest) (*payment.PaymentWithWalletRsp, error) {
	log.Printf("[Payment]: Processing Payment request with wallet %v", req)
	time.Sleep(time.Second * 5)
	money, err := s.GetDishMoney(req.DishId...)
	if err != nil {
		return nil, err
	}
	supplier := req.Supplier
	if supplier == "" {
		return nil, errors.New(fmt.Sprintf("bad payment header: %v", req.PaymentType))
	}
	walletInfo, err := s.repo.GetWalletInformation(supplier)
	walletUrl := "TODO: use engine to generate payment url for wallet and bill"
	expiredAt := time.Now().Add(time.Minute * 5).Unix()
	rsp := payment.PaymentWithWalletRsp{
		PaymentId: uuid.New().String(),
		OrderId:   req.OrderId,
		Price:     money,
		WalletUrl: walletUrl,
		Metadata:  *walletInfo,
		ExpiredAt: expiredAt,
	}
	log.Println("[Payment]: DONE")
	return &rsp, nil
}
