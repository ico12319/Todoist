package models

import (
	"time"
)

type List struct {
	Id          string    `json:"id"`
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description" validate:"required"`
	CreatedAt   time.Time `json:"created_at"`
	LastUpdated time.Time `json:"last_updated"`
	Owner       string    `json:"owner" validate:"required"`
}
