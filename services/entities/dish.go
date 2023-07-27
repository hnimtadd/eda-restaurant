package entities

type Dish struct {
	BaseEntity
	DishId      string   `json:"dish_id,omiempty" db:"dish_id,omiempty"`
	Ingredients []string `json:"ingredients_id,omiempty" db:"ingredients_id,omiempty"`
	Name        string   `json:"name,omiempty" db:"name,omiempty"`
	Description string   `json:"description,omiempty" db:"description,omiempty"`
	Price       float64  `json:"price,omiempty" db:"price,omiempty"`
}

type Ingredient struct {
	IngredientId string `json:"ingredient_id,omiempty" db:"ingredient_id,omiempty"`
	Name         string `json:"name,omiempty" db:"name,omiempty"`
	Quality      int    `json:"quality,omiempty" db:"quality,omiempty"`
}
