package paymentrepository

import (
	"context"
	"edaRestaurant/services/config"
	"edaRestaurant/services/entities"
	payment "edaRestaurant/services/payment/type"
	"errors"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type paymentRepository struct {
	db     *mongo.Database
	config config.MongoConfig
}

func NewPaymentRepository(config config.MongoConfig) (PaymentRepository, error) {
	repo := &paymentRepository{
		config: config,
	}
	if err := repo.initDB(); err != nil {
		return nil, err
	}
	return repo, nil
}
func (repo *paymentRepository) initDB() error {
	// TODO: init connection to mongodb
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientOpts := options.Client().ApplyURI(repo.config.Source).SetAuth(
		options.Credential{
			AuthSource: repo.config.AuthSource,
			Username:   repo.config.Username,
			Password:   repo.config.Password,
		},
	).SetTLSConfig(nil).SetTimeout(5 * time.Second)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return err
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return err
	}
	repo.db = client.Database(repo.config.Database)
	log.Printf("Connected to mongodb with config: %v\n", repo.config)
	return nil
}
func (repo *paymentRepository) GetWalletInformation(supplier string) (*entities.WalletInformation, error) {
	walletInfo := entities.WalletInformation{
		WalletSupplier: supplier,
		WalletId:       "sample-wallet",
		WalletEndpoint: "/url/wallet/endpoint",
	}
	return &walletInfo, nil
}
func (repo *paymentRepository) GetBankInformation(supplier string) (*entities.BankInformation, error) {
	bankInfo := entities.BankInformation{
		BankSupplier: supplier,
		BankId:       "sample-bank",
		BankEndpoint: "/url/bank/endpoint",
	}
	return &bankInfo, nil
}

func (s *paymentRepository) CreatePaymentHistory(payment payment.Payment) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	paymentLog := entities.Payment{
		PaymentId: payment.PaymentId,
		OrderId:   payment.OrderId,
		Price:     payment.Price,
		Metadata: entities.PaymentMetadata{
			PaymentSource: payment.Metadata.Supplier,
			SourceId:      payment.Metadata.SupplierId,
		},
	}
	_, err := s.db.Collection("payments").InsertOne(ctx, paymentLog)
	if err != nil {
		return err
	}
	return nil

}

func (s *paymentRepository) GetPaymentsHistory() ([]entities.Payment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	cur, err := s.db.Collection("payments").Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	payments := []entities.Payment{}
	for cur.Next(ctx) {
		var payment entities.Payment
		if err := cur.Decode(&payment); err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return payments, nil
}
func (s *paymentRepository) MarkDonePayment(paymentId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	filter := bson.D{primitive.E{Key: "paymentid", Value: paymentId}}
	update := bson.D{primitive.E{Key: "$set", Value: bson.D{primitive.E{Key: "status", Value: "done "}}}}
	res := s.db.Collection("payments").FindOneAndUpdate(ctx, filter, update)
	if err := res.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New(fmt.Sprintf("Not found document with id :%v in database", paymentId))
		}
		return err
	}
	return nil
}
