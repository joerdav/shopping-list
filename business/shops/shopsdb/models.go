// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package shopsdb

import (
	"github.com/google/uuid"
)

type Shop struct {
	ID   uuid.UUID
	Name string
}