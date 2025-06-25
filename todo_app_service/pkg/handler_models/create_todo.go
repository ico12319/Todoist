package handler_models

import (
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"time"
)

type CreateTodo struct {
	Name        string             `json:"name" validate:"required"`
	Description string             `json:"description" validate:"required"`
	ListId      string             `json:"list_id" validate:"required"`
	Priority    constants.Priority `json:"priority" validate:"required"`
	AssignedTo  *string            `json:"assigned_to,omitempty" validate:"omitempty,min=1"`
	DueDate     *time.Time         `json:"due_date,omitempty" validate:"omitempty,gte"`
}
