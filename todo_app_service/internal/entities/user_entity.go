package entities

import "github.com/gofrs/uuid"

type User struct {
	Id         uuid.UUID `db:"id"`
	Email      string    `db:"email"`
	Role       string    `db:"role"`
	TotalCount int       `db:"total_count"`
}
