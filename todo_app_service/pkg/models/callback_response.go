package models

type CallbackResponse struct {
	JwtToken     string `json:"jwt_token"`
	RefreshToken string `json:"refresh_token"`
}
