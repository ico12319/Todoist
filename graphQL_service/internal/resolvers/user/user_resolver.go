package user

import (
	gql "Todo-List/internProject/graphQL_service/graph/model"
	"Todo-List/internProject/graphQL_service/graph/utils"
	"Todo-List/internProject/graphQL_service/internal/gql_constants"
	"Todo-List/internProject/graphQL_service/internal/url_decorators"
	"Todo-List/internProject/graphQL_service/internal/url_decorators/url_filters"
	"Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"context"
	"fmt"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"io"
	"net/http"
	"reflect"
)

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type urlDecoratorFactory interface {
	CreateUrlDecorator(context.Context, string, url_decorators.UrlFilters) url_decorators.QueryParamsRetrievers
}

type requestAuthSetter interface {
	DecorateRequest(context.Context, *http.Request) (*http.Request, error)
}

type userConverter interface {
	ToGQL(*models.User) *gql.User
	ManyToGQL([]*models.User) []*gql.User
	FromGQLToDeleteUserPayload(*gql.User, bool) *gql.DeleteUserPayload
	ManyFromGQLToDeleteUserPayload([]*gql.User, bool) []*gql.DeleteUserPayload
}

//go:generate mockery --name=HttpRequester --output=./mocks --outpkg=mocks --filename=http_requester.go --with-expecter=true
type httpRequester interface {
	NewRequestWithContext(context.Context, string, string, io.Reader) (*http.Request, error)
}

type todoConverter interface {
	ManyToGQL([]*models.Todo) []*gql.Todo
}

type listConverter interface {
	ManyToGQL([]*models.List) []*gql.List
}

type resolver struct {
	client        httpClient
	uConverter    userConverter
	lConverter    listConverter
	tConverter    todoConverter
	factory       urlDecoratorFactory
	restUrl       string
	authSetter    requestAuthSetter
	httpRequester httpRequester
}

func NewResolver(client httpClient, uConverter userConverter, lConverter listConverter, tConverter todoConverter, restUrl string, factory urlDecoratorFactory, authSetter requestAuthSetter, httpRequester httpRequester) *resolver {
	return &resolver{
		client:        client,
		uConverter:    uConverter,
		lConverter:    lConverter,
		tConverter:    tConverter,
		restUrl:       restUrl,
		factory:       factory,
		authSetter:    authSetter,
		httpRequester: httpRequester,
	}
}

func (r *resolver) Users(ctx context.Context, filters *url_filters.BaseFilters) (*gql.UserPage, error) {
	log.C(ctx).Info("getting users in user resolver")
	decorator := r.factory.CreateUrlDecorator(ctx, gql_constants.USER_PATH, filters)

	url, err := decorator.DetermineCorrectQueryParams(ctx, r.restUrl)
	if err != nil {
		log.C(ctx).Errorf("failed to get users, error when calling factory function")
		return nil, err
	}

	req, err := r.httpRequester.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get users, error when making http request")
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
		log.C(ctx).Error("failed to get users, error when trying to get http response")
		return nil, err
	}
	defer resp.Body.Close()

	modelUsers, err := utils.HandleHttpCode[[]*models.User](resp)
	if err != nil {
		log.C(ctx).Error("failed to get users, error when decoding JSON response")
		return nil, err
	}

	gqlUsers := r.uConverter.ManyToGQL(modelUsers)

	pageInfo := utils.InitPageInfo[*gql.User](gqlUsers, func(user *gql.User) string {
		return user.ID
	})

	return &gql.UserPage{
		Data:       gqlUsers,
		PageInfo:   pageInfo,
		TotalCount: int32(len(gqlUsers)),
	}, nil
}

func (r *resolver) User(ctx context.Context, id string) (*gql.User, error) {
	log.C(ctx).Infof("getting user with id %s in user resolver", id)

	formattedSuffix := fmt.Sprintf("/%s", id)
	url := r.restUrl + gql_constants.USER_PATH + formattedSuffix

	req, err := r.httpRequester.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get user, error when making http request")
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
		log.C(ctx).Error("failed to get user, error when trying to get http response")
		return nil, err
	}
	defer resp.Body.Close()

	modelUser, err := utils.HandleHttpCode[*models.User](resp)
	if err != nil {
		log.C(ctx).Error("failed to get user, error when decoding JSON response")
		return nil, err
	}

	if reflect.ValueOf(modelUser).IsZero() {
		log.C(ctx).Info("http status not found received when calling rest api, empty struct...")
		return nil, nil
	}

	return r.uConverter.ToGQL(modelUser), nil
}

