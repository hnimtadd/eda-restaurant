package entities

type Order struct {
	BaseEntity
	OrderId  string   `json:"order_id,omitempty" db:"order_id,omitempty"`
	DishesId []string `json:"dishes_id,omiempty" db:"dishes_id,omiempty"`
	TableId  string   `json:"table_id,omiempty" db:"table_id,omiempty"`
}
