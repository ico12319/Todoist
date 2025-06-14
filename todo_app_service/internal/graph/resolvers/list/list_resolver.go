package list

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/gql_constants"
	gql "github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/model"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/url_decorators"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/url_decorators/url_filters"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/utils"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/configuration"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/handler_models"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/models"
	"net/http"
	"reflect"
)

//go:generate mockery --name=httpClient --output=./mocks --outpkg=mocks --filename=http_client.go --with-expecter=true
type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type requestAuthSetter interface {
	DecorateRequest(ctx context.Context, req *http.Request) (*http.Request, error)
}

//go:generate mockery --name=commonQueryParamsDecoratorFactory --output=./mocks --outpkg=mocks --filename=common_query_params_decorator_factory.go --with-expecter=true
type commonQueryParamsDecoratorFactory interface {
	CreateCommonUrlDecorator(ctx context.Context, initialUrl string, baseFilters *url_filters.BaseFilters) (url_decorators.QueryParamsRetrievers, error)
}

//go:generate mockery --name=concreteQueryParamsDecoratorFactory --output=./mocks --outpkg=mocks --filename=concrete_query_params_decorator_factory.go --with-expecter=true
type concreteQueryParamsDecoratorFactory interface {
	CreateConcreteUrlDecorator(ctx context.Context, initialUrl string, todoFilters *url_filters.TodoFilters) (url_decorators.QueryParamsRetrievers, error)
}

//go:generate mockery --name=listConverter --output=./mocks --outpkg=mocks --filename=list_converter.go --with-expecter=true
type listConverter interface {
	ToGQL(list *models.List) *gql.List
	ManyToGQL(lists []*models.List) []*gql.List
	ToModel(list *gql.List) *models.List
	CreateListInputGQLToHandlerModel(input gql.CreateListInput) *handler_models.CreateList
	UpdateListInputGQLToHandlerModel(input gql.UpdateListInput) *handler_models.UpdateList
	FromGQLModelToDeleteListPayload(list *gql.List, success bool) *gql.DeleteListPayload
	ManyFromGQLModelToDeleteListPayload(lists []*gql.List, success bool) []*gql.DeleteListPayload
}

//go:generate mockery --name=userConverter --output=./mocks --outpkg=mocks --filename=user_converter.go --with-expecter=true
type userConverter interface {
	ToGQL(user *models.User) *gql.User
	ManyToGQL(users []*models.User) []*gql.User
	FromCollaboratorInputToAddCollaboratorHandlerModel(user *gql.CollaboratorInput) *handler_models.AddCollaborator
}

//go:generate mockery --name=todoConverter --output=./mocks --outpkg=mocks --filename=todo_converter.go --with-expecter=true
type todoConverter interface {
	ToGQL(todo *models.Todo) *gql.Todo
	ManyToGQL(todos []*models.Todo) []*gql.Todo
}

type resolver struct {
	client          httpClient
	lConverter      listConverter
	uConverter      userConverter
	tConverter      todoConverter
	restUrl         string
	factory         commonQueryParamsDecoratorFactory
	concreteFactory concreteQueryParamsDecoratorFactory
	authSetter      requestAuthSetter
}

func NewResolver(client httpClient, lConverter listConverter, uConverter userConverter, tConverter todoConverter, restUrl string, factory commonQueryParamsDecoratorFactory, concreteFactory concreteQueryParamsDecoratorFactory, authSetter requestAuthSetter) *resolver {
	return &resolver{client: client, lConverter: lConverter, uConverter: uConverter, tConverter: tConverter, restUrl: restUrl, factory: factory, concreteFactory: concreteFactory, authSetter: authSetter}
}

