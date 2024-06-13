package shopsdb

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/joerdav/shopping-list/business/shops"
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

func (f *Storer) Create(ctx context.Context, shop shops.Shop) error {
	return f.store.CreateShop(ctx, db.CreateShopParams{
		ID:     shop.ID,
		UserID: shop.UserID,
		Name:   shop.Name,
	})
}

func (f *Storer) Query(ctx context.Context, id uuid.UUID) (shops.Shop, error) {
	shop, err := f.store.GetShop(ctx, id)
	if err != nil {
		return shops.Shop{}, err
	}
	return toCoreShop(shop), nil
}

func (f *Storer) QueryAll(ctx context.Context, userID string) ([]shops.Shop, error) {
	dshops, err := f.store.ListShops(ctx, userID)
	if err != nil {
		return []shops.Shop{}, err
	}
	shops := make([]shops.Shop, len(dshops))
	for i, s := range dshops {
		shops[i] = toCoreShop(s)
	}
	return shops, nil
}

func toCoreShop(s db.Shop) shops.Shop {
	return shops.Shop{
		ID:     s.ID,
		UserID: s.UserID,
		Name:   s.Name,
	}
}
