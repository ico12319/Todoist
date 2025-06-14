package tokens

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	UserId string `json:"id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}