func (r *resolver) Lists(ctx context.Context, filter *url_filters.BaseFilters) (*gql.ListPage, error) {
	log.C(ctx).Debugf("getting lists in list resolver")

	decorator, err := r.factory.CreateCommonUrlDecorator(ctx, gql_constants.LISTS_PATH, filter)
	if err != nil {
		log.C(ctx).Errorf("failed to create common decorator in list resolver, error when calling common factory function")
		return nil, err
	}

	url, err := decorator.DetermineCorrectQueryParams(ctx, r.restUrl)
	if err != nil {
		log.C(ctx).Errorf("failed to get lists in list resolver, error when calling factory function")
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get lists in list resolver, error %s when making request", err.Error())
		return nil, fmt.Errorf("error when making http request")
	}

	req, err = r.authSetter.DecorateRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorf("faield to decorate request auth header, error %s", err.Error())
		return nil, err
	}

	response, err := r.client.Do(req)
	if err != nil {
		log.C(ctx).Errorf("failed to get lists in list resolver, error %s when trying to get response", err.Error())
		return utils.InitEmptyListPage(), err
	}
	defer response.Body.Close()

	listModels, err := utils.HandleHttpCode[[]*models.List](response)
	if err != nil {
		log.C(ctx).Errorf("failed to get lists in list resolver, error %s when trying to decode JSON", err.Error())
		return utils.InitEmptyListPage(), err
	}

	gqlModels := r.lConverter.ManyToGQL(listModels)

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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get list with id %s, error when making request", err.Error())
		return nil, fmt.Errorf("error when making http request")
	}

	req, err = r.authSetter.DecorateRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorf("faield to decorate request auth header, error %s", err.Error())
		return nil, err
	}

	response, err := r.client.Do(req)
	if err != nil {
		log.C(ctx).Errorf("failed to get list with id %s, error when trying to get response", err.Error())
		return nil, fmt.Errorf("failed to fetch list response")
	}
	defer response.Body.Close()

	listModel, err := utils.HandleHttpCode[*models.List](response)
	if err != nil {
		log.C(ctx).Errorf("bad http status code %d received when calling rest api", response.StatusCode)
		return nil, err
	}

	if reflect.ValueOf(listModel).IsZero() {
		log.C(ctx).Error("http status code not found received, empty struct...")
		return nil, nil
	}

	return r.lConverter.ToGQL(listModel), nil
}

func (r *resolver) ListOwner(ctx context.Context, obj *gql.List) (*gql.User, error) {
	log.C(ctx).Infof("getting list %s owner in list resolver", obj.ID)

	formatedSuffix := fmt.Sprintf("/%s/owner", obj.ID)
	url := r.restUrl + gql_constants.LISTS_PATH + formatedSuffix

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Error("failed to get list owner, error in list resolcer when making http request")
		return nil, fmt.Errorf("error when making http request")
	}

	req, err = r.authSetter.DecorateRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorf("faield to decorate request auth header, error %s", err.Error())
		return nil, err
	}

	response, err := r.client.Do(req)
	if err != nil {
		log.C(ctx).Error("failed to get list owner, error in list resolver when trying to get http response")
		return nil, err
	}
	defer response.Body.Close()

	owner, err := utils.HandleHttpCode[*models.User](response)
	if err != nil {
		log.C(ctx).Errorf("bad http status code %d received when calling rest api", response.StatusCode)
		return nil, err
	}

	return r.uConverter.ToGQL(owner), nil
}

func (r *resolver) Todos(ctx context.Context, obj *gql.List, filters *url_filters.TodoFilters) (*gql.TodoPage, error) {
	log.C(ctx).Debugf("getting todos of list with id %s", obj.ID)

	formatedSuffix := fmt.Sprintf("/%s", obj.ID)
	decorator, err := r.concreteFactory.CreateConcreteUrlDecorator(ctx, gql_constants.LISTS_PATH+formatedSuffix+gql_constants.TODO_PATH, filters)
	if err != nil {
		log.C(ctx).Errorf("failed to create common decorator in list resolver, error when calling common factory function")
		return nil, err
	}

	url, err := decorator.DetermineCorrectQueryParams(ctx, r.restUrl)
	log.C(ctx).Errorf("wrong url is %s", url)
	if err != nil {
		log.C(ctx).Errorf("failed to get list todos, error when calling factory function")
		return utils.InitEmptyTodoPage(), err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get list's todos, error when making http request %s", err.Error())
		return utils.InitEmptyTodoPage(), fmt.Errorf("error when making http request")
	}

	req, err = r.authSetter.DecorateRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorf("faield to decorate request auth header, error %s", err.Error())
		return nil, err
	}

	response, err := r.client.Do(req)
	if err != nil {
		log.C(ctx).Errorf("failed to get list's todos, error when trying to get http response %s", err.Error())
		return utils.InitEmptyTodoPage(), err
	}
	defer response.Body.Close()

	todos, err := utils.HandleHttpCode[[]*models.Todo](response)
	if err != nil {
		log.C(ctx).Errorf("bad http status code %d received when calling rest api", response.StatusCode)
		return utils.InitEmptyTodoPage(), err
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

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to delete list with id %s, error when making http request %s", id, err.Error())
		return &gql.DeleteListPayload{
			Success: false,
		}, fmt.Errorf("error when making http request")
	}

	req, err = r.authSetter.DecorateRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorf("faield to decorate request auth header, error %s", err.Error())
		return nil, err
	}

	response, err := r.client.Do(req)
	if err != nil {
		log.C(ctx).Errorf("failed to delete list with id %s, error when trying to get http response %s", id, err.Error())
		return &gql.DeleteListPayload{
			Success: false,
		}, fmt.Errorf("failed to fetch list response")
	}
	defer response.Body.Close()

	return r.lConverter.FromGQLModelToDeleteListPayload(gqlList, true), nil
}

