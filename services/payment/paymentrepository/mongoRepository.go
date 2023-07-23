package paymentrepository

import (
	"context"
	"edaRestaurant/services/config"
	payment "edaRestaurant/services/payment/type"
	"log"
	"time"

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
func (repo *paymentRepository) GetWalletInformation(supplier string) (*payment.WalletInformation, error) {
	walletInfo := payment.WalletInformation{
		WalletSupplier: supplier,
		WalletId:       "sample-wallet",
		WalletEndpoint: "/url/wallet/endpoint",
	}
	return &walletInfo, nil
}
func (repo *paymentRepository) GetBankInformation(supplier string) (*payment.BankInformation, error) {
	bankInfo := payment.BankInformation{
		BankSupplier: supplier,
		BankId:       "sample-bank",
		BankEndpoint: "/url/bank/endpoint",
	}
	return &bankInfo, nil
}
