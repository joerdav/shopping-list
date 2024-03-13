package shops

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("shop not found")

type Storer interface {
	Create(context.Context, Shop) error
	Query(context.Context, uuid.UUID) (Shop, error)
	QueryAll(context.Context, string) ([]Shop, error)
}

type Core struct {
	storer Storer
}

func NewCore(storer Storer) *Core {
	return &Core{storer}
}

func (c *Core) Create(ctx context.Context, s NewShop) (Shop, error) {
	shop := Shop{
		ID:     uuid.New(),
		Name:   s.Name,
		UserID: s.UserID,
	}
	if err := c.storer.Create(ctx, shop); err != nil {
		return Shop{}, err
	}
	return shop, nil
}

func (c *Core) Query(ctx context.Context, id uuid.UUID) (Shop, error) {
	return c.storer.Query(ctx, id)
}

func (c *Core) QueryAll(ctx context.Context, userID string) ([]Shop, error) {
	return c.storer.QueryAll(ctx, userID)
}
