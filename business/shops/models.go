package shops

import "github.com/google/uuid"

type NewShop struct {
	Name   string
	UserID string
}

type Shop struct {
	ID     uuid.UUID
	Name   string
	UserID string
}
