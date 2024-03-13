package usersdb

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"

	"github.com/joerdav/shopping-list/business/users"
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

func (f *Storer) Create(ctx context.Context, user users.User) error {
	return f.store.CreateUser(ctx, user.ID)
}

func (f *Storer) Query(ctx context.Context, id string) (users.User, error) {
	user, err := f.store.GetUser(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return users.User{}, users.ErrNotFound
	}
	if err != nil {
		return users.User{}, err
	}
	return toCoreUser(User{user}), nil
}

func toCoreUser(s User) users.User {
	return users.User{
		ID: s.ID,
	}
}
