package order

type Order struct {
	TableId  string   `json:"table_id,omiempty" db:"table_id,omiempty"`
	OrderId  string   `json:"order_id,omitempty" db:"order_id,omitempty"`
	DishesId []string `json:"dishes_id,omiempty" db:"dishes_id,omiempty"`
}

type Dish struct {
	DishId        string   `json:"dish_id,omiempty" db:"dish_id,omiempty"`
	Name          string   `json:"name,omiempty" db:"name,omiempty"`
	Description   string   `json:"description,omiempty" db:"description,omiempty"`
	IngredientsId []string `json:"ingredient_id,omiempty" db:"ingredient_id,omiempty"`
}

type Ingredient struct {
	IngId   string `json:"inredient_id,omiempty" db:"ingredient_id,omiempty"`
	Name    string `json:"ingredient,omiempty" db:"ingredient,omiempty"`
	Quality int    `json:"quality,omiempty" db:"quality,omiempty"`
}

type TableCleanRequest struct {
	TableId string `json:"table_id,omiempty" db:"table_id,omiempty"`
}

type OrderPayment struct {
	TableId  string `json:"table_id,omiempty" db:"table_id,omiempty"`
	OrderId  string `json:"order_id,omitempty" db:"order_id,omitempty"`
	Price    string `json:"price,omiempty" db:"price,omiempty"`
	Discount string `json:"discount,omiempty" db:"discount,omiempty"`
}

type CheckPaymentRequest struct {
	TableId string   `json:"table_id,omiempty" db:"table_id,omiempty"`
	OrderId string   `json:"order_id,omitempty" db:"order_id,omitempty"`
	DishId  []string `json:"dish_id,omiempty" db:"dish_id,omiempty"`
}

type CheckPaymentResponse struct {
	TableId      string  `json:"table_id,omiempty" db:"table_id,omiempty"`
	OrderId      string  `json:"order_id,omitempty" db:"order_id,omitempty"`
	CurrentMoney float64 `json:"current_money,omiempty"`
}

type PaymentRequest struct {
	PaymentType string   `json:"payment_type,omiempty" db:"payment_type,omiempty"`
	TableId     string   `json:"table_id,omiempty" db:"table_id,omiempty"`
	OrderId     string   `json:"order_id,omitempty" db:"order_id,omitempty"`
	DishId      []string `json:"dish_id,omiempty" db:"dish_id,omiempty"`
	Supplier    string   `json:"supplier,omiempty" db:"supplier,omiempty"`
}

type PaymentWithCashRsp struct {
	PaymentId string  `json:"payment_id,omiempty" db:"payment_id,omiempty"`
	TableId   string  `json:"table_id,omiempty" db:"table_id,omiempty"`
	OrderId   string  `json:"order_id,omitempty" db:"order_id,omitempty"`
	Price     float64 `json:"price,omiempty" db:"price,omiempty"`
}

type BankInformation struct {
	BankId       string `json:"bank_id,omiempty" db:"bank_id,omiempty"`
	BankSupplier string `json:"bank_supplier"`
	BankEndpoint string `json:"bank_endpoint,omiempty" db:"bank_endpoint,omiempty"` // Username or card id
}

type PaymentWithBankRsp struct {
	PaymentId  string          `json:"payment_id,omiempty" db:"payment_id,omiempty"`
	OrderId    string          `json:"order_id,omitempty" db:"order_id,omitempty"`
	Price      float64         `json:"price,omiempty" db:"price,omiempty"`
	BankingUrl string          `json:"banking_url,omiempty"`
	Metadata   BankInformation `json:"metadata"`
	ExpiredAt  int64           `json:"expired_at,omiempty"`
}

type WalletInformation struct {
	WalletSupplier string `json:"wallet_supplier"`
	WalletId       string `json:"wallet_id,omiempty" db:"wallet_id,omiempty"`             //
	WalletEndpoint string `json:"wallet_endpoint,omiempty" db:"wallet_endpoint,omiempty"` //
}

type PaymentWithWalletRsp struct {
	PaymentId string            `json:"payment_id,omiempty" db:"payment_id,omiempty"`
	OrderId   string            `json:"order_id,omitempty" db:"order_id,omitempty"`
	Price     float64           `json:"price,omiempty" db:"price,omiempty"`
	WalletUrl string            `json:"wallet_url,omiempty"` // Wallet URL to render qr code
	Metadata  WalletInformation `json:"metadata"`
	ExpiredAt int64             `json:"expired_at,omiempty"`
}
