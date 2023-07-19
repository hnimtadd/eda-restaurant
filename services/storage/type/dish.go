package storage

type Dish struct {
	Ingredients []string `json:"ingredients_id,omiempty" db:"ingredients_id,omiempty"`
	Name        string   `json:"name,omiempty" db:"name,omiempty"`
	Description string   `json:"description,omiempty" db:"description,omiempty"`
}
