package restype

type Cook struct {
	CookId    string `json:"cook_id,omiempty" db:"cook_id,omiempty"`
	OrderId   string `json:"order_id,omiempty" db:"order_id,omiempty"`
	Timestamp int64  `json:"timestamp,omiempty" db:"timestamp,omiempty"`
}
