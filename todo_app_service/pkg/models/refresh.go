package models

type Refresh struct {
	RefreshToken string `json:"refresh_token"`
	UserId       string `json:"user_id"`
}
