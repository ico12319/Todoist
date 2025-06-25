package middlewares

import (
	"Todo-List/internProject/todo_app_service/internal/utils"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"net/http"
)

type usersGlobalAccessMiddleware struct {
	next http.Handler
}

func newGlobalAccessMiddleware(next http.Handler) *usersGlobalAccessMiddleware {
	return &usersGlobalAccessMiddleware{next: next}
}

// only admins can see all lists, all todos and all users!
func (g *usersGlobalAccessMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := ctx.Value(UserKey).(*models.User)

	if !ok {
		log.C(ctx).Error("empty user in context, error...")
		utils.EncodeError(w, "nil user value in context", http.StatusInternalServerError)
		return
	}

	if user.Role != constants.Admin {
		log.C(ctx).Infof("user trying to access entities is not an admin..., role %s", user.Role)
		utils.EncodeError(w, "only admins can access all lists", http.StatusForbidden)
		return
	}

	g.next.ServeHTTP(w, r)
}

func GlobalAccessMiddlewareFunc(next http.Handler) http.Handler {
	return newGlobalAccessMiddleware(next)
}
