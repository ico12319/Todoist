package tokens

import "github.com/golang-jwt/jwt/v5"

type tokenGetter struct{}

func NewJwtGetter() *tokenGetter {
	return &tokenGetter{}
}

func (*tokenGetter) GetSignedJWT(jwt *jwt.Token, key interface{}) (string, error) {
	return jwt.SignedString(key)
}

func (*tokenGetter) GetJWTWithClaims(method jwt.SigningMethod, claims jwt.Claims) *jwt.Token {
	return jwt.NewWithClaims(method, claims)
}
