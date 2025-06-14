package entities

import (
	"github.com/gofrs/uuid"
	"time"
)

type List struct {
	Id          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	CreatedAt   time.Time `db:"created_at"`
	LastUpdated time.Time `db:"last_updated"`
	Owner       uuid.UUID `db:"owner"`
	Description string    `db:"description"`
}
