package recipesdb

import (
	"context"
	"database/sql"
	_ "embed"

	"github.com/google/uuid"
	"github.com/joerdav/shopping-list/business/recipes"
	"github.com/joerdav/shopping-list/db"
)

type Storer struct {
	store *db.Queries
	conn  *sql.DB
}

func NewStorer(conn *sql.DB) *Storer {
	return &Storer{
		store: db.New(conn),
		conn:  conn,
	}
}

func (f *Storer) Create(ctx context.Context, recipe recipes.Recipe) error {
	return f.store.CreateRecipe(ctx, db.CreateRecipeParams{
		ID:     recipe.ID,
		UserID: recipe.UserID,
		Name:   recipe.Name,
	})
}

func (f *Storer) Update(ctx context.Context, recipe recipes.Recipe) error {
	ingredients, err := f.store.ListIngredientsByRecipe(ctx, recipe.ID)
	if err != nil {
		return err
	}
	tx, err := f.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	qtx := f.store.WithTx(tx)
	for id, qty := range recipe.Ingredients {
		if err := qtx.SetIngredient(ctx, db.SetIngredientParams{
			ItemID:   id,
			RecipeID: recipe.ID,
			Quantity: int64(qty),
		}); err != nil {
			return err
		}
	}
	for _, i := range ingredients {
		if _, ok := recipe.Ingredients[i.ItemID]; !ok {
			if err := qtx.DeleteIngredient(ctx, db.DeleteIngredientParams{
				ItemID:   i.ItemID,
				RecipeID: recipe.ID,
			}); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (f *Storer) Query(ctx context.Context, id uuid.UUID) (recipes.Recipe, error) {
	recipe, err := f.store.GetRecipe(ctx, id)
	if err != nil {
		return recipes.Recipe{}, err
	}
	ingredients, err := f.store.ListIngredientsByRecipe(ctx, id)
	if err != nil {
		return recipes.Recipe{}, err
	}
	cr := toCoreRecipe(recipe, ingredients)
	return cr, nil
}

func (f *Storer) QueryAll(ctx context.Context, userID string) ([]recipes.Recipe, error) {
	recipesList, err := f.store.ListRecipes(ctx, userID)
	if err != nil {
		return nil, err
	}
	var out []recipes.Recipe
	for _, r := range recipesList {
		ingredients, err := f.store.ListIngredientsByRecipe(ctx, r.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, toCoreRecipe(r, ingredients))
	}
	return out, nil
}

func toCoreRecipe(r db.Recipe, is []db.Ingredient) recipes.Recipe {
	return recipes.Recipe{
		ID:          r.ID,
		UserID:      r.UserID,
		Name:        r.Name,
		Ingredients: toCoreIngredients(is),
	}
}

func toCoreIngredients(ingredients []db.Ingredient) map[uuid.UUID]int {
	out := map[uuid.UUID]int{}
	for _, i := range ingredients {
		out[i.ItemID] = int(i.Quantity)
	}
	return out
}
