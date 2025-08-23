package todos

import (
	"Todo-List/internProject/todo_app_service/internal/application_errors"
	middlewares2 "Todo-List/internProject/todo_app_service/internal/middlewares"
	"Todo-List/internProject/todo_app_service/internal/persistence"
	"Todo-List/internProject/todo_app_service/internal/resource_identifier"
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
	GetTodoRecords(ctx context.Context, f filters.SqlFilters, rf resource_identifier.ResourceIdentifier) (*models.TodoPage, error)
	GetTodosByListId(ctx context.Context, f filters.SqlFilters, listId string, rf resource_identifier.ResourceIdentifier) (*models.TodoPage, error)
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
type Handler struct {
	serv       todoService
	fValidator fieldsValidator
	transact   persistence.Transactioner
}

func NewHandler(serv todoService, fValidator fieldsValidator, transact persistence.Transactioner) *Handler {
	return &Handler{
		serv:       serv,
		fValidator: fValidator,
		transact:   transact,
	}
}

func (h *Handler) HandleGetTodoAssignee(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("getting todo assignee in todo handler")

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transaction todo handler, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

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

	if err = json.NewEncoder(w).Encode(assignee); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction when trying to get todo assignee, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleTodoCreation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transaction todo handler, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	var todo handler_models.CreateTodo
	if err = json.NewDecoder(r.Body).Decode(&todo); err != nil {
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

	if err = json.NewEncoder(w).Encode(modelTodo); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction when trying to create todo, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) HandleGetTodos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("getting todo in todo handler")

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transaction todo handler, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	status := utils.GetContentFromUrl(r, constants.STATUS)
	priority := utils.GetContentFromUrl(r, constants.PRIORITY)
	name := utils.GetContentFromUrl(r, constants.NAME)
	first := utils.GetContentFromUrl(r, constants.FIRST)
	after := utils.GetContentFromUrl(r, constants.AFTER)
	before := utils.GetContentFromUrl(r, constants.BEFORE)
	last := utils.GetContentFromUrl(r, constants.LAST)

	if len(first) == 0 && len(last) == 0 {
		first = constants.DEFAULT_LIMIT_VALUE
	}

	overdue := utils.GetContentFromUrl(r, constants.OVERDUE)

	tFilter := &filters.TodoFilters{
		PaginationFilters: filters.PaginationFilters{
			First:  first,
			Last:   last,
			After:  after,
			Before: before,
		},
		Status:   status,
		Priority: priority,
		Overdue:  overdue,
		Name:     name,
	}

	resourceIdentifier := &resource_identifier.GenericResourceIdentifier{}
	resourceIdentifier.SetResourceIdentifier(constants.TodosIdentifier)

	todos, err := h.serv.GetTodoRecords(ctx, tFilter, resourceIdentifier)
	if err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(todos); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction when trying to get todos, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleDeleteTodo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("deleting todo in todo handler")

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transaction todo handler, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

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

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction when trying to delete todo with id %s, error %s", todoId, err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) HandleUpdateTodoRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Debug("updating todo in todo handler")

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transaction todo handler, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

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

	if err = json.NewEncoder(w).Encode(&updatedModel); err != nil {
		log.C(ctx).Errorf("failed to update todo record, error when trying to encode json reponse %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction when trying to update todo with id %s, error %s", todoId, err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleGetTodo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("getting todo in todo handler")

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transaction todo handler, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

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

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction when trying to get todo with id %s, error %s", todoId, err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleDeleteTodos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Debug("deleting todos in todo handler")

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transaction todo handler, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err = h.serv.DeleteTodosRecords(ctx); err != nil {
		log.C(ctx).Errorf("failed to delete todos in todo handler due to an erroo in todo service")
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction when trying to delete todos, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) HandleGetTodosByListId(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Debug("getting todos by list_id in todo handler")

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transaction todo handler, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	listId, err := utils.GetValueFromContext[string](ctx, middlewares2.ListId)
	if err != nil {
		log.C(ctx).Errorf("failed to get todos by list_id, missing list_id in context")
		utils.EncodeError(w, constants.CONTEXT_NOT_CONTAINING_VALID_LIST_ID, http.StatusBadRequest)
		return
	}

	first := utils.GetContentFromUrl(r, constants.FIRST)
	after := utils.GetContentFromUrl(r, constants.AFTER)
	before := utils.GetContentFromUrl(r, constants.BEFORE)
	last := utils.GetContentFromUrl(r, constants.LAST)

	if len(first) == 0 && len(last) == 0 {
		first = constants.DEFAULT_LIMIT_VALUE
	} else if len(first) != 0 && len(last) != 0 {
		log.C(ctx).Warn("both first and last passed as query params...")
		utils.EncodeError(w, "can't pass both first and last values as query params", http.StatusBadRequest)
		return
	}

	status := utils.GetContentFromUrl(r, constants.STATUS)
	priority := utils.GetContentFromUrl(r, constants.PRIORITY)
	name := utils.GetContentFromUrl(r, constants.NAME)
	overdue := utils.GetContentFromUrl(r, constants.OVERDUE)

	f := &filters.TodoFilters{
		PaginationFilters: filters.PaginationFilters{
			First:  first,
			After:  after,
			Before: before,
			Last:   last,
		},
		Status:   status,
		Priority: priority,
		ListID:   listId,
		Name:     name,
		Overdue:  overdue,
	}

	resourceIdentifier := &resource_identifier.GenericResourceIdentifier{}
	resourceIdentifier.SetResourceIdentifier(constants.TodosIdentifier)

	todos, err := h.serv.GetTodosByListId(ctx, f, listId, resourceIdentifier)
	if err != nil {
		log.C(ctx).Errorf("failed to get todos by list_id, error in todo service")
		utils.EncodeErrorWithCorrectStatusCode(w, err)
		return
	}

	if err = json.NewEncoder(w).Encode(todos); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction when trying to get todos by list id %s, error %s", listId, err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleDeleteTodosByListId(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Debug("deleting todos by list_id in todo handler")

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transaction todo handler, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

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

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction when trying to delete todos by list id %s, error %s", listId, err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) HandleGetTodoByListId(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("getting todo from list in todo handler")

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transaction todo handler, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

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

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction when trying to get todo by list id %s, error %s", listId, err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
