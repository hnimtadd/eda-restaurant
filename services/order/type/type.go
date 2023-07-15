package order

type Order struct {
	OrderId  string `json:"order_id,omitempty" db:"order_id,omitempty"`
	DishesId []int  `json:"dishes_id,omiempty" db:"dishes_id,omiempty"`
}

type Dish struct {
	DishId      string `json:"dish_id,omiempty" db:"dish_id,omiempty"`
	Name        string `json:"name,omiempty" db:"name,omiempty"`
	Description string `json:"description,omiempty" db:"description,omiempty"`
}