func (r *resolver) DeleteUser(ctx context.Context, id string) (*gql.DeleteUserPayload, error) {
	log.C(ctx).Infof("deleting user with id %s in user resolver", id)

	gqlUser, err := r.User(ctx, id)
	if err != nil {
		log.C(ctx).Errorf("failed to delete user, error when trying to get user with id %s", id)
		return &gql.DeleteUserPayload{
			Success: false,
		}, err
	}

	if gqlUser == nil {
		log.C(ctx).Infof("user with id %s not found", id)
		return &gql.DeleteUserPayload{
			Success: false,
		}, nil
	}

	formattedSuffix := fmt.Sprintf("/%s", id)
	url := r.restUrl + gql_constants.USER_PATH + formattedSuffix

	req, err := r.httpRequester.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to delete user with id %s, error when making http request", id)
		return r.uConverter.FromGQLToDeleteUserPayload(gqlUser, false), nil
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
		log.C(ctx).Errorf("failed to delete user with id %s, error when trying to get http response", id)
		return r.uConverter.FromGQLToDeleteUserPayload(gqlUser, false), nil
	}
	defer resp.Body.Close()

	return r.uConverter.FromGQLToDeleteUserPayload(gqlUser, true), nil
}

func (r *resolver) AssignedTo(ctx context.Context, obj *gql.User, baseFilters *url_filters.BaseFilters) (*gql.TodoPage, error) {
	log.C(ctx).Infof("getting todo assigned to user with id %s", obj.ID)

	formattedSuffix := fmt.Sprintf("/%s", obj.ID)

	decorator := r.factory.CreateUrlDecorator(ctx, gql_constants.USER_PATH+formattedSuffix+gql_constants.TODO_PATH, baseFilters)

	url, err := decorator.DetermineCorrectQueryParams(ctx, r.restUrl)
	if err != nil {
		log.C(ctx).Error("failed to get todos assigned to user, error when trying to build url")
		return nil, err
	}

	req, err := r.httpRequester.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Error("failed to get todos assigned to user, error when trying to make http request")
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
		log.C(ctx).Error("failed to get todos assigned to user, error trying to get http response")
		return nil, err
	}
	defer resp.Body.Close()

	modelTodos, err := utils.DecodeJsonResponse[[]*models.Todo](resp)
	if err != nil {
		log.C(ctx).Error("failed to decode http response body")
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

func (r *resolver) ParticipateIn(ctx context.Context, obj *gql.User, filters *url_filters.BaseFilters) (*gql.ListPage, error) {
	log.C(ctx).Infof("getting lists shared with user with id %s", obj.ID)

	formattedSuffix := fmt.Sprintf("/%s", obj.ID)

	decorator := r.factory.CreateUrlDecorator(ctx, gql_constants.USER_PATH+formattedSuffix+gql_constants.LISTS_PATH, filters)

	url, err := decorator.DetermineCorrectQueryParams(ctx, r.restUrl)
	if err != nil {
		log.C(ctx).Errorf("failed to get lists shared with user with id %s, error when trying to build url", obj.ID)
		return nil, err
	}

	req, err := r.httpRequester.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get lists shared with user with id %s, error when trying to make http request", obj.ID)
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
		log.C(ctx).Errorf("failed to get lists shared with user with id %s, error when trying to get response", obj.ID)
		return nil, err
	}
	defer resp.Body.Close()

	lists, err := utils.DecodeJsonResponse[[]*models.List](resp)
	if err != nil {
		log.C(ctx).Errorf("failed to get lists shared with user with id %s, error when trying to decode JSON", obj.ID)
		return nil, err
	}

	gqlLists := r.lConverter.ManyToGQL(lists)

	pageInfo := utils.InitPageInfo[*gql.List](gqlLists, func(list *gql.List) string {
		return list.ID
	})

	return &gql.ListPage{
		Data:       gqlLists,
		PageInfo:   pageInfo,
		TotalCount: int32(len(gqlLists)),
	}, nil
}

func (r *resolver) DeleteUsers(ctx context.Context) ([]*gql.DeleteUserPayload, error) {
	log.C(ctx).Info("deleting all users in user resolver")

	gqlUsers, err := r.Users(ctx, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to delete users, error %s when trying to get them", err.Error())
		return nil, err
	}

	url := r.restUrl + gql_constants.USER_PATH

	req, err := r.httpRequester.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to delete all users, error %s when trying to make http request", err.Error())
		return nil, err
	}

	req, err = r.authSetter.DecorateRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorf("failed to decorate http request, error %s", err.Error())
		return nil, &gqlerror.Error{
			Message: err.Error(),
		}
	}

	resp, err := r.client.Do(req)
	if err != nil {
		log.C(ctx).Errorf("failed to delete all users, error %s when trying to get http respose", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	return r.uConverter.ManyFromGQLToDeleteUserPayload(gqlUsers.Data, true), nil
}
