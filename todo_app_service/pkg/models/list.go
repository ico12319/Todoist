package models

import (
	"Todo-List/internProject/todo_app_service/pkg/pagination"
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

type ListPage struct {
	Data       []*List          `json:"data"`
	PageInfo   *pagination.Page `json:"page_info"`
	TotalCount int              `json:"total_count"`
}
