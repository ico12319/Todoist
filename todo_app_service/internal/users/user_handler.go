package users

import (
	"Todo-List/internProject/todo_app_service/internal/middlewares"
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
	GetUsersRecords(ctx context.Context, uFilters *filters.UserFilters) ([]*models.User, error)
	GetUserRecord(ctx context.Context, userId string) (*models.User, error)
	DeleteUserRecord(ctx context.Context, id string) error
	DeleteUsers(ctx context.Context) error
	GetUserListsRecords(ctx context.Context, userId string, uFilter *filters.UserFilters) ([]*models.List, error)
	GetTodosAssignedToUser(ctx context.Context, userId string, userFilters *filters.UserFilters) ([]*models.Todo, error)
}

type fieldsValidator interface {
	Struct(st interface{}) error
}

type handler struct {
	service    userService
}

func NewHandler(service userService) *handler {
	return &handler{service: service}
}

func (h *handler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("getting user in user handler")

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

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(user); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *handler) HandleGetUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("getting users in user handler")

	limit := utils.GetLimitFromUrl(r)
	cursor := utils.GetContentFromUrl(r, constants.CURSOR)

	uFilters := &filters.UserFilters{
		BaseFilters: filters.BaseFilters{
			Limit:  limit,
			Cursor: cursor,
		},
	}

	users, err := h.service.GetUsersRecords(ctx, uFilters)
	if err != nil {
		log.C(ctx).Errorf("failed to get users from user handler due to %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(users); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *handler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("deleting user in user handler")

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
}

func (h *handler) HandleDeleteUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("deleting users in user handler")

	if err := h.service.DeleteUsers(ctx); err != nil {
		log.C(ctx).Errorf("failed to delete users, error %s when calling user service", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *handler) HandleGetUserLists(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("getting lists where user participates in, in user handler")

	userId, err := utils.GetValueFromContext[string](ctx, middlewares.UserId)
	if err != nil {
		log.C(ctx).Errorf("failed to get lists where user participates in user handler due to %s", err.Error())
		utils.EncodeError(w, constants.MISSING_USER_ID, http.StatusBadRequest)
		return
	}

	limit := utils.GetLimitFromUrl(r)
	cursor := utils.GetContentFromUrl(r, constants.CURSOR)
	role := utils.GetContentFromUrl(r, constants.ROLE)
	ownerId, participantId := determineRole(role, userId)

	uFilters := &filters.UserFilters{
		BaseFilters: filters.BaseFilters{
			Limit:  limit,
			Cursor: cursor,
		},
		OwnerId:       ownerId,
		ParticipantId: participantId,
	}

	lists, err := h.service.GetUserListsRecords(ctx, userId, uFilters)
	if err != nil {
		log.C(ctx).Errorf("failed to get lists where user participates in, error %s when calling user service", err.Error())

		utils.EncodeErrorWithCorrectStatusCode(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(lists); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *handler) HandleGetTodosAssignedToUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("getting todos assigned to user in user handler")

	userId, err := utils.GetValueFromContext[string](ctx, middlewares.UserId)
	if err != nil {
		log.C(ctx).Error("failed to get todos assigned to user, error when trying to get user_id from context")
		utils.EncodeError(w, constants.MISSING_USER_ID, http.StatusBadRequest)
		return
	}

	limit := utils.GetLimitFromUrl(r)
	cursor := utils.GetContentFromUrl(r, constants.CURSOR)
	userFilters := &filters.UserFilters{
		BaseFilters: filters.BaseFilters{
			Limit:  limit,
			Cursor: cursor,
		},
	}

	modelTodos, err := h.service.GetTodosAssignedToUser(ctx, userId, userFilters)
	if err != nil {
		log.C(ctx).Errorf("failed to get todos assigned to user, error %s when calling todo service", err.Error())

		utils.EncodeErrorWithCorrectStatusCode(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(modelTodos); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func determineRole(role string, userId string) (string, string) {
	var ownerId string
	var participantId string

	if len(role) == 0 {
		ownerId = userId
		participantId = userId
	} else if role == constants.OWNER_ROLE {
		ownerId = userId
	} else if role == constants.PARTICIPANT_ROLE {
		participantId = userId
	}

	return ownerId, participantId
}
