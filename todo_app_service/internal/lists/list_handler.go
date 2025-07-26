package lists

import (
	"Todo-List/internProject/todo_app_service/internal/application_errors"
	"Todo-List/internProject/todo_app_service/internal/middlewares"
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
type listService interface {
	GetListRecord(context.Context, string) (*models.List, error)
	GetListsRecords(context.Context, *filters.BaseFilters) ([]*models.List, error)
	GetCollaborators(context.Context, string, *filters.BaseFilters) ([]*models.User, error)
	GetListOwnerRecord(context.Context, string) (*models.User, error)
	DeleteListRecord(context.Context, string) error
	DeleteLists(context.Context) error
	CreateListRecord(context.Context, *handler_models.CreateList, string) (*models.List, error)
	UpdateListPartiallyRecord(context.Context, string, *handler_models.UpdateList) (*models.List, error)
	AddCollaborator(context.Context, string, string) (*models.User, error)
	DeleteCollaborator(context.Context, string, string) error
}

type fieldValidator interface {
	Struct(interface{}) error
}

type handler struct {
	serv       listService
	fValidator fieldValidator
}

func NewHandler(service listService, fValidator fieldValidator) *handler {
	return &handler{serv: service, fValidator: fValidator}
}

func (h *handler) HandleGetLists(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("getting lists from list handler")

	limit := utils.GetLimitFromUrl(r)
	cursor := utils.GetContentFromUrl(r, constants.CURSOR)

	lFilter := &filters.BaseFilters{
		Limit:  limit,
		Cursor: cursor,
	}

	lists, err := h.serv.GetListsRecords(ctx, lFilter)
	if err != nil {
		log.C(ctx).Errorf("failed to get lists in list handler due to an error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(lists); err != nil {
		log.C(ctx).Errorf("failed to get lists in list handler due to an error %s when trying to encode JSON", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *handler) HandleGetCollaborators(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("getting list's collaborators in list handler")

	listId, err := utils.GetValueFromContext[string](r.Context(), middlewares.ListId)
	if err != nil {
		log.C(ctx).Errorf("failed to get list_id from the context in list handler %s", err.Error())
		utils.EncodeError(w, constants.CONTEXT_NOT_CONTAINING_VALID_LIST_ID, http.StatusBadRequest)
		return
	}

	limit := utils.GetLimitFromUrl(r)
	cursor := utils.GetContentFromUrl(r, constants.CURSOR)

	lFilters := &filters.BaseFilters{
		Limit:  limit,
		Cursor: cursor,
	}

	collaborators, err := h.serv.GetCollaborators(ctx, listId, lFilters)
	if err != nil {
		log.C(ctx).Errorf("failed to get list's collaborators in list handler due to an error %s in list service", err.Error())

		utils.EncodeErrorWithCorrectStatusCode(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(collaborators); err != nil {
		log.C(ctx).Error("failed to get list's collaborators due to an error when trying to encode JSON in list handler")
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *handler) HandleDeleteList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("deleting list in list handler")

	listId, err := utils.GetValueFromContext[string](r.Context(), middlewares.ListId)
	if err != nil {
		log.C(ctx).Error("failed to get list_id from the context in list handler")
		utils.EncodeError(w, constants.CONTEXT_NOT_CONTAINING_VALID_LIST_ID, http.StatusBadRequest)
		return
	}

	if err = h.serv.DeleteListRecord(ctx, listId); err != nil {
		log.C(ctx).Errorf("failed to get list's collaborators in list handler due to an error %s in list service", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) HandleDeleteLists(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("deleting lists in list handler")

	if err := h.serv.DeleteLists(ctx); err != nil {
		log.C(ctx).Errorf("failed to delete lists, error %s when calling list service", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) HandleCreateList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("creating list in list handler")

	var list handler_models.CreateList
	if err := json.NewDecoder(r.Body).Decode(&list); err != nil {
		log.C(ctx).Errorf("failed to decode list handler model %s", err.Error())
		utils.EncodeError(w, constants.INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	field, err := utils.CheckForValidationError(h.fValidator, list)
	if err != nil {
		log.C(ctx).Errorf("failed to create list, error because one of the required fields is missing %s", field)
		utils.EncodeError(w, application_errors.NewEmptyFieldError(field).Error(), http.StatusBadRequest)
		return

	}

	owner, err := utils.GetValueFromContext[*models.User](r.Context(), middlewares.UserKey)
	if err != nil {
		log.C(ctx).Errorf("failed to get list's owner due to an error %s when trying to get value from context in list handler", err.Error())
		utils.EncodeError(w, constants.CONTEXT_NOT_CONTAINING_VALID_USER, http.StatusInternalServerError)
		return
	}

	l, err := h.serv.CreateListRecord(ctx, &list, owner.Id)
	if err != nil {
		log.C(ctx).Errorf("failed to create list in list handler due to an error %s in list service", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(l); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func (h *handler) HandleUpdateListPartially(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Debug("changing list name in list handler")

	listId, err := utils.GetValueFromContext[string](r.Context(), middlewares.ListId)
	if err != nil {
		log.C(ctx).Errorf("failed to get list_id from the context in list handler")
		utils.EncodeError(w, constants.CONTEXT_NOT_CONTAINING_VALID_LIST_ID, http.StatusBadRequest)
		return
	}

	var updateModel handler_models.UpdateList
	if err = json.NewDecoder(r.Body).Decode(&updateModel); err != nil {
		log.C(ctx).Errorf("failed to decode list handler model")
		utils.EncodeError(w, constants.INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	field, err := utils.CheckForValidationError(h.fValidator, updateModel)
	if err != nil {
		log.C(ctx).Errorf("failed to update list, error because one of the required fields is missing %s", field)
		utils.EncodeError(w, application_errors.NewEmptyFieldError(field).Error(), http.StatusBadRequest)
		return

	}

	updatedModel, err := h.serv.UpdateListPartiallyRecord(ctx, listId, &updateModel)
	if err != nil {
		log.C(ctx).Errorf("failed to change list name, error when calling list service method %s", err.Error())

		utils.EncodeErrorWithCorrectStatusCode(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(&updatedModel); err != nil {
		log.C(ctx).Errorf("failed to encode JSON %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (h *handler) HandleAddCollaborator(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Debug("adding a collaborator in list handler")

	listId, err := utils.GetValueFromContext[string](r.Context(), middlewares.ListId)
	if err != nil {
		log.C(ctx).Errorf("failed to get list_id from the context in list handler")
		utils.EncodeError(w, constants.CONTEXT_NOT_CONTAINING_VALID_LIST_ID, http.StatusBadRequest)
		return
	}

	var user handler_models.AddCollaborator
	if err = json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.C(ctx).Error("failed to add a collaborator in list handler due to an error when trying to decode user handler model")
		utils.EncodeError(w, constants.INVALID_REQUEST_BODY, http.StatusBadRequest)
		return
	}

	modelUser, err := h.serv.AddCollaborator(ctx, listId, user.Id)
	if err != nil {
		log.C(ctx).Error("failed to add collaborator in list handler, error when calling service function")

		utils.EncodeErrorWithCorrectStatusCode(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(modelUser); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *handler) HandleDeleteCollaborator(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("deleting collaborator in list handler")

	listId, err := utils.GetValueFromContext[string](ctx, middlewares.ListId)
	if err != nil {
		log.C(ctx).Errorf("failed to get list_id from the context in list handler")
		utils.EncodeError(w, constants.CONTEXT_NOT_CONTAINING_VALID_LIST_ID, http.StatusBadRequest)
		return
	}

	userId, err := utils.GetValueFromContext[string](ctx, middlewares.UserId)
	if err != nil {
		log.C(ctx).Errorf("failed to get user_id from the context in list handler")
		utils.EncodeError(w, constants.CONTEXT_NOT_CONTAINING_VALID_USER_ID, http.StatusBadRequest)
		return
	}

	if err = h.serv.DeleteCollaborator(ctx, listId, userId); err != nil {
		log.C(ctx).Errorf("failed to delete collaborator from list in list handler")

		utils.EncodeErrorWithCorrectStatusCode(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) HandleGetListRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("getting list record in list handler")

	listId, err := utils.GetValueFromContext[string](r.Context(), middlewares.ListId)
	if err != nil {
		log.C(ctx).Errorf("failed to get list_id from the context in list handler")
		utils.EncodeError(w, constants.CONTEXT_NOT_CONTAINING_VALID_LIST_ID, http.StatusBadRequest)
		return
	}

	list, err := h.serv.GetListRecord(ctx, listId)
	if err != nil {
		log.C(ctx).Errorf("failed to get list record in list handler due to an error in list service")

		utils.EncodeErrorWithCorrectStatusCode(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(list); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *handler) HandleGetListOwner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("getting list owner in list handler")

	listId, err := utils.GetValueFromContext[string](r.Context(), middlewares.ListId)
	if err != nil {
		log.C(ctx).Errorf("failed to get list_id from the context in list handler")
		utils.EncodeError(w, constants.CONTEXT_NOT_CONTAINING_VALID_LIST_ID, http.StatusBadRequest)
		return
	}

	listOwner, err := h.serv.GetListOwnerRecord(ctx, listId)
	if err != nil {
		log.C(ctx).Errorf("failed to get list owner in list handler due to an error in list service")

		utils.EncodeErrorWithCorrectStatusCode(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(listOwner); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
