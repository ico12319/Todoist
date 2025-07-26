package list

import (
	gql "Todo-List/internProject/graphQL_service/graph/model"
	"Todo-List/internProject/graphQL_service/graph/utils"
	"Todo-List/internProject/graphQL_service/internal/gql_constants"
	"Todo-List/internProject/graphQL_service/internal/url_decorators"
	"Todo-List/internProject/graphQL_service/internal/url_decorators/url_filters"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/handler_models"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

//go:generate mockery --name=jsonMarshaller --exported --output=./mocks --outpkg=mocks --filename=json_marshaller.go --with-expecter=true
type jsonMarshaller interface {
	Marshal(interface{}) ([]byte, error)
}

type httpResponseGetter interface {
	GetHttpResponseWithAuthHeader(context.Context, string, string, io.Reader) (*http.Response, error)
}

type urlDecoratorFactory interface {
	CreateUrlDecorator(context.Context, string, url_decorators.UrlFilters) url_decorators.QueryParamsRetrievers
}

//go:generate mockery --name=listConverter --output=./mocks --outpkg=mocks --filename=list_converter.go --with-expecter=true
type listConverter interface {
	ToGQL(*models.List) *gql.List
	ManyToGQL([]*models.List) []*gql.List
	ToModel(*gql.List) *models.List
	CreateListInputGQLToHandlerModel(gql.CreateListInput) *handler_models.CreateList
	UpdateListInputGQLToHandlerModel(gql.UpdateListInput) *handler_models.UpdateList
	FromGQLModelToDeleteListPayload(*gql.List, bool) *gql.DeleteListPayload
	ManyFromGQLModelToDeleteListPayload([]*gql.List, bool) []*gql.DeleteListPayload
}

//go:generate mockery --name=userConverter --output=./mocks --outpkg=mocks --filename=user_converter.go --with-expecter=true
type userConverter interface {
	ToGQL(*models.User) *gql.User
	ManyToGQL([]*models.User) []*gql.User
	FromCollaboratorInputToAddCollaboratorHandlerModel(*gql.CollaboratorInput) *handler_models.AddCollaborator
}

//go:generate mockery --name=todoConverter --output=./mocks --outpkg=mocks --filename=todo_converter.go --with-expecter=true
type todoConverter interface {
	ToGQL(*models.Todo) *gql.Todo
	ManyToGQL([]*models.Todo) []*gql.Todo
}

type resolver struct {
	lConverter     listConverter
	uConverter     userConverter
	tConverter     todoConverter
	restUrl        string
	factory        urlDecoratorFactory
	responseGetter httpResponseGetter
	marshaller     jsonMarshaller
}

func NewResolver(lConverter listConverter, uConverter userConverter, tConverter todoConverter, restUrl string, factory urlDecoratorFactory, responseGetter httpResponseGetter, marshaller jsonMarshaller) *resolver {
	return &resolver{
		lConverter:     lConverter,
		uConverter:     uConverter,
		tConverter:     tConverter,
		restUrl:        restUrl,
		factory:        factory,
		responseGetter: responseGetter,
		marshaller:     marshaller,
	}
}

func (r *resolver) Lists(ctx context.Context, filter *url_filters.BaseFilters) (*gql.ListPage, error) {
	log.C(ctx).Debugf("getting lists in list resolver")

	decorator := r.factory.CreateUrlDecorator(ctx, gql_constants.LISTS_PATH, filter)

	url, err := decorator.DetermineCorrectQueryParams(ctx, r.restUrl)
	if err != nil {
		log.C(ctx).Errorf("failed to get lists in list resolver, error when calling factory function")
		return nil, err
	}

	resp, err := r.responseGetter.GetHttpResponseWithAuthHeader(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if err = utils.HandleHttpCode(resp.StatusCode); err != nil {
		log.C(ctx).Errorf("failed to get collaborators in a list in list resolver, error %s due to bad response status code", err.Error())
		return nil, err
	}

	var lists []*models.List
	if err = json.NewDecoder(resp.Body).Decode(&lists); err != nil {
		log.C(ctx).Errorf("failed to decode http response body, error %s", err.Error())
		return nil, err
	}

	gqlModels := r.lConverter.ManyToGQL(lists)

	pageInfo := utils.InitPageInfo[*gql.List](gqlModels, func(list *gql.List) string {
		return list.ID
	})

	return &gql.ListPage{
		Data:       gqlModels,
		PageInfo:   pageInfo,
		TotalCount: int32(len(gqlModels)),
	}, nil
}

func (r *resolver) List(ctx context.Context, id string) (*gql.List, error) {
	log.C(ctx).Infof("getting list with id %s in list resover", id)

	formatedId := fmt.Sprintf("/%s", id)
	url := r.restUrl + gql_constants.LISTS_PATH + formatedId

	resp, err := r.responseGetter.GetHttpResponseWithAuthHeader(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		log.C(ctx).Debugf("http status not found received...")
		return nil, nil
	}

	if err = utils.HandleHttpCode(resp.StatusCode); err != nil {
		log.C(ctx).Errorf("failed to get collaborators in a list in list resolver, error %s due to bad response status code", err.Error())
		return nil, err
	}

	var list models.List
	if err = json.NewDecoder(resp.Body).Decode(&list); err != nil {
		log.C(ctx).Errorf("failed to decode http response body, error %s", err.Error())
		return nil, err
	}

	return r.lConverter.ToGQL(&list), nil
}

func (r *resolver) ListOwner(ctx context.Context, obj *gql.List) (*gql.User, error) {
	log.C(ctx).Infof("getting list %s owner in list resolver", obj.ID)

	formatedSuffix := fmt.Sprintf("/%s/owner", obj.ID)
	url := r.restUrl + gql_constants.LISTS_PATH + formatedSuffix

	resp, err := r.responseGetter.GetHttpResponseWithAuthHeader(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if err = utils.HandleHttpCode(resp.StatusCode); err != nil {
		log.C(ctx).Errorf("failed to get collaborators in a list in list resolver, error %s due to bad response status code", err.Error())
		return nil, err
	}

	var owner models.User
	if err = json.NewDecoder(resp.Body).Decode(&owner); err != nil {
		log.C(ctx).Errorf("failed to decode http response body, error %s", err.Error())
		return nil, err
	}

	return r.uConverter.ToGQL(&owner), nil
}

func (r *resolver) Todos(ctx context.Context, obj *gql.List, filters *url_filters.TodoFilters) (*gql.TodoPage, error) {
	log.C(ctx).Debugf("getting todos of list with id %s", obj.ID)

	formatedSuffix := fmt.Sprintf("/%s", obj.ID)

	decorator := r.factory.CreateUrlDecorator(ctx, gql_constants.LISTS_PATH+formatedSuffix+gql_constants.TODO_PATH, filters)

	url, err := decorator.DetermineCorrectQueryParams(ctx, r.restUrl)
	if err != nil {
		log.C(ctx).Errorf("failed to get list todos, error when calling factory function")
		return utils.InitEmptyTodoPage(), err
	}

	resp, err := r.responseGetter.GetHttpResponseWithAuthHeader(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if err = utils.HandleHttpCode(resp.StatusCode); err != nil {
		log.C(ctx).Errorf("failed to get collaborators in a list in list resolver, error %s due to bad response status code", err.Error())
		return nil, err
	}

	var todos []*models.Todo
	if err = json.NewDecoder(resp.Body).Decode(&todos); err != nil {
		log.C(ctx).Errorf("failed to decode http response body, error %s", err.Error())
		return nil, err
	}

	gqlTodos := r.tConverter.ManyToGQL(todos)

	pageInfo := utils.InitPageInfo[*gql.Todo](gqlTodos, func(todo *gql.Todo) string {
		return todo.ID
	})

	return &gql.TodoPage{
		Data:       gqlTodos,
		PageInfo:   pageInfo,
		TotalCount: int32(len(gqlTodos)),
	}, nil
}

func (r *resolver) DeleteList(ctx context.Context, id string) (*gql.DeleteListPayload, error) {
	log.C(ctx).Infof("deleting list with id %s in list resolver", id)

	gqlList, err := r.List(ctx, id)
	if err != nil {
		log.C(ctx).Errorf("failed to delete list with id %s, error when trying to get list", id)
		return nil, err
	}

	if gqlList == nil {
		log.C(ctx).Errorf("failed to delete list with id %s, error gqlList is nil", id)
		return &gql.DeleteListPayload{
			Success: false,
		}, nil
	}

	formattedUrl := fmt.Sprintf("/%s", id)
	url := r.restUrl + gql_constants.LISTS_PATH + formattedUrl

	resp, err := r.responseGetter.GetHttpResponseWithAuthHeader(ctx, http.MethodDelete, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	return r.lConverter.FromGQLModelToDeleteListPayload(gqlList, true), nil
}

func (r *resolver) UpdateList(ctx context.Context, id string, input gql.UpdateListInput) (*gql.List, error) {
	log.C(ctx).Infof("updating list with id %s in list resolver", id)

	formattedSuffix := fmt.Sprintf("/%s", id)
	url := r.restUrl + gql_constants.LISTS_PATH + formattedSuffix

	restModel := r.lConverter.UpdateListInputGQLToHandlerModel(input)

	jsonBody, err := r.marshaller.Marshal(restModel)
	if err != nil {
		log.C(ctx).Errorf("failed to update list with id %s, error when trying to marshal rest model", id)
		return nil, err
	}

	resp, err := r.responseGetter.GetHttpResponseWithAuthHeader(ctx, http.MethodPatch, url, bytes.NewReader(jsonBody))
	if err != nil {
		log.C(ctx).Errorf("failed to get http response, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		log.C(ctx).Debugf("http status not found received...")
		return nil, nil
	}

	if err = utils.HandleHttpCode(resp.StatusCode); err != nil {
		log.C(ctx).Errorf("failed to get collaborators in a list in list resolver, error %s due to bad response status code", err.Error())
		return nil, err
	}

	var updatedList models.List
	if err = json.NewDecoder(resp.Body).Decode(&updatedList); err != nil {
		log.C(ctx).Errorf("failed to decode http response body, error %s", err.Error())
		return nil, err
	}

	return r.lConverter.ToGQL(&updatedList), nil
}

func (r *resolver) AddListCollaborator(ctx context.Context, input gql.CollaboratorInput) (*gql.CreateCollaboratorPayload, error) {
	log.C(ctx).Debugf("adding collaborator %s in list %s", input.UserID, input.ListID)

	formattedUrl := fmt.Sprintf("/%s%s", input.ListID, gql_constants.COLLABORATOR_PATH)
	url := r.restUrl + gql_constants.LISTS_PATH + formattedUrl

	restModel := r.uConverter.FromCollaboratorInputToAddCollaboratorHandlerModel(&input)

	jsonBody, err := r.marshaller.Marshal(restModel)
	if err != nil {
		log.C(ctx).Errorf("failed to add collaborator, error when trying to matshal user")
		return &gql.CreateCollaboratorPayload{
			Success: false,
		}, err
	}

	resp, err := r.responseGetter.GetHttpResponseWithAuthHeader(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		log.C(ctx).Errorf("failed to get http response, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		log.C(ctx).Debugf("http status not found received...")
		return &gql.CreateCollaboratorPayload{
			Success: false,
		}, nil
	}

	if err = utils.HandleHttpCode(resp.StatusCode); err != nil {
		log.C(ctx).Errorf("failed to get lists in list resolver, error %s due to bad response status code", err.Error())
		return &gql.CreateCollaboratorPayload{
			Success: false,
		}, err
	}

	var collaborator models.User
	if err = json.NewDecoder(resp.Body).Decode(&collaborator); err != nil {
		log.C(ctx).Errorf("failed to decode json body, error %s", err.Error())
		return nil, err
	}

	gqlUser := r.uConverter.ToGQL(&collaborator)

	gqlList, err := r.List(ctx, input.ListID)
	if err != nil {
		log.C(ctx).Error("failed to get gql list in list resolver, error when calling list resolver function")
		return &gql.CreateCollaboratorPayload{
			Success: false,
		}, err
	}

	return &gql.CreateCollaboratorPayload{
		List:    gqlList,
		User:    gqlUser,
		Success: true,
	}, nil
}

func (r *resolver) DeleteListCollaborator(ctx context.Context, id string, userID string) (*gql.DeleteCollaboratorPayload, error) {
	log.C(ctx).Infof("deleting a collaborator with id %s from a list with id %s", userID, id)

	formattedSuffix := fmt.Sprintf("/%s%s/%s", id, gql_constants.COLLABORATOR_PATH, userID)
	url := r.restUrl + gql_constants.LISTS_PATH + formattedSuffix

	resp, err := r.responseGetter.GetHttpResponseWithAuthHeader(ctx, http.MethodDelete, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to delete collaborator with id %s, error when trying to get http response", userID)
		return &gql.DeleteCollaboratorPayload{
			ListID:  id,
			UserID:  userID,
			Success: false,
		}, err
	}
	defer resp.Body.Close()

	return &gql.DeleteCollaboratorPayload{
		ListID:  id,
		UserID:  userID,
		Success: true,
	}, nil
}

func (r *resolver) Collaborators(ctx context.Context, obj *gql.List, filters *url_filters.BaseFilters) (*gql.UserPage, error) {
	log.C(ctx).Info("getting list collaborators in list resolver")

	formattedSuffix := fmt.Sprintf("/%s", obj.ID)

	decorator := r.factory.CreateUrlDecorator(ctx, gql_constants.LISTS_PATH+formattedSuffix+gql_constants.COLLABORATOR_PATH, filters)

	url, err := decorator.DetermineCorrectQueryParams(ctx, r.restUrl)
	if err != nil {
		log.C(ctx).Error("failed to get list collaborators, error when calling common decorator function")
		return nil, err
	}

	resp, err := r.responseGetter.GetHttpResponseWithAuthHeader(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if err = utils.HandleHttpCode(resp.StatusCode); err != nil {
		log.C(ctx).Errorf("failed to get collaborators in a list in list resolver, error %s due to bad response status code", err.Error())
		return nil, err
	}

	var collaborators []*models.User
	if err = json.NewDecoder(resp.Body).Decode(&collaborators); err != nil {
		log.C(ctx).Errorf("failed to decode json body, error %s", err.Error())
		return nil, err
	}

	gqlUsers := r.uConverter.ManyToGQL(collaborators)

	pageInfo := utils.InitPageInfo[*gql.User](gqlUsers, func(user *gql.User) string {
		return user.ID
	})

	return &gql.UserPage{
		Data:       gqlUsers,
		PageInfo:   pageInfo,
		TotalCount: int32(len(gqlUsers)),
	}, nil
}

func (r *resolver) CreateList(ctx context.Context, input gql.CreateListInput) (*gql.List, error) {
	log.C(ctx).Info("creating list in list resolver")

	url := r.restUrl + gql_constants.LISTS_PATH

	restModel := r.lConverter.CreateListInputGQLToHandlerModel(input)

	jsonBody, err := r.marshaller.Marshal(restModel)
	if err != nil {
		log.C(ctx).Errorf("failed to create list, error %s when trying to marshal json body", err.Error())
		return nil, err
	}

	resp, err := r.responseGetter.GetHttpResponseWithAuthHeader(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		log.C(ctx).Errorf("failed to get http response, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if err = utils.HandleHttpCode(resp.StatusCode); err != nil {
		log.C(ctx).Errorf("failed to create list in list resolver, error %s due to bad response status code", err.Error())
		return nil, err
	}

	var list models.List
	if err = json.NewDecoder(resp.Body).Decode(&list); err != nil {
		log.C(ctx).Errorf("failed to decode json body, error %s", err.Error())
		return nil, err
	}

	return r.lConverter.ToGQL(&list), nil
}

func (r *resolver) DeleteLists(ctx context.Context) ([]*gql.DeleteListPayload, error) {
	log.C(ctx).Info("deleting all lists in list resolver")

	url := r.restUrl + gql_constants.LISTS_PATH

	gqlLists, err := r.getLists(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to delete lists, error %s when trying to get them", err.Error())
		return nil, err
	}

	resp, err := r.responseGetter.GetHttpResponseWithAuthHeader(ctx, http.MethodDelete, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	return r.lConverter.ManyFromGQLModelToDeleteListPayload(gqlLists, true), nil
}

func (r *resolver) getLists(ctx context.Context) ([]*gql.List, error) {
	log.C(ctx).Info("getting all lists without filters in list resolver")

	url := r.restUrl + gql_constants.LISTS_PATH

	resp, err := r.responseGetter.GetHttpResponseWithAuthHeader(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if err = utils.HandleHttpCode(resp.StatusCode); err != nil {
		log.C(ctx).Errorf("failed to get lists in list resolver, error %s due to bad response status code", err.Error())
		return nil, err
	}

	var lists []*models.List
	if err = json.NewDecoder(resp.Body).Decode(&lists); err != nil {
		log.C(ctx).Errorf("failed to decode json body, error %s", err.Error())
		return nil, err
	}

	return r.lConverter.ManyToGQL(lists), nil
}
