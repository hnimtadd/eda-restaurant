package restype

type Dish struct {
	DishId      string   `json:"dish_id,omiempty" db:"dish_id,omiempty"`
	Ingredients []string `json:"ingredients_id,omiempty" db:"ingredients_id,omiempty"`
	Name        string   `json:"name,omiempty" db:"name,omiempty"`
	Description string   `json:"description,omiempty" db:"description,omiempty"`
}

type Ingredient struct {
	IngId   string `json:"inredient_id,omiempty" db:"ingredient_id,omiempty"`
	Name    string `json:"ingredient,omiempty" db:"ingredient,omiempty"`
	Quality int    `json:"quality,omiempty" db:"quality,omiempty"`
}
