package middlewares

import (
	"Todo-List/internProject/todo_app_service/internal/utils"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"net/http"
)

type userAccessMiddleware struct {
	next http.Handler
}

func newUserAccessMiddleware(next http.Handler) *userAccessMiddleware {
	return &userAccessMiddleware{next: next}
}

func (u *userAccessMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId, ok := ctx.Value(UserId).(string)
	if !ok {
		log.C(ctx).Error("failed to serve http, nil value userId in context...")
		utils.EncodeError(w, "nil userId in context", http.StatusInternalServerError)
		return
	}

	user, ok := ctx.Value(UserKey).(*models.User)
	if !ok {
		log.C(ctx).Error("failed to serve http, nil value user in context...")
		utils.EncodeError(w, "nil user in context", http.StatusInternalServerError)
		return
	}

	if user.Id != userId && user.Role != constants.Admin {
		log.C(ctx).Info("user trying to access user that is not him...")
		utils.EncodeError(w, "you can access/modify only your profile", http.StatusForbidden)
		return
	}

	u.next.ServeHTTP(w, r)
}

func UserAccessMiddlewareFunc(next http.Handler) http.Handler {
	return newUserAccessMiddleware(next)
}
