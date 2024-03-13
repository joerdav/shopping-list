package lists

import (
	"time"

	"github.com/google/uuid"
)

type NewList struct {
	CreatedDate time.Time
	UserID      string
}

type List struct {
	ID          uuid.UUID
	UserID      string
	CreatedDate time.Time
	Recipes     map[uuid.UUID]int
	Items       map[uuid.UUID]int
}

type UpdateList struct {
	ID      uuid.UUID
	UserID  string
	Recipes *map[uuid.UUID]int
	Items   *map[uuid.UUID]int
}
