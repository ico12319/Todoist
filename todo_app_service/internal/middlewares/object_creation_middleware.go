package middlewares

import (
	"Todo-List/internProject/todo_app_service/internal/utils"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"net/http"
)

type objectCreationMiddleware struct {
	next http.Handler
}

func newObjectCreationMiddleware(next http.Handler) *objectCreationMiddleware {
	return &objectCreationMiddleware{next: next}
}

// only admins and writers can create new entities!
func (l *objectCreationMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(UserKey).(*models.User)
	if user.Role != constants.Admin && user.Role != constants.Writer {
		utils.EncodeError(w, "only admins and writers can create or modify entities", http.StatusForbidden)
		return
	}
	l.next.ServeHTTP(w, r)
}

func ObjectCreationMiddlewareFunc(next http.Handler) http.Handler {
	return newObjectCreationMiddleware(next)
}
