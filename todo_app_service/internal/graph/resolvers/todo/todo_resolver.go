package todo

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
	"github.com/vektah/gqlparser/v2/gqlerror"
	"net/http"
	"reflect"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type requestAuthSetter interface {
	DecorateRequest(ctx context.Context, req *http.Request) (*http.Request, error)
}

type concreteQueryParamsDecoratorFactory interface {
	CreateConcreteUrlDecorator(ctx context.Context, initialUrl string, todoFilters *url_filters.TodoFilters) (url_decorators.QueryParamsRetrievers, error)
}

type todoConverter interface {
	ToGQL(todo *models.Todo) *gql.Todo
	ManyToGQL(todos []*models.Todo) []*gql.Todo
	ToHandlerModel(todoInput *gql.UpdateTodoInput) *handler_models.UpdateTodo
	CreateTodoInputToModel(todoInput *gql.CreateTodoInput) *handler_models.CreateTodo
	FromGQLModelToDeleteTodoPayload(todo *gql.Todo, success bool) *gql.DeleteTodoPayload
	ManyToDeleteTodoPayload(todos []*gql.Todo, success bool) []*gql.DeleteTodoPayload
}

type listConverter interface {
	ToGQL(list *models.List) *gql.List
}

type userConverter interface {
	ToGQL(user *models.User) *gql.User
}

type resolver struct {
	client          httpClient
	concreteFactory concreteQueryParamsDecoratorFactory
	tConverter      todoConverter
	uConverter      userConverter
	lConverter      listConverter
	restUrl         string
	authSetter      requestAuthSetter
}

func NewResolver(client httpClient, concreteFactory concreteQueryParamsDecoratorFactory, tConverter todoConverter, uConverter userConverter, lConverter listConverter, restUrl string, authSetter requestAuthSetter) *resolver {
	return &resolver{
		client:          client,
		concreteFactory: concreteFactory,
		tConverter:      tConverter,
		uConverter:      uConverter,
		lConverter:      lConverter,
		restUrl:         restUrl,
		authSetter:      authSetter,
	}
}

func (r *resolver) Todos(ctx context.Context, filter *url_filters.TodoFilters) (*gql.TodoPage, error) {
	log.C(ctx).Info("getting todos in todo resolver")

	decorator, err := r.concreteFactory.CreateConcreteUrlDecorator(ctx, gql_constants.TODO_PATH, filter)
	if err != nil {
		log.C(ctx).Error("failed to get todos in todo resolver, error when calling concrete factory")
		return nil, err
	}

	requestUrl, err := decorator.DetermineCorrectQueryParams(ctx, r.restUrl)

	if err != nil {
		log.C(ctx).Errorf("failed to determine correct query param in todo resolver")
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestUrl, nil)

	if err != nil {
		log.C(ctx).Errorf("failed to get todos in todo resolver, error when making http request %s", err.Error())
		return nil, err
	}

	req, err = r.authSetter.DecorateRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorf("failed to decorate http request's auth header, error %s", err.Error())
		return nil, &gqlerror.Error{
			Message: err.Error(),
		}
	}

	resp, err := r.client.Do(req)
	if err != nil {
		log.C(ctx).Errorf("failed to get todos in todo resolver, error when trying to get response %s", err.Error())
		return nil, err
	}

	modelTodos, err := utils.HandleHttpCode[[]*models.Todo](resp)
	if err != nil {
		log.C(ctx).Error("failed to get todos in todo resolver, error when trying to decode JSON")
		return nil, err
	}

	gqlTodos := r.tConverter.ManyToGQL(modelTodos)

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

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to delete todos by list_id %s, error when making http request %s", id, err.Error())
		return r.tConverter.ManyToDeleteTodoPayload(gqlTodos, false), err
	}

	req, err = r.authSetter.DecorateRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorf("failed to decorate http request's auth header, error %s", err.Error())
		return nil, &gqlerror.Error{
			Message: "unauthorized user",
		}
	}

	resp, err := r.client.Do(req)

	if err != nil {
		log.C(ctx).Errorf("failed to delete todos by list_id %s, error when trying to get http response %s", id, err.Error())
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

	jsonBody, err := json.Marshal(restModel)
	if err != nil {
		log.C(ctx).Errorf("failed to updated todo with id %s, error when trying to marshal handler todo model %s", id, err.Error())
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewReader(jsonBody))
	if err != nil {
		log.C(ctx).Errorf("failed to update todo with id %s, error when trying to make http request %s", id, err)
		return nil, err
	}

	req, err = r.authSetter.DecorateRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorf("failed to decorate http request's auth header, error %s", err.Error())
		return nil, &gqlerror.Error{
			Message: "unauthorized user",
		}
	}

	resp, err := r.client.Do(req)

	if err != nil {
		log.C(ctx).Errorf("failed to update todo with id %s, error when trying to get http response %s", id, err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	modelTodo, err := utils.HandleHttpCode[*models.Todo](resp)
	if err != nil {
		log.C(ctx).Errorf("failed to update todo, error when trying to decode JSON body %s", err.Error())
		return nil, err
	}

	if reflect.ValueOf(modelTodo).IsZero() {
		log.C(ctx).Info("http status not found received when calling rest api, empty struct...")
		return nil, nil
	}

	return r.tConverter.ToGQL(modelTodo), nil
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

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to delete todo with id %s, error when calling http method %s", id, err.Error())
		return r.tConverter.FromGQLModelToDeleteTodoPayload(gqlTodo, false), err
	}

	req, err = r.authSetter.DecorateRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorf("failed to decorate http request's auth header, error %s", err.Error())
		return nil, &gqlerror.Error{
			Message: "unauthorized user",
		}
	}

	resp, err := r.client.Do(req)
	if err != nil {
		log.C(ctx).Errorf("failed to delete todo with id %s, error when trying to get response %s", id, err.Error())
		return r.tConverter.FromGQLModelToDeleteTodoPayload(gqlTodo, false), err
	}
	defer resp.Body.Close()

	return r.tConverter.FromGQLModelToDeleteTodoPayload(gqlTodo, true), nil
}

