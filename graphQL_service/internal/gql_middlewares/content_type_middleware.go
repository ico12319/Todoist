package gql_middlewares

import (
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"github.com/gorilla/mux"
	"net/http"
)

type contentTypeMiddleware struct {
	next http.Handler
}

func newContentTypeMiddleware(next http.Handler) *contentTypeMiddleware {
	return &contentTypeMiddleware{next: next}
}

func (c *contentTypeMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("setting content type to JSON in gql_service")

	w.Header().Set("Content-Type", constants.CONTENT_TYPE)
	c.next.ServeHTTP(w, r)
}

func ContentTypeMiddlewareFunc() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return newContentTypeMiddleware(next)
	}
}
