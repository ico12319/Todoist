package middlewares

import (
	"Todo-List/internProject/todo_app_service/internal/persistence"
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
	GetCollaborators(context.Context, string, *filters.BaseFilters) (*models.UserPage, error)
}

type todoModifyMiddleware struct {
	next        http.Handler
	todoService tService
	lService    lService
	transact    persistence.Transactioner
}

func newTodoModifyMiddleware(next http.Handler, service tService, lService lService, transact persistence.Transactioner) *todoModifyMiddleware {
	return &todoModifyMiddleware{
		next:        next,
		todoService: service,
		lService:    lService,
		transact:    transact,
	}
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

	tx, err := t.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transaction in todo modify middleware, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer t.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

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

	collaboratorsPage, err := t.lService.GetCollaborators(ctx, todo.ListId, &filters.BaseFilters{})
	if err != nil {
		log.C(ctx).Errorf("failed to serve http, error %s when trying to get list with id %s collaborators", err.Error(), todo.ListId)

		utils.EncodeErrorWithCorrectStatusCode(w, err)
		return
	}

	assignee, err := t.todoService.GetTodoAssigneeToRecord(ctx, todoId)
	if err != nil {
		log.C(ctx).Errorf("faield to serve http, error %s when trying to get todo assingee", err.Error())

		utils.EncodeErrorWithCorrectStatusCode(w, err)
		return
	}

	if !determineWhetherUserHasAccess(assignee, listOwner, user, collaboratorsPage.Data) {
		utils.EncodeError(
			w,
			"forbidden: only administrators, collaborators, list owners, or the assignee may access todo",
			http.StatusForbidden,
		)
		return
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction in todo modify middleware, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.next.ServeHTTP(w, r)
}

func NewTodoModifyMiddlewareFunc(todoService tService, listService lService, transact persistence.Transactioner) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return newTodoModifyMiddleware(next, todoService, listService, transact)
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
