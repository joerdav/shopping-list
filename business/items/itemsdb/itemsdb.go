package itemsdb

import (
	"context"
	"database/sql"
	_ "embed"

	"github.com/google/uuid"
	"github.com/joerdav/shopping-list/business/items"
)

//go:embed schema.sql
var schema string

type Storer struct {
	store *Queries
	conn  *sql.DB
}

func NewStorer(conn *sql.DB) *Storer {
	return &Storer{
		store: New(conn),
		conn:  conn,
	}
}

func (f *Storer) Migrate(ctx context.Context) error {
	_, err := f.conn.ExecContext(ctx, schema)
	return err
}

func (f *Storer) Create(ctx context.Context, item items.Item) error {
	return f.store.CreateItem(ctx, CreateItemParams{
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
	item, err := f.store.UpdateItem(ctx, UpdateItemParams{
		Name:   p.Name,
		ShopID: p.ShopID,
		ID:     p.ID,
	})
	if err != nil {
		return items.Item{}, err
	}
	return toCoreItem(item), nil
}

func toCoreItem(s Item) items.Item {
	return items.Item{
		ID:     s.ID,
		Name:   s.Name,
		UserID: s.UserID,
		ShopID: s.ShopID,
	}
}
