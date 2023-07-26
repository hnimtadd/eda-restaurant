package entities

type Payment struct {
	BaseEntity
	PaymentId string          `json:"payment_id,omiempty" db:"payment_id,omiempty"`
	OrderId   string          `json:"order_id,omiempty" db:"order_id,omiempty"`
	Price     float64         `json:"price,omiempty" db:"price,omiempty"`
	Status    string          `json:"status,omiempty" db:"status,omiempty"`
	Metadata  PaymentMetadata `json:"metadata,omiempty" db:"metadata,omiempty"`
}

type PaymentMetadata struct {
	SourceId       string `json:"source_id,omiempty" db:"source_id,omiempty"`
	PaymentSource  string `json:"payment_source,omiempty" db:"payment_source,omiempty"`
	SourceEndpoint string `json:"source_endpoint,omiempty" db:"source_endpoint,omiempty"` // Username or card id
}

type BankInformation struct {
	BankId       string `json:"bank_id,omiempty" db:"bank_id,omiempty"`
	BankSupplier string `json:"bank_supplier"`
	BankEndpoint string `json:"bank_endpoint,omiempty" db:"bank_endpoint,omiempty"` // Username or card id
}

type WalletInformation struct {
	WalletSupplier string `json:"wallet_supplier"`
	WalletId       string `json:"wallet_id,omiempty" db:"wallet_id,omiempty"`             //
	WalletEndpoint string `json:"wallet_endpoint,omiempty" db:"wallet_endpoint,omiempty"` //
}
