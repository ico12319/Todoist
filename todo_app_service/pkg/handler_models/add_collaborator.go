package handler_models

type AddCollaborator struct {
	Email string `json:"email" validate:"required"`
}
