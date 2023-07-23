package cook

type Cook struct {
	CookId    string `json:"cook_id,omiempty" db:"cook_id,omiempty"`
	OrderId   string `json:"order_id,omiempty" db:"order_id,omiempty"`
	Timestamp int64  `json:"timestamp,omiempty" db:"timestamp,omiempty"`
}

type TableServeRequest struct {
	OrderId string   `json:"order_id,omiempty" db:"order_id,omiempty"`
	TableId string   `json:"table_id,omiempty" db:"table_id,omiempty"`
	DishId  []string `json:"dish_id,omiempty" db:"dish_id,omiempty"`
}
