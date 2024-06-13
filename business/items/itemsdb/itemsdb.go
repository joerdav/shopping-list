package itemsdb

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/joerdav/shopping-list/business/items"
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

func (f *Storer) Create(ctx context.Context, item items.Item) error {
	return f.store.CreateItem(ctx, db.CreateItemParams{
		ID:     item.ID,
		Name:   item.Name,
		UserID: item.UserID,
		ShopID: item.ShopID,
	})
}

func (f *Storer) QueryAllByShopID(ctx context.Context, shopID uuid.UUID) ([]items.Item, error) {
	ditems, err := f.store.ListItemsByShop(ctx, shopID)
	if err != nil {
		return []items.Item{}, err
	}
	items := make([]items.Item, len(ditems))
	for i, s := range ditems {
		items[i] = toCoreItem(s)
	}
	return items, nil
}

func (f *Storer) Query(ctx context.Context, id uuid.UUID) (items.Item, error) {
	item, err := f.store.GetItem(ctx, id)
	if err != nil {
		return items.Item{}, err
	}
	return toCoreItem(item), nil
}

func (f *Storer) QueryAll(ctx context.Context, userID string) ([]items.Item, error) {
	ditems, err := f.store.ListItems(ctx, userID)
	if err != nil {
		return []items.Item{}, err
	}
	items := make([]items.Item, len(ditems))
	for i, s := range ditems {
		items[i] = toCoreItem(s)
	}
	return items, nil
}

func (f *Storer) Update(ctx context.Context, p items.Item) (items.Item, error) {
	item, err := f.store.UpdateItem(ctx, db.UpdateItemParams{
		Name:   p.Name,
		ShopID: p.ShopID,
		ID:     p.ID,
	})
	if err != nil {
		return items.Item{}, err
	}
	return toCoreItem(item), nil
}

func toCoreItem(s db.Item) items.Item {
	return items.Item{
		ID:     s.ID,
		Name:   s.Name,
		UserID: s.UserID,
		ShopID: s.ShopID,
	}
}
