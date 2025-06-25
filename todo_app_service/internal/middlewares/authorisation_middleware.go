package middlewares

import (
	"Todo-List/internProject/todo_app_service/internal/utils"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"Todo-List/internProject/todo_app_service/pkg/tokens"
	"context"
	"github.com/gorilla/mux"
	"net/http"
)

type ctxKeyType struct{}

var UserKey = ctxKeyType{}

type UserService interface {
	GetUserRecordByEmail(context.Context, string) (*models.User, error)
}

type jwtParser interface {
	ParseJWT(context.Context, string) (*tokens.Claims, error)
}

type authorisationMiddleware struct {
	next        http.Handler
	service     UserService
	tokenParser jwtParser
}

func newAuthorisationMiddleware(next http.Handler, service UserService, tokenParser jwtParser) *authorisationMiddleware {
	return &authorisationMiddleware{next: next, service: service, tokenParser: tokenParser}
}

func (a *authorisationMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	authHeader := r.Header.Get("Authorization")

	jwtClaims, err := a.tokenParser.ParseJWT(ctx, authHeader)
	if err != nil {
		log.C(ctx).Errorf("failed to parse JWT, error %s", err.Error())
		utils.EncodeError(w, utils.DetermineCorrectJwtErrorMessage(err), http.StatusUnauthorized)
		return
	}

	user, err := a.service.GetUserRecordByEmail(ctx, jwtClaims.Email)
	if err != nil {
		log.C(ctx).Errorf("failed to get user by email, internal server error... %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx = context.WithValue(ctx, UserKey, user)
	a.next.ServeHTTP(w, r.WithContext(ctx))
}

func AuthorisationMiddlewareFunc(service UserService, tokenParser jwtParser) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return newAuthorisationMiddleware(next, service, tokenParser)
	}
}
