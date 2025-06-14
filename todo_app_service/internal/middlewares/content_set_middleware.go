package middlewares

import (
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/constants"
	"net/http"
)

type contentTypeMiddleware struct {
	next http.Handler
}

func newContentTypeMiddleware(next http.Handler) *contentTypeMiddleware {
	return &contentTypeMiddleware{next: next}
}

func (c *contentTypeMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", constants.CONTENT_TYPE)
	c.next.ServeHTTP(w, r)
}

func ContentTypeMiddlewareFunc(next http.Handler) http.Handler {
	return newContentTypeMiddleware(next)
}
