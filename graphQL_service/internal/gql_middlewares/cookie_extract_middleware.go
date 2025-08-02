package gql_middlewares

import (
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"github.com/gorilla/mux"
	"net/http"
	"unicode/utf8"
)

type cookieExtractMiddleware struct {
	next http.Handler
}

func newCookieExtractMiddleware(next http.Handler) *cookieExtractMiddleware {
	return &cookieExtractMiddleware{next: next}
}

func (c *cookieExtractMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("extracting cookie value in cookie extract middleware")

	cookie, err := r.Cookie("access-token")
	if err != nil {
		log.C(ctx).Warn("cookie with name access-token is not provided...")
	}

	if cookie != nil && utf8.RuneCountInString(cookie.Value) != 0 {
		r.Header.Set("Authorization", cookie.Value)
	}

	c.next.ServeHTTP(w, r)
}

func CookieExtractMiddlewareFunc() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return newCookieExtractMiddleware(next)
	}
}