func (r *resolver) UpdateList(ctx context.Context, id string, input gql.UpdateListInput) (*gql.List, error) {
	log.C(ctx).Infof("updating list with id %s in list resolver", id)

	formattedSuffix := fmt.Sprintf("/%s", id)
	url := r.restUrl + gql_constants.LISTS_PATH + formattedSuffix

	restModel := r.lConverter.UpdateListInputGQLToHandlerModel(input)

	jsonBody, err := json.Marshal(restModel)
	if err != nil {
		log.C(ctx).Errorf("failed to update list with id %s, error when trying to marshal rest model", id)
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewReader(jsonBody))
	if err != nil {
		log.C(ctx).Errorf("failed to update list with id %s, error when making http request", id)
		return nil, fmt.Errorf("error when making http request")
	}

	req, err = r.authSetter.DecorateRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorf("faield to decorate request auth header, error %s", err.Error())
		return nil, err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		log.C(ctx).Errorf("failed to update list with id %s, error when trying to get http response", id)
		return nil, fmt.Errorf("failed to fetch list response")
	}
	defer resp.Body.Close()

	updatedList, err := utils.HandleHttpCode[*models.List](resp)
	if err != nil {
		log.C(ctx).Errorf("bad http status code %d received when calling rest api", resp.StatusCode)
		return nil, err
	}

	if reflect.ValueOf(updatedList).IsZero() {
		log.C(ctx).Infof("http status code not found received, empty struct...")
		return nil, nil
	}

	return r.lConverter.ToGQL(updatedList), nil
}

