package middlewares

import (
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators/filters"
	"Todo-List/internProject/todo_app_service/internal/utils"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"context"
	"github.com/gorilla/mux"
	"net/http"
)

type userRoleKey struct {
	role    constants.UserRole
	isOwner bool
}

var UserRoleKey = userRoleKey{}

type ListService interface {
	GetListRecord(context.Context, string) (*models.List, error)
	GetCollaborators(context.Context, *filters.ListFilters) ([]*models.User, error)
}

type listModifyMiddleware struct {
	next    http.Handler
	serv    ListService
	factory statusCodeEncoderFactory
}

func newListAccessMiddleware(next http.Handler, serv ListService, factory statusCodeEncoderFactory) *listModifyMiddleware {
	return &listModifyMiddleware{next: next, serv: serv, factory: factory}
}

func (a *listModifyMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctxUser, ok := r.Context().Value(UserKey).(*models.User)
	if !ok {
		log.C(ctx).Error("failed to serve http, empty user in context...")
		utils.EncodeError(w, "nil value in context", http.StatusInternalServerError)
		return
	}

	listId, ok := r.Context().Value(ListId).(string)
	if !ok {
		log.C(ctx).Error("failed to serve http, empty list_id in context...")
		utils.EncodeError(w, "nil value in context", http.StatusInternalServerError)
		return
	}

	list, err := a.serv.GetListRecord(r.Context(), listId)
	if err != nil {
		log.C(ctx).Errorf("failed to serve http, error %s when trying to get list with id %s", err.Error(), list)

		encoder := a.factory.CreateStatusCodeEncoder(ctx, w, err)
		encoder.EncodeErrorWithCorrectStatusCode(ctx, w)
		return
	}

	if ctxUser.Id == list.Owner {
		configUserRoleKey(ctxUser.Role, true, a.next, w, r)
		return
	}

	if ctxUser.Role == constants.Admin {
		configUserRoleKey(ctxUser.Role, false, a.next, w, r)
		return
	}

	authUsers, err := a.serv.GetCollaborators(r.Context(), &filters.ListFilters{
		ListId: listId,
	})

	if err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, user := range authUsers {
		if ctxUser.Id == user.Id {
			configUserRoleKey(ctxUser.Role, false, a.next, w, r)
			return
		}
	}

	utils.EncodeError(w, "access forbidden: only administrators, collaborators or the list creator, may create/modify list related entity", http.StatusForbidden)
}

func ListAccessPermissionMiddlewareFunc(serv ListService, facotry statusCodeEncoderFactory) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return newListAccessMiddleware(next, serv, facotry)
	}
}

func configUserRoleKey(userRole constants.UserRole, isOwner bool, next http.Handler, w http.ResponseWriter, r *http.Request) {
	uRole := userRoleKey{
		role:    userRole,
		isOwner: isOwner,
	}
	ctx := context.WithValue(r.Context(), UserRoleKey, uRole)
	next.ServeHTTP(w, r.WithContext(ctx))
}
