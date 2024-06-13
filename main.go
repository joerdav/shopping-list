package main

import (
	"cmp"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/joerdav/shopping-list/app"
)

//go:embed migrations/*.sql
var fs embed.FS

func run() error {
	_ = godotenv.Load()
	db, err := sql.Open("sqlite3", cmp.Or(os.Getenv("DB"), ":memory:"))
	if err != nil {
		return err
	}
	d, err := iofs.New(fs, "migrations")
	if err != nil {
		return err
	}
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("iofs", d, "mydb", driver)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	s := app.NewServer(db)
	slog.Info("Listening on http://0.0.0.0:8080")
	return http.ListenAndServe("0.0.0.0:8080", s)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