func (r *resolver) AddListCollaborator(ctx context.Context, input gql.CollaboratorInput) (*gql.CreateCollaboratorPayload, error) {
	log.C(ctx).Debugf("adding collaborator %s in list %s", input.UserID, input.ListID)

	formattedUrl := fmt.Sprintf("/%s%s", input.ListID, gql_constants.COLLABORATOR_PATH)
	url := r.restUrl + gql_constants.LISTS_PATH + formattedUrl

	restModel := r.uConverter.FromCollaboratorInputToAddCollaboratorHandlerModel(&input)

	jsonBody, err := json.Marshal(restModel)
	if err != nil {
		log.C(ctx).Errorf("failed to add collaborator, error when trying to matshal user")
		return &gql.CreateCollaboratorPayload{
			Success: false,
		}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		log.C(ctx).Errorf("failed to add collaborator, error when making http request")
		return &gql.CreateCollaboratorPayload{
			Success: false,
		}, fmt.Errorf("error when making http request")
	}

	req, err = r.authSetter.DecorateRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorf("faield to decorate request auth header, error %s", err.Error())
		return nil, err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		log.C(ctx).Errorf("failed to add collaborator, error when trying to get http response")
		return &gql.CreateCollaboratorPayload{
			Success: false,
		}, err
	}
	defer resp.Body.Close()

	collaborator, err := utils.HandleHttpCode[*models.User](resp)
	if err != nil {
		log.C(ctx).Errorf("bad http status code %d received when calling rest api", resp.StatusCode)
		return &gql.CreateCollaboratorPayload{
			Success: false,
		}, err
	}

	if reflect.ValueOf(collaborator).IsZero() {
		log.C(ctx).Infof("http status code not found received, empty struct...")
		return &gql.CreateCollaboratorPayload{
			Success: false,
		}, nil
	}

	gqlUser := r.uConverter.ToGQL(collaborator)

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
	log.C(ctx).Errorf("url is %s", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to delete collaborator with id %s, error when making http request", userID)
		return &gql.DeleteCollaboratorPayload{
			ListID:  id,
			UserID:  userID,
			Success: false,
		}, fmt.Errorf("error when making http request")
	}

	req, err = r.authSetter.DecorateRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorf("faield to decorate request auth header, error %s", err.Error())
		return nil, err
	}

	resp, err := r.client.Do(req)
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
	decorator, err := r.factory.CreateCommonUrlDecorator(ctx, gql_constants.LISTS_PATH+formattedSuffix+gql_constants.COLLABORATOR_PATH, filters)
	if err != nil {
		log.C(ctx).Error("failed to get list collaborators, error when calling common factory function")
		return nil, err
	}

	url, err := decorator.DetermineCorrectQueryParams(ctx, r.restUrl)
	if err != nil {
		log.C(ctx).Error("failed to get list collaborators, error when calling common decorator function")
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Error("failed to get list collaborators, error when make http request")
		return nil, fmt.Errorf("error when making http request")
	}

	req, err = r.authSetter.DecorateRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorf("faield to decorate request auth header, error %s", err.Error())
		return nil, err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		log.C(ctx).Error("failed to get list collaborators, error when trying to get http response")
		return nil, err
	}
	defer resp.Body.Close()

	collaborators, err := utils.HandleHttpCode[[]*models.User](resp)
	if err != nil {
		log.C(ctx).Errorf("bad http status code %d received when calling rest api", resp.StatusCode)
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

	jsonBody, err := json.Marshal(restModel)
	if err != nil {
		log.C(ctx).Errorf("failed to create list, error %s when trying to marshal json body", err.Error())
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		log.C(ctx).Errorf("failed to create user, error %s when making http requet", err.Error())
		return nil, fmt.Errorf("error when making http request")
	}

	req, err = r.authSetter.DecorateRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorf("faield to decorate request auth header, error %s", err.Error())
		return nil, err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		log.C(ctx).Errorf("failed to create list, error %s when trying to get http response", err.Error())
		return nil, fmt.Errorf("failed to fetch list response")
	}
	defer resp.Body.Close()

	list, err := utils.HandleHttpCode[*models.List](resp)
	if err != nil {
		log.C(ctx).Errorf("bad http status code %d received when calling rest api", resp.StatusCode)
		return nil, err
	}

	return r.lConverter.ToGQL(list), nil
}

func (r *resolver) getModelList(ctx context.Context, id string) (*models.List, error) {
	log.C(ctx).Infof("getting model list with id %s in list resolver", id)

	formattedSuffix := fmt.Sprintf("/%s", id)
	url := r.restUrl + gql_constants.LISTS_PATH + formattedSuffix

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get model list with id %s, error when making http request", id)
		return nil, fmt.Errorf("error when making http request")
	}

	req, err = r.authSetter.DecorateRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorf("faield to decorate request auth header, error %s", err.Error())
		return nil, err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		log.C(ctx).Errorf("failed to get model list with id %s, error when trying to get http response", id)
		return nil, err
	}
	defer resp.Body.Close()

	modelList, err := utils.HandleHttpCode[*models.List](resp)
	if err != nil {
		log.C(ctx).Errorf("bad http status code %d received when calling rest api", resp.StatusCode)
		return nil, err
	}

	if reflect.ValueOf(modelList).IsZero() {
		log.C(ctx).Infof("http status code not found received, empty struct...")
		return nil, nil
	}

	return modelList, nil
}

func (r *resolver) DeleteLists(ctx context.Context) ([]*gql.DeleteListPayload, error) {
	log.C(ctx).Info("deleting all lists in list resolver")

	url := r.restUrl + gql_constants.LISTS_PATH

	gqlLists, err := r.Lists(ctx, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to delete lists, error %s when trying to get them", err.Error())
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to delete lists, error %s when trying to make http request", err.Error())
		return nil, err
	}

	req, err = r.authSetter.DecorateRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorf("failed to decorate http request, error %s", err.Error())
		return nil, err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		log.C(ctx).Errorf("failed to delete lists, error %s when trying to get http response", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	return r.lConverter.ManyFromGQLModelToDeleteListPayload(gqlLists.Data, true), nil
}
