package gql_middlewares

import (
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"github.com/gorilla/mux"
	"net/http"
)

type corsMiddleware struct {
	next        http.Handler
	frontendUrl string
}

func newCorsMiddleware(next http.Handler, frontendUrl string) *corsMiddleware {
	return &corsMiddleware{
		next:        next,
		frontendUrl: frontendUrl,
	}
}

func (c *corsMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", c.frontendUrl)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PATCH, DELETE, PUT")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Vary", "Origin")

	if c.next == nil {
		log.C(r.Context()).Info("nil e bee basiii")
	}

	log.C(r.Context()).Info("plsssss")
	c.next.ServeHTTP(w, r)
	log.C(r.Context()).Info("pak bugove")
}

func CorsMiddlewareFunc(frontendUrl string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return newCorsMiddleware(next, frontendUrl)
	}
}
