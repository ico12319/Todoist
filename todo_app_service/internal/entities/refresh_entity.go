package entities

import "github.com/gofrs/uuid"

type Refresh struct {
	RefreshToken string    `db:"refresh_token"`
	UserId       uuid.UUID `db:"user_id"`
}
