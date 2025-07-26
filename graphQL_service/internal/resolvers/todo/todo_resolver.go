package todo

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

type httpResponseGetter interface {
	GetHttpResponseWithAuthHeader(context.Context, string, string, io.Reader) (*http.Response, error)
}

type urlDecoratorFactory interface {
	CreateUrlDecorator(context.Context, string, url_decorators.UrlFilters) url_decorators.QueryParamsRetrievers
}

type jsonMarshaller interface {
	Marshal(interface{}) ([]byte, error)
}

type todoConverter interface {
	ToGQL(*models.Todo) *gql.Todo
	ManyToGQL([]*models.Todo) []*gql.Todo
	ToHandlerModel(*gql.UpdateTodoInput) *handler_models.UpdateTodo
	CreateTodoInputToModel(*gql.CreateTodoInput) *handler_models.CreateTodo
	FromGQLModelToDeleteTodoPayload(*gql.Todo, bool) *gql.DeleteTodoPayload
	ManyToDeleteTodoPayload([]*gql.Todo, bool) []*gql.DeleteTodoPayload
}

type listConverter interface {
	ToGQL(*models.List) *gql.List
}

type userConverter interface {
	ToGQL(*models.User) *gql.User
}

type resolver struct {
	factory        urlDecoratorFactory
	tConverter     todoConverter
	uConverter     userConverter
	lConverter     listConverter
	restUrl        string
	jsonMarshaller jsonMarshaller
	responseGetter httpResponseGetter
}

func NewResolver(factory urlDecoratorFactory, tConverter todoConverter, uConverter userConverter, lConverter listConverter, restUrl string, jsonMarshaller jsonMarshaller, responseGetter httpResponseGetter) *resolver {
	return &resolver{
		factory:        factory,
		tConverter:     tConverter,
		uConverter:     uConverter,
		lConverter:     lConverter,
		restUrl:        restUrl,
		jsonMarshaller: jsonMarshaller,
		responseGetter: responseGetter,
	}
}

