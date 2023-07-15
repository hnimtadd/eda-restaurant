package entities

type Dish struct {
	BaseEntity
	DishId      string `json:"dish_id,omiempty" db:"dish_id,omiempty"`
	Name        string `json:"name,omiempty" db:"name,omiempty"`
	Description string `json:"description,omiempty" db:"description,omiempty"`
}
