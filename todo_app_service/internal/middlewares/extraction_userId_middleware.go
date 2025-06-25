package middlewares

import (
	"Todo-List/internProject/todo_app_service/internal/utils"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"context"
	"github.com/gorilla/mux"
	"net/http"
)

type userIdKey struct{}

var UserId = userIdKey{}

type extractionUserIdMiddleware struct {
	next http.Handler
}

func newExtractionUserIdMiddleware(next http.Handler) *extractionUserIdMiddleware {
	return &extractionUserIdMiddleware{next: next}
}

func (e *extractionUserIdMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userId, ok := params["user_id"]
	if !ok {
		utils.EncodeError(w, constants.MISSING_USER_ID, http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, UserId, userId)
	e.next.ServeHTTP(w, r.WithContext(ctx))
}

func ExtractionUserIdMiddlewareFunc(next http.Handler) http.Handler {
	return newExtractionUserIdMiddleware(next)
}
