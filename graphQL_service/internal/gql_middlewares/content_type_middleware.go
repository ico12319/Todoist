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

	if c.next == nil {
		log.C(ctx).Info("nil e bee ei tuka ei tuiii")
	}

	log.C(ctx).Info("nema kak da e tuka")
	c.next.ServeHTTP(w, r)
	log.C(ctx).Info("plssdasdaf")
}

func ContentTypeMiddlewareFunc() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return newContentTypeMiddleware(next)
	}
}
