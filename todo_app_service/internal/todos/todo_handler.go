package todos

import (
	"Todo-List/internProject/todo_app_service/internal/application_errors"
	middlewares2 "Todo-List/internProject/todo_app_service/internal/middlewares"
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators/filters"
	"Todo-List/internProject/todo_app_service/internal/utils"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"Todo-List/internProject/todo_app_service/pkg/handler_models"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"context"
	"encoding/json"
	"net/http"
)

//go:generate mockery --name=IService --output=./mocks --outpkg=mocks --filename=Iservice.go --with-expecter=true
type todoService interface {
	CreateTodoRecord(ctx context.Context, todo *handler_models.CreateTodo, creator *models.User) (*models.Todo, error)
	GetTodoRecords(ctx context.Context, filters *filters.TodoFilters) (*models.TodoPage, error)
	GetTodosByListId(ctx context.Context, filters *filters.TodoFilters, listId string) (*models.TodoPage, error)
	GetTodoByListId(ctx context.Context, listId string, todoId string) (*models.Todo, error)
	GetTodoRecord(ctx context.Context, todoId string) (*models.Todo, error)
	GetTodoAssigneeToRecord(ctx context.Context, todoId string) (*models.User, error)
	DeleteTodoRecord(ctx context.Context, todoId string) error
	DeleteTodosRecords(ctx context.Context) error
	DeleteTodosRecordsByListId(ctx context.Context, listId string) error
	UpdateTodoRecord(ctx context.Context, todoId string, todo *handler_models.UpdateTodo) (*models.Todo, error)
}

type fieldsValidator interface {
	Struct(st interface{}) error
}
type handler struct {
	serv       todoService
	fValidator fieldsValidator
}

func NewHandler(serv todoService, fValidator fieldsValidator) *handler {
	return &handler{serv: serv, fValidator: fValidator}
}

