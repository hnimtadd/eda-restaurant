package cook

import "time"

type Cook struct {
	CookId    string    `json:"cook_id,omiempty" db:"cook_id,omiempty"`
	OrderId   string    `json:"order_id,omiempty" db:"order_id,omiempty"`
	Timestamp time.Time `json:"timestamp,omiempty" db:"timestamp,omiempty"`
}
