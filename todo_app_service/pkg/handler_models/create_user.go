package handler_models

type CreateUser struct {
	Email string `json:"email" validate:"required"`
	Role  string `json:"role" validate:"required"`
}