func (h *handler) HandleGetTodoAssignee(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("getting todo assignee in todo handler")

	todoId, err := utils.GetValueFromContext[string](ctx, middlewares2.TodoId)
	if err != nil {
		log.C(ctx).Errorf("failed to get todo assignee, missing todo_id in context")
		utils.EncodeError(w, constants.CONTEXT_NOT_CONTAINING_VALID_TODO_ID, http.StatusBadRequest)
		return
	}

	assignee, err := h.serv.GetTodoAssigneeToRecord(ctx, todoId)
	if err != nil {
		log.C(ctx).Errorf("faile to get todo assignee, error %s when callin todo service", err.Error())

		utils.EncodeErrorWithCorrectStatusCode(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(assignee); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *handler) HandleTodoCreation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var todo handler_models.CreateTodo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		utils.EncodeError(w, constants.INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	field, err := utils.CheckForValidationError(h.fValidator, todo)
	if err != nil {
		log.C(ctx).Error("failed to create todo, error because one of the required fields is missing")
		utils.EncodeError(w, application_errors.NewEmptyFieldError(field).Error(), http.StatusBadRequest)
		return
	}

	creator, err := utils.GetValueFromContext[*models.User](ctx, middlewares2.UserKey)
	if err != nil {
		log.C(ctx).Error("failed to create todo, missing creator in context...")
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	modelTodo, err := h.serv.CreateTodoRecord(ctx, &todo, creator)
	if err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(modelTodo); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *handler) HandleGetTodos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("getting todo in todo handler")

	status := utils.GetContentFromUrl(r, constants.STATUS)
	priority := utils.GetContentFromUrl(r, constants.PRIORITY)
	limit := utils.GetLimitFromUrl(r)
	cursor := utils.GetContentFromUrl(r, constants.CURSOR)
	overdue := utils.GetContentFromUrl(r, constants.OVERDUE)

	tFilter := &filters.TodoFilters{
		BaseFilters: filters.BaseFilters{
			Limit:  limit,
			Cursor: cursor,
		},
		Status:   status,
		Priority: priority,
		Overdue:  overdue,
	}

	todos, err := h.serv.GetTodoRecords(ctx, tFilter)
	if err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(todos); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *handler) HandleDeleteTodo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("deleting todo in todo handler")

	todoId, err := utils.GetValueFromContext[string](ctx, middlewares2.TodoId)
	if err != nil {
		log.C(ctx).Errorf("failed to delete todo, missing todo_id in context")
		utils.EncodeError(w, constants.CONTEXT_NOT_CONTAINING_VALID_TODO_ID, http.StatusBadRequest)
		return
	}

	if err = h.serv.DeleteTodoRecord(ctx, todoId); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) HandleUpdateTodoRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Debug("updating todo in todo handler")

	todoId, err := utils.GetValueFromContext[string](ctx, middlewares2.TodoId)
	if err != nil {
		log.C(ctx).Errorf("failed to update todo, missing todo_id in context")
		utils.EncodeError(w, constants.CONTEXT_NOT_CONTAINING_VALID_TODO_ID, http.StatusBadRequest)
		return
	}

	var todo handler_models.UpdateTodo
	if err = json.NewDecoder(r.Body).Decode(&todo); err != nil {
		log.C(ctx).Errorf("failed to decode request body in todo handler")
		utils.EncodeError(w, constants.INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	field, err := utils.CheckForValidationError(h.fValidator, todo)
	if err != nil {
		log.C(ctx).Errorf("failed to update todo, error %s because one of the required fields is missing", err.Error())
		utils.EncodeError(w, application_errors.NewEmptyFieldError(field).Error(), http.StatusBadRequest)
		return
	}

	updatedModel, err := h.serv.UpdateTodoRecord(ctx, todoId, &todo)
	if err != nil {
		log.C(ctx).Errorf("failed to update todo in todo handler, error %s when calling todo service", err.Error())

		utils.EncodeErrorWithCorrectStatusCode(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(&updatedModel); err != nil {
		log.C(ctx).Errorf("failed to update todo record, error when trying to encode json reponse %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *handler) HandleGetTodo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("getting todo in todo handler")

	todoId, err := utils.GetValueFromContext[string](ctx, middlewares2.TodoId)
	if err != nil {
		log.C(ctx).Errorf("failed to get todo, missing todo_id in context")
		utils.EncodeError(w, constants.CONTEXT_NOT_CONTAINING_VALID_TODO_ID, http.StatusBadRequest)
		return
	}

	todo, err := h.serv.GetTodoRecord(ctx, todoId)
	if err != nil {
		log.C(ctx).Errorf("failed to get todo in todo handler due to an error in todo service")

		utils.EncodeErrorWithCorrectStatusCode(w, err)
		return
	}

	if err = json.NewEncoder(w).Encode(todo); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *handler) HandleDeleteTodos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Debug("deleting todos in todo handler")

	err := h.serv.DeleteTodosRecords(ctx)

	if err != nil {
		log.C(ctx).Errorf("failed to delete todos in todo handler due to an erroo in todo service")
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) HandleGetTodosByListId(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Debug("getting todos by list_id in todo handler")

	listId, err := utils.GetValueFromContext[string](ctx, middlewares2.ListId)
	if err != nil {
		log.C(ctx).Errorf("failed to get todos by list_id, missing list_id in context")
		utils.EncodeError(w, constants.CONTEXT_NOT_CONTAINING_VALID_LIST_ID, http.StatusBadRequest)
		return
	}

	limit := utils.GetLimitFromUrl(r)
	cursor := utils.GetContentFromUrl(r, constants.CURSOR)
	status := utils.GetContentFromUrl(r, constants.STATUS)
	priority := utils.GetContentFromUrl(r, constants.PRIORITY)

	f := &filters.TodoFilters{
		BaseFilters: filters.BaseFilters{
			Limit:  limit,
			Cursor: cursor,
		},
		Status:   status,
		Priority: priority,
	}

	todos, err := h.serv.GetTodosByListId(ctx, f, listId)
	if err != nil {
		log.C(ctx).Errorf("failed to get todos by list_id, error in todo service")

		utils.EncodeErrorWithCorrectStatusCode(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(todos); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *handler) HandleDeleteTodosByListId(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Debug("deleting todos by list_id in todo handler")

	listId, err := utils.GetValueFromContext[string](ctx, middlewares2.ListId)
	if err != nil {
		log.C(ctx).Errorf("failed to get todos by list_id, missing list_id in context")
		utils.EncodeError(w, constants.CONTEXT_NOT_CONTAINING_VALID_LIST_ID, http.StatusBadRequest)
		return
	}

	err = h.serv.DeleteTodosRecordsByListId(ctx, listId)
	if err != nil {
		log.C(ctx).Errorf("failed to delete todos by list_id, error in todo service")

		utils.EncodeErrorWithCorrectStatusCode(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) HandleGetTodoByListId(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("getting todo from list in todo handler")

	listId, err := utils.GetValueFromContext[string](ctx, middlewares2.ListId)
	if err != nil {
		log.C(ctx).Errorf("failed to get todo by list_id, missing list_id in context")
		utils.EncodeError(w, constants.CONTEXT_NOT_CONTAINING_VALID_LIST_ID, http.StatusBadRequest)
		return
	}

	todoId, err := utils.GetValueFromContext[string](ctx, middlewares2.TodoId)
	if err != nil {
		log.C(ctx).Errorf("failed to get todo by list_id, missing todo_id in context")
		utils.EncodeError(w, constants.CONTEXT_NOT_CONTAINING_VALID_TODO_ID, http.StatusBadRequest)
		return
	}

	todo, err := h.serv.GetTodoByListId(ctx, listId, todoId)
	if err != nil {
		log.C(ctx).Errorf("failed to get todo with id %s from list with id %s, error %s", todoId, listId, err.Error())

		utils.EncodeErrorWithCorrectStatusCode(w, err)
		return
	}

	if err = json.NewEncoder(w).Encode(todo); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *handler) HandleDeleteTodoByListId(w http.ResponseWriter, r *http.Request) {}
