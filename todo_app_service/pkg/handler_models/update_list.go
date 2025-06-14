package handler_models

type UpdateList struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1"`
	Description *string `json:"description,omitempty" validate:"omitempty,min=1"`
}
