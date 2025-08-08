package entities

import "github.com/gofrs/uuid"

type PaginationInfo struct {
	FirstID    uuid.NullUUID `db:"first_id"`
	LastID     uuid.NullUUID `db:"last_id"`
	TotalCount int           `db:"total_count"`
}
