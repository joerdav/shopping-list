package main

import (
	"cmp"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"

	"github.com/joerdav/shopping-list/app"
	"github.com/joerdav/shopping-list/business/items"
	"github.com/joerdav/shopping-list/business/items/itemsdb"
	"github.com/joerdav/shopping-list/business/lists"
	"github.com/joerdav/shopping-list/business/lists/listsdb"
	"github.com/joerdav/shopping-list/business/recipes"
	"github.com/joerdav/shopping-list/business/recipes/recipesdb"
	"github.com/joerdav/shopping-list/business/shops"
	"github.com/joerdav/shopping-list/business/shops/shopsdb"
)

func run() error {
	if err := godotenv.Load(); err != nil {
		return err
	}
	db, err := sql.Open("sqlite3", cmp.Or(os.Getenv("DB"), ":memory:"))
	if err != nil {
		return err
	}
	if err := migrate(db); err != nil {
		slog.Warn("Migration error", "error", err)
	}
	if os.Getenv("SEED") != "" {
		if err := seed(db); err != nil {
			return err
		}
	}
	s := app.NewServer(db)
	slog.Info("Listening on http://localhost:8080")
	return http.ListenAndServe("localhost:8080", s)
}

func migrate(db *sql.DB) error {
	return errors.Join(
		shopsdb.NewStorer(db).Migrate(context.Background()),
		itemsdb.NewStorer(db).Migrate(context.Background()),
		recipesdb.NewStorer(db).Migrate(context.Background()),
		listsdb.NewStorer(db).Migrate(context.Background()),
	)
}

func seed(db *sql.DB) error {
	shopsCore := shops.NewCore(shopsdb.NewStorer(db))
	itemsCore := items.NewCore(itemsdb.NewStorer(db))
	recipesCore := recipes.NewCore(recipesdb.NewStorer(db))
	listsCore := lists.NewCore(listsdb.NewStorer(db))

	// return if already shops exist
	shopsList, err := shopsCore.QueryAll(context.Background())
	if err != nil {
		return err
	}
	if len(shopsList) > 0 {
		slog.Info("Shops already exist, skipping seed")
		return nil
	}

	tesco, err := shopsCore.Create(context.Background(), shops.NewShop{Name: "Tesco"})
	if err != nil {
		return err
	}
	morrisons, err := shopsCore.Create(context.Background(), shops.NewShop{Name: "Morrisons"})
	if err != nil {
		return err
	}

	cashews, err := itemsCore.Create(context.Background(), items.NewItem{Name: "Cashews", ShopID: tesco.ID})
	if err != nil {
		return err
	}
	chicken, err := itemsCore.Create(context.Background(), items.NewItem{Name: "Chicken", ShopID: morrisons.ID})
	if err != nil {
		return err
	}
	burgerBuns, err := itemsCore.Create(context.Background(), items.NewItem{Name: "Burger Buns", ShopID: tesco.ID})
	if err != nil {
		return err
	}
	burgers, err := itemsCore.Create(context.Background(), items.NewItem{Name: "Burgers", ShopID: tesco.ID})
	if err != nil {
		return err
	}

	cashewChicken, err := recipesCore.Create(context.Background(), recipes.NewRecipe{Name: "Cashew Chicken"})
	if err != nil {
		return err
	}
	cashewChicken.Ingredients[cashews.ID] = 1
	cashewChicken.Ingredients[chicken.ID] = 1
	cashewChicken, err = recipesCore.Update(context.Background(), recipes.UpdateRecipe{ID: cashewChicken.ID, Ingredients: &cashewChicken.Ingredients})
	if err != nil {
		return err
	}

	burger, err := recipesCore.Create(context.Background(), recipes.NewRecipe{Name: "Burger"})
	if err != nil {
		return err
	}
	burger.Ingredients[burgers.ID] = 1
	burger.Ingredients[burgerBuns.ID] = 2
	burger, err = recipesCore.Update(context.Background(), recipes.UpdateRecipe{ID: burger.ID, Ingredients: &burger.Ingredients})
	if err != nil {
		return err
	}

	list, err := listsCore.Create(context.Background(), lists.NewList{CreatedDate: time.Now()})
	if err != nil {
		return err
	}
	list.Recipes[cashewChicken.ID] = 1
	list.Recipes[burger.ID] = 1
	list.Items[burgerBuns.ID] = 1
	_, err = listsCore.Update(context.Background(), lists.UpdateList{ID: list.ID, Recipes: &list.Recipes, Items: &list.Items})
	if err != nil {
		return err
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
