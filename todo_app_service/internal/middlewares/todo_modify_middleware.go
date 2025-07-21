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

type tService interface {
	GetTodoRecord(context.Context, string) (*models.Todo, error)
	GetTodoAssigneeToRecord(context.Context, string) (*models.User, error)
}

type lService interface {
	GetListOwnerRecord(context.Context, string) (*models.User, error)
	GetCollaborators(context.Context, *filters.ListFilters) ([]*models.User, error)
}

type todoModifyMiddleware struct {
	next        http.Handler
	todoService tService
	lService    lService
}

func newTodoModifyMiddleware(next http.Handler, service tService, lService lService) *todoModifyMiddleware {
	return &todoModifyMiddleware{next: next, todoService: service, lService: lService}
}

func (t *todoModifyMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	todoId, ok := r.Context().Value(TodoId).(string)
	if !ok {
		utils.EncodeError(w, "missing todo_id", http.StatusBadRequest)
		return
	}

	user, ok := r.Context().Value(UserKey).(*models.User)
	if !ok {
		utils.EncodeError(w, "missing user in context", http.StatusBadRequest)
		return
	}

	todo, err := t.todoService.GetTodoRecord(ctx, todoId)
	if err != nil {
		log.C(ctx).Errorf("failed to get todo record, error %s when trying to get it", err.Error())

		utils.EncodeErrorWithCorrectStatusCode(w, err)
		return
	}

	listOwner, err := t.lService.GetListOwnerRecord(ctx, todo.ListId)
	if err != nil {
		log.C(ctx).Errorf("failed to serve http, error %s when trying to get list owner of list with id %s", err.Error(), todo.ListId)

		utils.EncodeErrorWithCorrectStatusCode(w, err)
		return
	}

	collaborators, err := t.lService.GetCollaborators(ctx, &filters.ListFilters{
		ListId: todo.ListId,
	})

	if err != nil {
		log.C(ctx).Errorf("failed to serve http, error %s when trying to get list with id %s collaborators", err.Error(), todo.ListId)

		utils.EncodeErrorWithCorrectStatusCode(w, err)
		return
	}

	assignee, err := t.todoService.GetTodoAssigneeToRecord(r.Context(), todoId)
	if err != nil {
		log.C(ctx).Errorf("faield to serve http, error %s when trying to get todo assingee", err.Error())

		utils.EncodeErrorWithCorrectStatusCode(w, err)
		return
	}

	if !determineWhetherUserHasAccess(assignee, listOwner, user, collaborators) {
		utils.EncodeError(
			w,
			"forbidden: only administrators, collaborators, list owners, or the assignee may access todo",
			http.StatusForbidden,
		)
		return
	}
	t.next.ServeHTTP(w, r)
}

func NewTodoModifyMiddlewareFunc(todoService tService, listService lService) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return newTodoModifyMiddleware(next, todoService, listService)
	}
}

func isCollaborator(collaborators []*models.User, user *models.User) bool {
	for _, collab := range collaborators {
		if collab.Id == user.Id {
			return true
		}
	}
	return false
}

func determineAssigneeFlag(assignee *models.User, candidate *models.User) bool {
	var flag bool = true

	if assignee == nil {
		flag = false
	}

	if assignee != nil && candidate.Id != assignee.Id {
		flag = false
	}

	return flag
}

func determineWhetherUserHasAccess(assignee *models.User, listOwner *models.User, user *models.User, collaborators []*models.User) bool {
	flag := determineAssigneeFlag(assignee, user)

	if !flag && user.Role != constants.Admin && user.Id != listOwner.Id && !isCollaborator(collaborators, user) {
		return false
	}

	return true
}
