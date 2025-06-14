package middlewares

import (
	"context"
	"errors"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/utils"
	"github.com/gorilla/mux"
	"net/http"
)

type todoIdKey struct{}

var TodoId = todoIdKey{}

type extractionTodoIdMiddleware struct {
	next http.Handler
}

func newExtractionTodoIdMiddleware(next http.Handler) *extractionTodoIdMiddleware {
	return &extractionTodoIdMiddleware{next: next}
}

func (e *extractionTodoIdMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	todoId, ok := params["todo_id"]
	if !ok {
		utils.EncodeError(w, errors.New("invalid request: missing todo_id").Error(), http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	ctx = context.WithValue(ctx, TodoId, todoId)
	e.next.ServeHTTP(w, r.WithContext(ctx))
}

func ExtractionTodoIdMiddlewareFunc(next http.Handler) http.Handler {
	return newExtractionTodoIdMiddleware(next)
}
