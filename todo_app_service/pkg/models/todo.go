package models

import (
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"Todo-List/internProject/todo_app_service/pkg/pagination"
	"time"
)

type Todo struct {
	Id          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	ListId      string               `json:"list_id"`
	Status      constants.TodoStatus `json:"status"`
	CreatedAt   time.Time            `json:"created_at"`
	LastUpdated time.Time            `json:"last_updated"`
	Priority    constants.Priority   `json:"priority"`
	AssignedTo  *string              `json:"assigned_to,omitempty"`
	DueDate     *time.Time           `json:"due_date,omitempty"`
}

type TodoPage struct {
	Data       []*Todo          `json:"data"`
	PageInfo   *pagination.Page `json:"page_info"`
	TotalCount int              `json:"total_count"`
}
