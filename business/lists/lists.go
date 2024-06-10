package lists

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("list not found")

type Storer interface {
	Create(context.Context, List) error
	Update(context.Context, List) error
	Delete(context.Context, uuid.UUID) error
	Query(context.Context, uuid.UUID) (List, error)
	QueryAll(context.Context, string) ([]List, error)
}

type Core struct {
	storer Storer
}

func NewCore(storer Storer) *Core {
	return &Core{storer}
}

func (c *Core) Update(ctx context.Context, s UpdateList) (List, error) {
	list, err := c.storer.Query(ctx, s.ID)
	if err != nil {
		return List{}, err
	}
	if s.Recipes != nil {
		list.Recipes = *s.Recipes
	}
	if s.Items != nil {
		list.Items = *s.Items
	}
	if err := c.storer.Update(ctx, list); err != nil {
		return List{}, err
	}
	return list, nil
}

func (c *Core) Create(ctx context.Context, s NewList) (List, error) {
	list := List{
		ID:          uuid.New(),
		UserID:      s.UserID,
		CreatedDate: s.CreatedDate,
		Recipes:     map[uuid.UUID]int{},
		Items:       map[uuid.UUID]int{},
	}
	if err := c.storer.Create(ctx, list); err != nil {
		return List{}, err
	}
	return list, nil
}

func (c *Core) Query(ctx context.Context, id uuid.UUID) (List, error) {
	return c.storer.Query(ctx, id)
}

func (c *Core) QueryAll(ctx context.Context, userID string) ([]List, error) {
	return c.storer.QueryAll(ctx, userID)
}

func (c *Core) Delete(ctx context.Context, listID uuid.UUID) error {
	return c.storer.Delete(ctx, listID)
}
