package handler_models

type UpdateUser struct {
	Email *string `json:"email" validate:"omitempty,min=1"`
	Role  *string `json:"role"   validate:"omitempty,min=1"`
}
