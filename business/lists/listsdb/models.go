// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package listsdb

import (
	"github.com/google/uuid"
)

type List struct {
	ID          uuid.UUID
	CreatedDate int64
}

type ListItem struct {
	ItemID   uuid.UUID
	ListID   uuid.UUID
	Quantity int64
}

type ListRecipe struct {
	RecipeID uuid.UUID
	ListID   uuid.UUID
	Quantity int64
}
