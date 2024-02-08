package lists

import (
	"time"

	"github.com/google/uuid"
)

type NewList struct {
	CreatedDate time.Time
}

type List struct {
	ID          uuid.UUID
	CreatedDate time.Time
	Recipes     map[uuid.UUID]int
	Items       map[uuid.UUID]int
}

type UpdateList struct {
	ID      uuid.UUID
	Recipes *map[uuid.UUID]int
	Items   *map[uuid.UUID]int
}
