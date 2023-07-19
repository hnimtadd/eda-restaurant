package storage

type Ingedient struct {
	Name    string `json:"name,omiempty" db:"name,omiempty"`
	Quality int    `json:"quality,omiempty" db:"quality,omiempty"`
}
