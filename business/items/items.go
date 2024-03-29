package items

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("item not found")

type NewItem struct {
	Name   string
	ShopID uuid.UUID
}

type Item struct {
	ID     uuid.UUID
	Name   string
	ShopID uuid.UUID
}

type Storer interface {
	Create(context.Context, Item) error
	Query(context.Context, uuid.UUID) (Item, error)
	QueryAll(context.Context) ([]Item, error)
	QueryAllByShopID(context.Context, uuid.UUID) ([]Item, error)
}

type Core struct {
	storer Storer
}

func NewCore(storer Storer) *Core {
	return &Core{storer}
}

func (c *Core) Create(ctx context.Context, s NewItem) (Item, error) {
	item := Item{
		ID:   uuid.New(),
		Name: s.Name,
		ShopID: s.ShopID,
	}
	if err := c.storer.Create(ctx, item); err != nil {
		return Item{}, err
	}
	return item, nil
}

func (c *Core) Query(ctx context.Context, id uuid.UUID) (Item, error) {
	return c.storer.Query(ctx, id)
}

func (c *Core) QueryAll(ctx context.Context) ([]Item, error) {
	return c.storer.QueryAll(ctx)
}

func (c *Core) QueryByShopID(ctx context.Context, shopID uuid.UUID) ([]Item, error) {
	return c.storer.QueryAllByShopID(ctx, shopID)
}