func (r *resolver) DeleteTodos(ctx context.Context) ([]*gql.DeleteTodoPayload, error) {
	log.C(ctx).Info("deleting all todos in todo resolver")

	gqlTodos, err := r.Todos(ctx, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to delete all todos in todo resolver, error when trying to get all todos %s", err.Error())
		return nil, err
	}

	url := r.restUrl + gql_constants.TODO_PATH
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to delete all todos in todo resolver, error when trying to make http request %s", err.Error())
		return r.tConverter.ManyToDeleteTodoPayload(gqlTodos.Data, false), err
	}

	req, err = r.authSetter.DecorateRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorf("failed to decorate http request's auth header, error %s", err.Error())
		return nil, &gqlerror.Error{
			Message: err.Error(),
		}
	}

	resp, err := r.client.Do(req)
	if err != nil {
		log.C(ctx).Errorf("failed to delete all todos in todo resolver, error when trying to get response %s", err.Error())
		return r.tConverter.ManyToDeleteTodoPayload(gqlTodos.Data, false), err
	}
	defer resp.Body.Close()

	return r.tConverter.ManyToDeleteTodoPayload(gqlTodos.Data, true), nil
}

func (r *resolver) CreateTodo(ctx context.Context, input gql.CreateTodoInput) (*gql.Todo, error) {
	log.C(ctx).Info("creating todo in todo resolver")

	restModelTodo := r.tConverter.CreateTodoInputToModel(&input)

	url := r.restUrl + gql_constants.TODO_PATH

	jsonBody, err := json.Marshal(restModelTodo)
	if err != nil {
		log.C(ctx).Errorf("failed to JSON marshal model Todo %s", err.Error())
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		log.C(ctx).Errorf("failed to create todo, error when trying to make http request %s", err.Error())
		return nil, err
	}

	req, err = r.authSetter.DecorateRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorf("failed to decorate http request's auth header, error %s", err.Error())
		return nil, &gqlerror.Error{
			Message: "unauthorized user",
		}
	}

	resp, err := r.client.Do(req)
	if err != nil {
		log.C(ctx).Errorf("failed to create todo, error when trying to get http response %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	modelTodo, err := utils.HandleHttpCode[*models.Todo](resp)
	if err != nil {
		log.C(ctx).Errorf("failed to decode JSON body %s", err.Error())
		return nil, err
	}

	return r.tConverter.ToGQL(modelTodo), nil
}

func (r *resolver) AssignedTo(ctx context.Context, obj *gql.Todo) (*gql.User, error) {
	log.C(ctx).Info("getting todo assignee in todo resolver")

	formattedSuffix := fmt.Sprintf("/%s/assignee", obj.ID)
	url := r.restUrl + gql_constants.TODO_PATH + formattedSuffix

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get todo's assignee, error when trying to make http request %s", err.Error())
		return nil, err
	}

	req, err = r.authSetter.DecorateRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorf("failed to decorate http request's auth header, error %s", err.Error())
		return nil, &gqlerror.Error{
			Message: "unauthorized user",
		}
	}

	resp, err := r.client.Do(req)
	if err != nil {
		log.C(ctx).Errorf("failed to get todo's assignee, error when trying to get http response %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	modelUser, err := utils.HandleHttpCode[*models.User](resp)
	if err != nil {
		log.C(ctx).Error("error when trying to decode assignee from http response body")
		return nil, err
	}

	return r.uConverter.ToGQL(modelUser), nil
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		log.C(ctx).Errorf("failed to get todo's list, error when making http request %s", err.Error())
		return nil, err
	}

	req, err = r.authSetter.DecorateRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorf("failed to decorate http request's auth header, error %s", err.Error())
		return nil, &gqlerror.Error{
			Message: err.Error(),
		}
	}

	resp, err := r.client.Do(req)

	if err != nil {
		log.C(ctx).Errorf("failed to get todo' list, error when trying to receive http response %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	list, err := utils.HandleHttpCode[*models.List](resp)
	if err != nil {
		log.C(ctx).Errorf("failed to decode http response body %s", err.Error())
		return nil, err
	}

	return r.lConverter.ToGQL(list), nil
}

func (r *resolver) getModelTodo(ctx context.Context, id string) (*models.Todo, error) {
	log.C(ctx).Info("getting model todo")

	formattedSuffix := fmt.Sprintf("/%s", id)
	url := r.restUrl + gql_constants.TODO_PATH + formattedSuffix

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		log.C(ctx).Errorf("failed to get todo with id %s in todo resolver, error when making http request %s", id, err.Error())
		return nil, err
	}

	req, err = r.authSetter.DecorateRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorf("failed to decorate http request's auth header, error %s", err.Error())
		return nil, &gqlerror.Error{
			Message: "unauthorized user",
		}
	}

	resp, err := r.client.Do(req)
	if err != nil {
		log.C(ctx).Errorf("failed to get todo with id %s in todo resolver, error when trying to get response %s", id, err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	todo, err := utils.HandleHttpCode[*models.Todo](resp)
	if err != nil {
		log.C(ctx).Error("failed to get todo with id %s in todo resolver, error when trying to decode JSON body")
		return nil, err
	}

	if reflect.ValueOf(todo).IsZero() {
		log.C(ctx).Info("http status code not found received when calling rest api, empty struct...")
		return nil, nil
	}

	return todo, nil
}

func (r *resolver) getTodosByListId(ctx context.Context, listId string) ([]*gql.Todo, error) {
	log.C(ctx).Infof("getting todos by list_id %s", listId)

	formattedSuffix := fmt.Sprintf("/%s%s", listId, gql_constants.TODO_PATH)
	url := r.restUrl + gql_constants.LISTS_PATH + formattedSuffix

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get todos by list_id %s, error when making http request %s", listId, err.Error())
		return nil, err
	}

	req, err = r.authSetter.DecorateRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorf("failed to decorate http request's auth header, error %s", err.Error())
		return nil, &gqlerror.Error{
			Message: err.Error(),
		}
	}

	resp, err := r.client.Do(req)
	if err != nil {
		log.C(ctx).Errorf("failed to get todos by list_id %s, error when trying to get http response %s", listId, err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	modelTodos, err := utils.HandleHttpCode[[]*models.Todo](resp)
	if err != nil {
		log.C(ctx).Errorf("failed to get todos by list_id %s, error when trying to decode JSON", listId)
		return nil, err
	}

	if reflect.ValueOf(modelTodos).IsZero() {
		log.C(ctx).Info("http not found status code received, empty struct...")
		return nil, nil
	}

	return r.tConverter.ManyToGQL(modelTodos), nil
}
