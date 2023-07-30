package storage

type Ingedient struct {
	Id      string `json:"ingredient_id,omiempty" db:"ingredient_id,omiempty"`
	Name    string `json:"name,omiempty" db:"name,omiempty"`
	Quality int    `json:"quality,omiempty" db:"quality,omiempty"`
}
