package models

import (
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"Todo-List/internProject/todo_app_service/pkg/pagination"
)

type User struct {
	Id    string             `json:"id"`
	Email string             `json:"email" validate:"required"`
	Role  constants.UserRole `json:"role" validate:"required"`
}

type UserPage struct {
	Data       []*User          `json:"data"`
	PageInfo   *pagination.Page `json:"page_info"`
	TotalCount int              `json:"total_count"`
}
