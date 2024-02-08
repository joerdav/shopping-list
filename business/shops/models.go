package shops

import "github.com/google/uuid"

type NewShop struct {
	Name string
}

type Shop struct {
	ID   uuid.UUID
	Name string
}
