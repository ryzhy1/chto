package dto

import "github.com/google/uuid"

type Friend struct {
	Id       uuid.UUID `json:"id"`
	Username string    `json:"username" db:"username"`
}
