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
	UserID string
}

type UpdateItem struct {
	ID     uuid.UUID
	Name   string
	ShopID uuid.UUID
}

type Item struct {
	ID     uuid.UUID
	Name   string
	ShopID uuid.UUID
	UserID string
}

type Storer interface {
	Create(context.Context, Item) error
	Update(context.Context, Item) (Item, error)
	Query(context.Context, uuid.UUID) (Item, error)
	QueryAll(context.Context, string) ([]Item, error)
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
		ID:     uuid.New(),
		Name:   s.Name,
		UserID: s.UserID,
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

func (c *Core) QueryAll(ctx context.Context, userID string) ([]Item, error) {
	return c.storer.QueryAll(ctx, userID)
}

func (c *Core) Update(ctx context.Context, u UpdateItem) (Item, error) {
	item, err := c.Query(ctx, u.ID)
	if err != nil {
		return Item{}, err
	}
	item.ShopID = u.ShopID
	item.Name = u.Name
	item, err = c.storer.Update(ctx, item)
	if err != nil {
		return Item{}, err
	}
	return item, nil
}

func (c *Core) QueryByShopID(ctx context.Context, shopID uuid.UUID) ([]Item, error) {
	return c.storer.QueryAllByShopID(ctx, shopID)
}
