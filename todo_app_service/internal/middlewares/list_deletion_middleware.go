package middlewares

import (
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/utils"
	log "github.com/I763039/Todo-List/internProject/todo_app_service/pkg/configuration"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/constants"
	"net/http"
)

type listDeletionMiddleware struct {
	next http.Handler
}

func newListDeletionMiddleware(next http.Handler) *listDeletionMiddleware {
	return &listDeletionMiddleware{next: next}
}

func (l *listDeletionMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userRole, ok := r.Context().Value(UserRoleKey).(userRoleKey)
	if !ok {
		log.C(ctx).Error("failed to retrieve userRole from context...")
		utils.EncodeError(w, "nil value in context", http.StatusInternalServerError)
		return
	}

	if userRole.role != constants.Admin && !userRole.isOwner {
		utils.EncodeError(w, "only admins and list owner can delete list", http.StatusForbidden)
		return
	}
	l.next.ServeHTTP(w, r)
}

func ListDeletionMiddlewareFunc(next http.Handler) http.Handler {
	return newListDeletionMiddleware(next)
}
