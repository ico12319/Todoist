package middlewares

import (
	"context"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/sql_query_decorators/filters"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/status_code_encoders"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/utils"
	log "github.com/I763039/Todo-List/internProject/todo_app_service/pkg/configuration"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/constants"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/models"
	"github.com/gorilla/mux"
	"net/http"
)

type TodoService interface {
	GetTodoRecord(ctx context.Context, todoId string) (*models.Todo, error)
	GetTodoAssigneeToRecord(ctx context.Context, todoId string) (*models.User, error)
}

type listService interface {
	GetListOwnerRecord(ctx context.Context, listId string) (*models.User, error)
	GetCollaborators(ctx context.Context, lFilters *filters.ListFilters) ([]*models.User, error)
}

type statusCodeEncoderFactory interface {
	CreateStatusCodeEncoder(ctx context.Context, w http.ResponseWriter, err error) status_code_encoders.StatusCodeEncoder
}

type todoModifyMiddleware struct {
	next        http.Handler
	todoService TodoService
	lService    listService
	factory     statusCodeEncoderFactory
}

func newTodoModifyMiddleware(next http.Handler, service TodoService, lService listService, factory statusCodeEncoderFactory) *todoModifyMiddleware {
	return &todoModifyMiddleware{next: next, todoService: service, lService: lService, factory: factory}
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

		encoder := t.factory.CreateStatusCodeEncoder(ctx, w, err)
		encoder.EncodeErrorWithCorrectStatusCode(ctx, w)
		return
	}

	listOwner, err := t.lService.GetListOwnerRecord(ctx, todo.ListId)
	if err != nil {
		log.C(ctx).Errorf("failed to serve http, error %s when trying to get list owner of list with id %s", err.Error(), todo.ListId)

		encoder := t.factory.CreateStatusCodeEncoder(ctx, w, err)
		encoder.EncodeErrorWithCorrectStatusCode(ctx, w)
		return
	}

	//TODO add filters
	collaborators, err := t.lService.GetCollaborators(ctx, &filters.ListFilters{
		ListId: todo.ListId,
	})

	if err != nil {
		log.C(ctx).Errorf("failed to serve http, error %s when trying to get list with id %s collaborators", err.Error(), todo.ListId)

		encoder := t.factory.CreateStatusCodeEncoder(ctx, w, err)
		encoder.EncodeErrorWithCorrectStatusCode(ctx, w)
		return
	}

	assignee, err := t.todoService.GetTodoAssigneeToRecord(r.Context(), todoId)
	if err != nil {
		log.C(ctx).Errorf("faield to serve http, error %s when trying to get todo assingee", err.Error())

		encoder := t.factory.CreateStatusCodeEncoder(ctx, w, err)
		encoder.EncodeErrorWithCorrectStatusCode(ctx, w)
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

func NewTodoModifyMiddlewareFunc(todoService TodoService, listService listService, factory statusCodeEncoderFactory) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return newTodoModifyMiddleware(next, todoService, listService, factory)
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
