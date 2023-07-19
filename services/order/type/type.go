package order

type Order struct {
	OrderId  string   `json:"order_id,omitempty" db:"order_id,omitempty"`
	DishesId []string `json:"dishes_id,omiempty" db:"dishes_id,omiempty"`
}

type Dish struct {
	DishId        string   `json:"dish_id,omiempty" db:"dish_id,omiempty"`
	Name          string   `json:"name,omiempty" db:"name,omiempty"`
	Description   string   `json:"description,omiempty" db:"description,omiempty"`
	IngredientsId []string `json:"ingredient_id,omiempty" db:"ingredient_id,omiempty"`
}

type Ingredient struct {
	IngId   string `json:"inredient_id,omiempty" db:"ingredient_id,omiempty"`
	Name    string `json:"ingredient,omiempty" db:"ingredient,omiempty"`
	Quality int    `json:"quality,omiempty" db:"quality,omiempty"`
}
