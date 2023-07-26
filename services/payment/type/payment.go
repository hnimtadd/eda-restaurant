package payment

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

type PaymentEntity struct {
	PaymentId string  `json:"payment_id,omiempty" db:"payment_id,omiempty"`
	TableId   string  `json:"table_id,omiempty" db:"table_id,omiempty"`
	OrderId   string  `json:"order_id,omitempty" db:"order_id,omitempty"`
	Price     float64 `json:"price,omiempty" db:"price,omiempty"`
	Type      string  `json:"payment_type,omiempty" db:"payment_type,omiempty"`
	Status    string  `json:"status,omiempty" db:"status,omiempty"`
}

type Payment struct {
	PaymentEntity
	Metadata PaymentMetadata `json:"metadata,omiempty" db:"metadata,omiempty"`
}

type PaymentMetadata struct {
	SupplierId string `json:"source_id,omiempty" db:"source_id,omiempty"`
	Supplier   string `json:"payment_source,omiempty" db:"payment_source,omiempty"`
	Endpoint   string `json:"source_endpoint,omiempty" db:"source_endpoint,omiempty"` // Username or card id
	ExpiredAt  int64  `json:"expired_at,omiempty"`
}
