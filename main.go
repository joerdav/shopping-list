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

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"

	"github.com/joerdav/shopping-list/app"
	"github.com/joerdav/shopping-list/business/items/itemsdb"
	"github.com/joerdav/shopping-list/business/lists/listsdb"
	"github.com/joerdav/shopping-list/business/recipes/recipesdb"
	"github.com/joerdav/shopping-list/business/shops/shopsdb"
	"github.com/joerdav/shopping-list/business/users/usersdb"
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
		usersdb.NewStorer(db).Migrate(context.Background()),
	)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
