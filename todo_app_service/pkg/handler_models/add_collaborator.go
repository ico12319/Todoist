package handler_models

type AddCollaborator struct {
	Id string `json:"user_id" validate:"required"`
}
