package handler_models

import (
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"time"
)

type UpdateTodo struct {
	Name        *string               `json:"name" validate:"omitempty,min=1"`
	Description *string               `json:"description,omitempty" validate:"omitempty,min=1"`
	Status      *constants.TodoStatus `json:"status,omitempty" validate:"omitempty,min=1"`
	Priority    *constants.Priority   `json:"priority,omitempty" validate:"omitempty,min=1"`
	AssignedTo  *string               `json:"assigned_to,omitempty" validate:"omitempty,min=1"`
	DueDate     *time.Time            `json:"due_date,omitempty" validate:"omitempty,gte"`
}