func (r *resolver) Todos(ctx context.Context, filter *url_filters.TodoFilters) (*gql.TodoPage, error) {
	log.C(ctx).Info("getting todos in todo resolver")

	decorator := r.factory.CreateUrlDecorator(ctx, gql_constants.TODO_PATH, filter)

	url, err := decorator.DetermineCorrectQueryParams(ctx, r.restUrl)

	log.C(ctx).Errorf("greshkata batyo %s", url)

	if err != nil {
		log.C(ctx).Errorf("failed to determine correct query param in todo resolver")
		return nil, err
	}

	resp, err := r.responseGetter.GetHttpResponseWithAuthHeader(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response in todo resolver, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if err = utils.HandleHttpCode(resp.StatusCode); err != nil {
		log.C(ctx).Errorf("failed to get todos in todo resolver, error %s due to bad response status code", err.Error())
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

func (r *resolver) Todo(ctx context.Context, id string) (*gql.Todo, error) {
	log.C(ctx).Infof("getting todo with id %s in todo resolver", id)

	todo, err := r.getModelTodo(ctx, id)
	if err != nil {
		log.C(ctx).Errorf("failed to get model todo in todo resolver %s", err.Error())
		return nil, err
	}

	if todo == nil {
		log.C(ctx).Info("http status not found received when calling rest api, empty struct...t")
		return nil, nil
	}

	return r.tConverter.ToGQL(todo), nil
}

func (r *resolver) DeleteTodosByListID(ctx context.Context, id string) ([]*gql.DeleteTodoPayload, error) {
	log.C(ctx).Infof("deleting todos from list with id %s in todo resolver", id)

	formattedSuffix := fmt.Sprintf("/%s%s", id, gql_constants.TODO_PATH)
	url := r.restUrl + gql_constants.LISTS_PATH + formattedSuffix

	gqlTodos, err := r.getTodosByListId(ctx, id)
	if err != nil {
		log.C(ctx).Errorf("failed to delete todos by list_id %s, error when trying to get list's todos", id)
		return nil, err
	}

	resp, err := r.responseGetter.GetHttpResponseWithAuthHeader(ctx, http.MethodDelete, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response in todo resolver, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	return r.tConverter.ManyToDeleteTodoPayload(gqlTodos, true), nil
}

func (r *resolver) UpdateTodo(ctx context.Context, id string, input gql.UpdateTodoInput) (*gql.Todo, error) {
	log.C(ctx).Info("updating todo in todo resolver")

	restModel := r.tConverter.ToHandlerModel(&input)

	formattedSuffix := fmt.Sprintf("/%s", id)
	url := r.restUrl + gql_constants.TODO_PATH + formattedSuffix

	jsonBody, err := r.jsonMarshaller.Marshal(restModel)
	if err != nil {
		log.C(ctx).Errorf("failed to updated todo with id %s, error when trying to marshal handler todo model %s", id, err.Error())
		return nil, err
	}

	resp, err := r.responseGetter.GetHttpResponseWithAuthHeader(ctx, http.MethodPatch, url, bytes.NewReader(jsonBody))
	if err != nil {
		log.C(ctx).Errorf("failed to get http response in todo resolver, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		log.C(ctx).Debugf("http status not found received...")
		return nil, nil
	}

	if err = utils.HandleHttpCode(resp.StatusCode); err != nil {
		log.C(ctx).Errorf("failed to update todo in todo resolver, error %s due to bad response status code", err.Error())
		return nil, err
	}

	var todo models.Todo
	if err = json.NewDecoder(resp.Body).Decode(&todo); err != nil {
		log.C(ctx).Errorf("failed to decode http response body, error %s", err.Error())
		return nil, err
	}

	return r.tConverter.ToGQL(&todo), nil
}

func (r *resolver) DeleteTodo(ctx context.Context, id string) (*gql.DeleteTodoPayload, error) {
	log.C(ctx).Infof("deleting todo with id %s in todo resolver", id)

	gqlTodo, err := r.Todo(ctx, id)
	if err != nil {
		log.C(ctx).Errorf("failed to delete todo with id %s, error when trying to get todo", id)
		return &gql.DeleteTodoPayload{
			Success: false,
		}, err
	}

	if gqlTodo == nil {
		log.C(ctx).Infof("todo with id %s not found", id)
		return &gql.DeleteTodoPayload{
			Success: false,
		}, nil
	}

	formattedSuffix := fmt.Sprintf("/%s", id)
	url := r.restUrl + gql_constants.TODO_PATH + formattedSuffix

	resp, err := r.responseGetter.GetHttpResponseWithAuthHeader(ctx, http.MethodDelete, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response in todo resolver, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	return r.tConverter.FromGQLModelToDeleteTodoPayload(gqlTodo, true), nil
}

func (r *resolver) DeleteTodos(ctx context.Context) ([]*gql.DeleteTodoPayload, error) {
	log.C(ctx).Info("deleting all todos in todo resolver")

	gqlTodos, err := r.getTodos(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to delete all todos in todo resolver, error when trying to get all todos %s", err.Error())
		return nil, err
	}

	url := r.restUrl + gql_constants.TODO_PATH

	resp, err := r.responseGetter.GetHttpResponseWithAuthHeader(ctx, http.MethodDelete, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response in todo resolver, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	return r.tConverter.ManyToDeleteTodoPayload(gqlTodos, true), nil
}

func (r *resolver) CreateTodo(ctx context.Context, input gql.CreateTodoInput) (*gql.Todo, error) {
	log.C(ctx).Info("creating todo in todo resolver")

	restModelTodo := r.tConverter.CreateTodoInputToModel(&input)

	url := r.restUrl + gql_constants.TODO_PATH

	jsonBody, err := r.jsonMarshaller.Marshal(restModelTodo)
	if err != nil {
		log.C(ctx).Errorf("failed to JSON marshal model Todo %s", err.Error())
		return nil, err
	}

	resp, err := r.responseGetter.GetHttpResponseWithAuthHeader(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		log.C(ctx).Errorf("failed to get http response in todo resolver, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if err = utils.HandleHttpCode(resp.StatusCode); err != nil {
		log.C(ctx).Errorf("failed to create todo in todo resolver, error %s due to bad response status code", err.Error())
		return nil, err
	}

	var todo models.Todo
	if err = json.NewDecoder(resp.Body).Decode(&todo); err != nil {
		log.C(ctx).Errorf("failed to decode http response body, error %s", err.Error())
		return nil, err
	}

	return r.tConverter.ToGQL(&todo), nil
}

func (r *resolver) AssignedTo(ctx context.Context, obj *gql.Todo) (*gql.User, error) {
	log.C(ctx).Info("getting todo assignee in todo resolver")

	formattedSuffix := fmt.Sprintf("/%s/assignee", obj.ID)
	url := r.restUrl + gql_constants.TODO_PATH + formattedSuffix

	resp, err := r.responseGetter.GetHttpResponseWithAuthHeader(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response in todo resolver, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if err = utils.HandleHttpCode(resp.StatusCode); err != nil {
		log.C(ctx).Errorf("failed to get user assigend to todo in todo resolver, error %s due to bad response status code", err.Error())
		return nil, err
	}

	var user models.User
	if err = json.NewDecoder(resp.Body).Decode(&user); err != nil {
		log.C(ctx).Errorf("failed to decode http response body, error %s", err.Error())
		return nil, err
	}

	return r.uConverter.ToGQL(&user), nil
}

func (r *resolver) List(ctx context.Context, obj *gql.Todo) (*gql.List, error) {
	log.C(ctx).Info("getting todo's list in todo resolver")

	modelTodo, err := r.getModelTodo(ctx, obj.ID)
	if err != nil {
		log.C(ctx).Error("failed to get todo list, error when trying to get model todo")
		return nil, err
	}

	formattedSuffix := fmt.Sprintf("/%s", modelTodo.ListId)
	url := r.restUrl + gql_constants.LISTS_PATH + formattedSuffix

	resp, err := r.responseGetter.GetHttpResponseWithAuthHeader(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response in todo resolver, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if err = utils.HandleHttpCode(resp.StatusCode); err != nil {
		log.C(ctx).Errorf("failed to get list where todo is in todo resolver, error %s due to bad response status code", err.Error())
		return nil, err
	}

	var list models.List
	if err = json.NewDecoder(resp.Body).Decode(&list); err != nil {
		log.C(ctx).Errorf("failed to decode http response body, error %s", err.Error())
		return nil, err
	}

	return r.lConverter.ToGQL(&list), nil
}

func (r *resolver) getModelTodo(ctx context.Context, id string) (*models.Todo, error) {
	log.C(ctx).Info("getting model todo")

	formattedSuffix := fmt.Sprintf("/%s", id)
	url := r.restUrl + gql_constants.TODO_PATH + formattedSuffix

	resp, err := r.responseGetter.GetHttpResponseWithAuthHeader(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response in todo resolver, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		log.C(ctx).Debugf("http status not found received...")
		return nil, nil
	}

	if err = utils.HandleHttpCode(resp.StatusCode); err != nil {
		log.C(ctx).Errorf("failed to get model todo in todo resolver, error %s due to bad response status code", err.Error())
		return nil, err
	}

	var todo models.Todo
	if err = json.NewDecoder(resp.Body).Decode(&todo); err != nil {
		log.C(ctx).Errorf("failed to decode http response body, error %s", err.Error())
		return nil, err
	}

	return &todo, nil
}

func (r *resolver) getTodosByListId(ctx context.Context, listId string) ([]*gql.Todo, error) {
	log.C(ctx).Infof("getting todos by list_id %s", listId)

	formattedSuffix := fmt.Sprintf("/%s%s", listId, gql_constants.TODO_PATH)
	url := r.restUrl + gql_constants.LISTS_PATH + formattedSuffix

	resp, err := r.responseGetter.GetHttpResponseWithAuthHeader(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response in todo resolver, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		log.C(ctx).Debugf("http status not found received...")
		return nil, nil
	}

	if err = utils.HandleHttpCode(resp.StatusCode); err != nil {
		log.C(ctx).Errorf("failed to get todos by list id %s in todo resolver, error %s due to bad response status code", listId, err.Error())
		return nil, err
	}

	var todos []*models.Todo
	if err = json.NewDecoder(resp.Body).Decode(&todos); err != nil {
		log.C(ctx).Errorf("failed to decode http response, error %s", err.Error())
		return nil, err
	}

	return r.tConverter.ManyToGQL(todos), nil
}

func (r *resolver) getTodos(ctx context.Context) ([]*gql.Todo, error) {
	log.C(ctx).Info("getting todos without filters in todo resolver")

	url := r.restUrl + gql_constants.TODO_PATH

	resp, err := r.responseGetter.GetHttpResponseWithAuthHeader(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response in todo resolver, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if err = utils.HandleHttpCode(resp.StatusCode); err != nil {
		log.C(ctx).Errorf("failed to get todos without filters in todo resolver, error %s due to bad response status code", err.Error())
		return nil, err
	}

	var todos []*models.Todo
	if err = json.NewDecoder(resp.Body).Decode(&todos); err != nil {
		log.C(ctx).Errorf("failed to decode http response body, error %s", err.Error())
		return nil, err
	}

	return r.tConverter.ManyToGQL(todos), nil
}
