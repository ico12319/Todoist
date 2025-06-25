package models

import (
	"Todo-List/internProject/todo_app_service/pkg/constants"
)

type User struct {
	Id    string             `json:"id"`
	Email string             `json:"email" validate:"required"`
	Role  constants.UserRole `json:"role" validate:"required"`
}
