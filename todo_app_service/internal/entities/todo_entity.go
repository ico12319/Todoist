package entities

import (
	"database/sql"
	"github.com/gofrs/uuid"
	"time"
)

type Todo struct {
	Id          uuid.UUID     `db:"id"`
	Name        string        `db:"name"`
	Description string        `db:"description"`
	ListId      uuid.UUID     `db:"list_id"`
	Status      string        `db:"status"`
	CreatedAt   time.Time     `db:"created_at"`
	LastUpdated time.Time     `db:"last_updated"`
	AssignedTo  uuid.NullUUID `db:"assigned_to"`
	DueDate     sql.NullTime  `db:"due_date"`
	Priority    string        `db:"priority"`
	TotalCount  int           `db:"total_count"`
}
