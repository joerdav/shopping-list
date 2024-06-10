package listsdb

import (
	"context"
	"database/sql"
	_ "embed"
	"time"

	"github.com/google/uuid"
	"github.com/joerdav/shopping-list/business/lists"
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

func (f *Storer) Create(ctx context.Context, list lists.List) error {
	return f.store.CreateList(ctx, CreateListParams{
		ID:          list.ID,
		UserID:      list.UserID,
		CreatedDate: list.CreatedDate.Unix(),
	})
}

func (f *Storer) Delete(ctx context.Context, listID uuid.UUID) error {
	tx, err := f.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	qtx := f.store.WithTx(tx)
	if err := qtx.DeleteItemsByList(ctx, listID); err != nil {
		return err
	}
	if err := qtx.DeleteRecipesByList(ctx, listID); err != nil {
		return err
	}
	if err := qtx.DeleteList(ctx, listID); err != nil {
		return err
	}
	return tx.Commit()
}

func (f *Storer) Update(ctx context.Context, list lists.List) error {
	recipes, err := f.store.GetRecipesByList(ctx, list.ID)
	if err != nil {
		return err
	}
	items, err := f.store.GetItemsByList(ctx, list.ID)
	if err != nil {
		return err
	}
	tx, err := f.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	qtx := f.store.WithTx(tx)
	// Update recipes
	for recipeID, quantity := range list.Recipes {
		if err := qtx.SetRecipe(ctx, SetRecipeParams{
			ListID:   list.ID,
			RecipeID: recipeID,
			Quantity: int64(quantity),
		}); err != nil {
			return err
		}
	}
	// Delete recipes
	for _, recipes := range recipes {
		if _, ok := list.Recipes[recipes.RecipeID]; !ok {
			if err := qtx.DeleteRecipe(ctx, DeleteRecipeParams{
				ListID:   list.ID,
				RecipeID: recipes.RecipeID,
			}); err != nil {
				return err
			}
		}
	}
	// Update items
	for itemID, quantity := range list.Items {
		if err := qtx.SetItem(ctx, SetItemParams{
			ListID:   list.ID,
			ItemID:   itemID,
			Quantity: int64(quantity),
		}); err != nil {
			return err
		}
	}
	// Delete items
	for _, item := range items {
		if _, ok := list.Items[item.ItemID]; !ok {
			if err := qtx.DeleteItem(ctx, DeleteItemParams{
				ListID: list.ID,
				ItemID: item.ItemID,
			}); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (f *Storer) Query(ctx context.Context, id uuid.UUID) (lists.List, error) {
	list, err := f.store.GetList(ctx, id)
	if err != nil {
		return lists.List{}, err
	}
	items, err := f.store.GetItemsByList(ctx, id)
	if err != nil {
		return lists.List{}, err
	}
	recipes, err := f.store.GetRecipesByList(ctx, id)
	if err != nil {
		return lists.List{}, err
	}
	return toCoreList(list, items, recipes), nil
}

func (f *Storer) QueryAll(ctx context.Context, userID string) ([]lists.List, error) {
	listsList, err := f.store.GetAllLists(ctx, userID)
	if err != nil {
		return nil, err
	}
	var coreLists []lists.List
	for _, list := range listsList {
		items, err := f.store.GetItemsByList(ctx, list.ID)
		if err != nil {
			return nil, err
		}
		recipes, err := f.store.GetRecipesByList(ctx, list.ID)
		if err != nil {
			return nil, err
		}
		coreLists = append(coreLists, toCoreList(list, items, recipes))
	}
	return coreLists, nil
}

func toCoreList(list List, items []ListItem, recipes []ListRecipe) lists.List {
	coreList := lists.List{
		ID:          list.ID,
		UserID:      list.UserID,
		CreatedDate: time.Unix(list.CreatedDate, 0),
		Items:       make(map[uuid.UUID]int),
		Recipes:     make(map[uuid.UUID]int),
	}
	for _, item := range items {
		coreList.Items[item.ItemID] = int(item.Quantity)
	}
	for _, recipe := range recipes {
		coreList.Recipes[recipe.RecipeID] = int(recipe.Quantity)
	}
	return coreList
}
