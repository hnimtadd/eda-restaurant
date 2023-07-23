package tableRunner

type TableServeRequest struct {
	TableId string   `json:"table_id,omiempty" db:"table_id,omiempty"`
	DishId  []string `json:"dish_id,omiempty" db:"dish_id,omiempty"`
}

type TableCleanRequest struct {
	TableId string `json:"table_id,omiempty" db:"table_id,omiempty"`
}
