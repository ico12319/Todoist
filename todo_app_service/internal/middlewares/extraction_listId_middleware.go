package middlewares

import (
	"Todo-List/internProject/todo_app_service/internal/utils"
	"context"
	"errors"
	"github.com/gorilla/mux"
	"net/http"
)

type listIdKeyType struct{}

var ListId = listIdKeyType{}

type extractionListIdMiddleware struct {
	next http.Handler
}

func newExtractionListIdMiddleware(next http.Handler) *extractionListIdMiddleware {
	return &extractionListIdMiddleware{next: next}
}

func (e *extractionListIdMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	listId, ok := params["list_id"]
	if !ok {
		utils.EncodeError(w, errors.New("invalid request: missing list_id").Error(), http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	ctx = context.WithValue(ctx, ListId, listId)
	e.next.ServeHTTP(w, r.WithContext(ctx))
}

func ExtractionListIdMiddlewareFunc(next http.Handler) http.Handler {
	return newExtractionListIdMiddleware(next)
}
