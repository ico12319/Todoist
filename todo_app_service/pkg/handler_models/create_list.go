package handler_models

type CreateList struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
}
