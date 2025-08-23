package users

import (
	"Todo-List/internProject/todo_app_service/internal/middlewares"
	"Todo-List/internProject/todo_app_service/internal/persistence"
	"Todo-List/internProject/todo_app_service/internal/resource_identifier"
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators/filters"
	"Todo-List/internProject/todo_app_service/internal/utils"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"context"
	"encoding/json"
	"net/http"
)

type userService interface {
	GetUsersRecords(ctx context.Context, uFilters filters.SqlFilters, rf resource_identifier.ResourceIdentifier) (*models.UserPage, error)
	GetUserRecord(ctx context.Context, userId string) (*models.User, error)
	DeleteUserRecord(ctx context.Context, id string) error
	DeleteUsers(ctx context.Context) error
	GetUserListsRecords(ctx context.Context, userId string, uFilter filters.SqlFilters, rf resource_identifier.ResourceIdentifier) (*models.ListPage, error)
	GetTodosAssignedToUser(ctx context.Context, userId string, userFilters filters.SqlFilters, rf resource_identifier.ResourceIdentifier) (*models.TodoPage, error)
}

type Handler struct {
	service  userService
	transact persistence.Transactioner
}

func NewHandler(service userService, transact persistence.Transactioner) *Handler {
	return &Handler{
		service:  service,
		transact: transact,
	}
}

func (h *Handler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("getting user in user handler")

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transactin in user handler, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	userId, err := utils.GetValueFromContext[string](ctx, middlewares.UserId)
	if err != nil {
		log.C(ctx).Errorf("failed to get user from user handler due to %s", err.Error())
		utils.EncodeError(w, constants.MISSING_USER_ID, http.StatusBadRequest)
		return
	}

	user, err := h.service.GetUserRecord(ctx, userId)
	if err != nil {
		log.C(ctx).Errorf("failed to get user from user handler due to %s", err.Error())

		utils.EncodeErrorWithCorrectStatusCode(w, err)
		return
	}

	if err = json.NewEncoder(w).Encode(user); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction when trying to get user with id %s, error %s", userId, err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleGetUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("getting users in user handler")

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transactin in user handler, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	first := utils.GetContentFromUrl(r, constants.FIRST)
	after := utils.GetContentFromUrl(r, constants.AFTER)
	last := utils.GetContentFromUrl(r, constants.LAST)
	before := utils.GetContentFromUrl(r, constants.BEFORE)

	if len(first) == 0 && len(last) == 0 {
		first = constants.DEFAULT_LIMIT_VALUE
	}

	uFilters := &filters.UserFilters{
		PaginationFilters: filters.PaginationFilters{
			First:  first,
			After:  after,
			Last:   last,
			Before: before,
		},
	}

	resourceIdentifier := &resource_identifier.GenericResourceIdentifier{}
	resourceIdentifier.SetResourceIdentifier(constants.UsersIdentifier)

	users, err := h.service.GetUsersRecords(ctx, uFilters, resourceIdentifier)
	if err != nil {
		log.C(ctx).Errorf("failed to get users from user handler due to %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(users); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction when trying to get user records, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("deleting user in user handler")

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transtaction when trying to delete user, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	userId, err := utils.GetValueFromContext[string](ctx, middlewares.UserId)
	if err != nil {
		log.C(ctx).Errorf("failed to delete user in user handler due to %s", err.Error())
		utils.EncodeError(w, constants.MISSING_USER_ID, http.StatusBadRequest)
		return
	}

	if err = h.service.DeleteUserRecord(ctx, userId); err != nil {
		log.C(ctx).Errorf("failed to delete user in user handler, error %s when calling user service", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction in user delete handler, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleDeleteUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("deleting users in user handler")

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transtaction when trying to delete users, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err = h.service.DeleteUsers(ctx); err != nil {
		log.C(ctx).Errorf("failed to delete users, error %s when calling user service", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction in users delete handler, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleGetUserLists(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("getting lists where user participates in, in user handler")

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transtaction when trying to get user lists, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	userId, err := utils.GetValueFromContext[string](ctx, middlewares.UserId)
	if err != nil {
		log.C(ctx).Errorf("failed to get lists where user participates in user handler due to %s", err.Error())
		utils.EncodeError(w, constants.MISSING_USER_ID, http.StatusBadRequest)
		return
	}

	first := utils.GetContentFromUrl(r, constants.FIRST)
	after := utils.GetContentFromUrl(r, constants.AFTER)
	before := utils.GetContentFromUrl(r, constants.BEFORE)
	last := utils.GetContentFromUrl(r, constants.LAST)

	name := utils.GetContentFromUrl(r, constants.NAME)

	if len(first) == 0 && len(last) == 0 {
		first = constants.DEFAULT_LIMIT_VALUE
	}

	lFilters := &filters.ListFilters{
		PaginationFilters: filters.PaginationFilters{
			First:  first,
			After:  after,
			Before: before,
			Last:   last,
		},
		OwnerID: userId,
		Name:    name,
	}

	resourceIdentifier := &resource_identifier.GenericResourceIdentifier{}
	resourceIdentifier.SetResourceIdentifier(constants.UsersListsIdentifier)

	lists, err := h.service.GetUserListsRecords(ctx, userId, lFilters, resourceIdentifier)
	if err != nil {
		log.C(ctx).Errorf("failed to get lists where user participates in, error %s when calling user service", err.Error())

		utils.EncodeErrorWithCorrectStatusCode(w, err)
		return
	}

	if err = json.NewEncoder(w).Encode(lists); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction in user lists handler, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleGetTodosAssignedToUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("getting todos assigned to user in user handler")

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transtaction when trying to get user todos assigned to user, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	userId, err := utils.GetValueFromContext[string](ctx, middlewares.UserId)
	if err != nil {
		log.C(ctx).Error("failed to get todos assigned to user, error when trying to get user_id from context")
		utils.EncodeError(w, constants.MISSING_USER_ID, http.StatusBadRequest)
		return
	}

	first := utils.GetContentFromUrl(r, constants.FIRST)
	after := utils.GetContentFromUrl(r, constants.AFTER)
	before := utils.GetContentFromUrl(r, constants.BEFORE)
	last := utils.GetContentFromUrl(r, constants.LAST)

	status := utils.GetContentFromUrl(r, constants.STATUS)
	priority := utils.GetContentFromUrl(r, constants.PRIORITY)
	overdue := utils.GetContentFromUrl(r, constants.OVERDUE)
	name := utils.GetContentFromUrl(r, constants.NAME)

	todoFilters := &filters.TodoFilters{
		PaginationFilters: filters.PaginationFilters{
			First:  first,
			After:  after,
			Before: before,
			Last:   last,
		},
		Status:   status,
		Priority: priority,
		Overdue:  overdue,
		Name:     name,
		UserID:   userId,
	}

	resourceIdentifier := &resource_identifier.GenericResourceIdentifier{}
	resourceIdentifier.SetResourceIdentifier(constants.UsersTodosIdentifier)

	modelTodos, err := h.service.GetTodosAssignedToUser(ctx, userId, todoFilters, resourceIdentifier)
	if err != nil {
		log.C(ctx).Errorf("failed to get todos assigned to user, error %s when calling todo service", err.Error())

		utils.EncodeErrorWithCorrectStatusCode(w, err)
		return
	}

	if err = json.NewEncoder(w).Encode(modelTodos); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction in user lists handler, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
