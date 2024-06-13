package usersdb

import (
	"context"
	"database/sql"
	"errors"

	"github.com/joerdav/shopping-list/business/users"
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
	return toCoreUser(db.User{user}), nil
}

func toCoreUser(s db.User) users.User {
	return users.User{
		ID: s.ID,
	}
}
