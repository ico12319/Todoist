package gql_middlewares

import (
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
	"github.com/gorilla/mux"
	"net/http"
)

type authToken struct{}

var AuthToken = authToken{}

type populateJwtMiddleware struct {
	next http.Handler
}

func newPopulateJwtMiddleware(next http.Handler) *populateJwtMiddleware {
	return &populateJwtMiddleware{next: next}
}

func (p *populateJwtMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("population jwt auth token in gql middleware")

	jwt := r.Header.Get("Authorization")
	ctx = context.WithValue(ctx, AuthToken, jwt)
	p.next.ServeHTTP(w, r.WithContext(ctx))
}

func NewJwtPopulateMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return newPopulateJwtMiddleware(next)
	}
}
