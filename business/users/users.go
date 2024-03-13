package users

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("user not found")

type Storer interface {
	Create(context.Context, User) error
	Query(context.Context, string) (User, error)
}

type Core struct {
	storer Storer
}

func NewCore(storer Storer) *Core {
	return &Core{storer}
}

func (c *Core) Create(ctx context.Context, s User) (User, error) {
	if err := c.storer.Create(ctx, s); err != nil {
		return User{}, err
	}
	return s, nil
}

func (c *Core) Query(ctx context.Context, id string) (User, error) {
	return c.storer.Query(ctx, id)
}
