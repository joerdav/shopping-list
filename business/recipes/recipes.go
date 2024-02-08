package recipes

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("recipe not found")

type NewRecipe struct {
	Name string
}

type Recipe struct {
	ID          uuid.UUID
	Name        string
	Ingredients map[uuid.UUID]int
}

type UpdateRecipe struct {
	ID          uuid.UUID
	Name        *string
	Ingredients *map[uuid.UUID]int
}

type Storer interface {
	Create(context.Context, Recipe) error
	Update(context.Context, Recipe) error
	Query(context.Context, uuid.UUID) (Recipe, error)
	QueryAll(context.Context) ([]Recipe, error)
}

type Core struct {
	storer Storer
}

func NewCore(storer Storer) *Core {
	return &Core{storer}
}

func (c *Core) Create(ctx context.Context, s NewRecipe) (Recipe, error) {
	recipe := Recipe{
		ID:          uuid.New(),
		Name:        s.Name,
		Ingredients: map[uuid.UUID]int{},
	}
	if err := c.storer.Create(ctx, recipe); err != nil {
		return Recipe{}, err
	}
	return recipe, nil
}

func (c *Core) Update(ctx context.Context, s UpdateRecipe) (Recipe, error) {
	recipe, err := c.storer.Query(ctx, s.ID)
	if err != nil {
		return Recipe{}, err
	}
	if s.Name != nil {
		recipe.Name = *s.Name
	}
	if s.Ingredients != nil {
		recipe.Ingredients = *s.Ingredients
	}
	return recipe, c.storer.Update(ctx, recipe)
}

func (c *Core) Query(ctx context.Context, id uuid.UUID) (Recipe, error) {
	return c.storer.Query(ctx, id)
}

func (c *Core) QueryAll(ctx context.Context) ([]Recipe, error) {
	return c.storer.QueryAll(ctx)
}
