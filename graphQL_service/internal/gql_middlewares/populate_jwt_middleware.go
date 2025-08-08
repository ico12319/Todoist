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

	if p.next == nil {
		log.C(ctx).Info("nil e bee")
	}

	p.next.ServeHTTP(w, r.WithContext(ctx))
	log.C(ctx).Info("pisna mi bate")
}

func NewJwtPopulateMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return newPopulateJwtMiddleware(next)
	}
}
