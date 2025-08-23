package lists

import (
	"Todo-List/internProject/todo_app_service/internal/application_errors"
	"Todo-List/internProject/todo_app_service/internal/middlewares"
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
type listService interface {
	GetListRecord(ctx context.Context, listId string) (*models.List, error)
	GetListsRecords(ctx context.Context, filters filters.SqlFilters, rf resource_identifier.ResourceIdentifier) (*models.ListPage, error)
	GetCollaborators(ctx context.Context, listId string, filters filters.SqlFilters, rf resource_identifier.ResourceIdentifier) (*models.UserPage, error)
	GetListOwnerRecord(ctx context.Context, listId string) (*models.User, error)
	DeleteListRecord(ctx context.Context, listId string) error
	DeleteLists(ctx context.Context) error
	CreateListRecord(ctx context.Context, list *handler_models.CreateList, ownerId string) (*models.List, error)
	UpdateListPartiallyRecord(ctx context.Context, listId string, list *handler_models.UpdateList) (*models.List, error)
	AddCollaborator(ctx context.Context, listId string, userEmail string) (*models.User, error)
	DeleteCollaborator(ctx context.Context, listId string, userId string) error
}

type fieldValidator interface {
	Struct(interface{}) error
}

type Handler struct {
	serv       listService
	fValidator fieldValidator
	transact   persistence.Transactioner
}

func NewHandler(service listService, fValidator fieldValidator, transact persistence.Transactioner) *Handler {
	return &Handler{
		serv:       service,
		fValidator: fValidator,
		transact:   transact,
	}
}

func (h *Handler) HandleGetLists(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("getting lists from list handler")

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transaction in list handler, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	first := utils.GetContentFromUrl(r, constants.FIRST)
	after := utils.GetContentFromUrl(r, constants.AFTER)
	before := utils.GetContentFromUrl(r, constants.BEFORE)
	last := utils.GetContentFromUrl(r, constants.LAST)

	name := utils.GetContentFromUrl(r, constants.NAME)

	if len(first) == 0 && len(last) == 0 {
		first = constants.DEFAULT_LIMIT_VALUE
	}

	lFilter := &filters.ListFilters{
		PaginationFilters: filters.PaginationFilters{
			First:  first,
			Last:   last,
			After:  after,
			Before: before,
		},
		Name: name,
	}

	rf := &resource_identifier.GenericResourceIdentifier{}
	rf.SetResourceIdentifier(constants.ListIdentifier)

	lists, err := h.serv.GetListsRecords(ctx, lFilter, rf)
	if err != nil {
		log.C(ctx).Errorf("failed to get lists in list handler due to an error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(lists); err != nil {
		log.C(ctx).Errorf("failed to get lists in list handler due to an error %s when trying to encode JSON", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit transaction when trying to get lists in list handler, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleGetCollaborators(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("getting list's collaborators in list handler")

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transaction in list handler, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	listId, err := utils.GetValueFromContext[string](r.Context(), middlewares.ListId)
	if err != nil {
		log.C(ctx).Errorf("failed to get list_id from the context in list handler %s", err.Error())
		utils.EncodeError(w, constants.CONTEXT_NOT_CONTAINING_VALID_LIST_ID, http.StatusBadRequest)
		return
	}

	first := utils.GetContentFromUrl(r, constants.FIRST)
	after := utils.GetContentFromUrl(r, constants.AFTER)
	before := utils.GetContentFromUrl(r, constants.BEFORE)
	last := utils.GetContentFromUrl(r, constants.LAST)

	if len(first) == 0 && len(last) == 0 {
		first = constants.DEFAULT_LIMIT_VALUE
	}

	lFilters := &filters.UserFilters{
		PaginationFilters: filters.PaginationFilters{
			First:  first,
			Last:   last,
			Before: before,
			After:  after,
		},
		ListID: listId,
	}

	rf := &resource_identifier.GenericResourceIdentifier{}
	rf.SetResourceIdentifier(constants.ListsUsersIdentifier)

	collaborators, err := h.serv.GetCollaborators(ctx, listId, lFilters, rf)
	if err != nil {
		log.C(ctx).Errorf("failed to get list's collaborators in list handler due to an error %s in list service", err.Error())

		utils.EncodeErrorWithCorrectStatusCode(w, err)
		return
	}

	if err = json.NewEncoder(w).Encode(collaborators); err != nil {
		log.C(ctx).Error("failed to get list's collaborators due to an error when trying to encode JSON in list handler")
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit trasnaction when trying to get collaborators, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleDeleteList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("deleting list in list handler")

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transaction in list handler, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

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

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit trasnaction when trying to delete list with id %s, error %s", listId, err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) HandleDeleteLists(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("deleting lists in list handler")

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transaction in list handler, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err = h.serv.DeleteLists(ctx); err != nil {
		log.C(ctx).Errorf("failed to delete lists, error %s when calling list service", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit trasnaction when trying to delete lists, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) HandleCreateList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("creating list in list handler")

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transaction in list handler, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	var list handler_models.CreateList
	if err = json.NewDecoder(r.Body).Decode(&list); err != nil {
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

	if err = json.NewEncoder(w).Encode(l); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit trasnaction when trying to delete lists, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
func (h *Handler) HandleUpdateListPartially(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Debug("changing list name in list handler")

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transaction in list handler, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

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

	if err = json.NewEncoder(w).Encode(&updatedModel); err != nil {
		log.C(ctx).Errorf("failed to encode JSON %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit trasnaction when trying to delete lists, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}

func (h *Handler) HandleAddCollaborator(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Debug("adding a collaborator in list handler")

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transaction in list handler, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

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

	modelUser, err := h.serv.AddCollaborator(ctx, listId, user.Email)
	if err != nil {
		log.C(ctx).Error("failed to add collaborator in list handler, error when calling service function")

		utils.EncodeErrorWithCorrectStatusCode(w, err)
		return
	}

	if err = json.NewEncoder(w).Encode(modelUser); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit trasnaction when trying to delete lists, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) HandleDeleteCollaborator(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("deleting collaborator in list handler")

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transaction in list handler, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

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

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit trasnaction when trying to delete lists, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) HandleGetListRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("getting list record in list handler")

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transaction in list handler, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

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

	if err = json.NewEncoder(w).Encode(list); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit trasnaction when trying to delete lists, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleGetListOwner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("getting list owner in list handler")

	tx, err := h.transact.BeginContext(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to begin transaction in list handler, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

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

	if err = json.NewEncoder(w).Encode(listOwner); err != nil {
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).Errorf("failed to commit trasnaction when trying to delete lists, error %s", err.Error())
		utils.EncodeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
