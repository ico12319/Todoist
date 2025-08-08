package user

import (
	gql "Todo-List/internProject/graphQL_service/graph/model"
	"Todo-List/internProject/graphQL_service/graph/utils"
	"Todo-List/internProject/graphQL_service/internal/gql_constants"
	"Todo-List/internProject/graphQL_service/internal/url_decorators"
	"Todo-List/internProject/graphQL_service/internal/url_decorators/url_filters"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

//go:generate mockery --name=httpService --exported --output=./mocks --outpkg=mocks --filename=http_service.go --with-expecter=true
type httpService interface {
	GetHttpResponseWithAuthHeader(ctx context.Context, httpMethod string, url string, body io.Reader) (*http.Response, error)
}

//go:generate mockery --name=urlDecoratorFactory --exported --output=./mocks --outpkg=mocks --filename=url_decorator_factory.go --with-expecter=true
type urlDecoratorFactory interface {
	CreateUrlDecorator(ctx context.Context, serverAddress string, uFilters url_decorators.UrlFilters) url_decorators.QueryParamsRetrievers
}

//go:generate mockery --name=userConverter --exported --output=./mocks --outpkg=mocks --filename=user_converter.go --with-expecter=true
type userConverter interface {
	ToGQL(user *models.User) *gql.User
	ToUserPageGQL(userPage *models.UserPage) *gql.UserPage
	FromGQLToDeleteUserPayload(user *gql.User, success bool) *gql.DeleteUserPayload
	ManyFromGQLToDeleteUserPayload(users []*gql.User, success bool) []*gql.DeleteUserPayload
}

//go:generate mockery --name=todoConverter --exported --output=./mocks --outpkg=mocks --filename=todo_converter.go --with-expecter=true
type todoConverter interface {
	ToTodoPageGQL(todoPage *models.TodoPage) *gql.TodoPage
}

//go:generate mockery --name=listConverter --exported --output=./mocks --outpkg=mocks --filename=list_converter.go --with-expecter=true
type listConverter interface {
	ToListPageGQL(listPage *models.ListPage) *gql.ListPage
}

type resolver struct {
	uConverter  userConverter
	lConverter  listConverter
	tConverter  todoConverter
	factory     urlDecoratorFactory
	restUrl     string
	httpService httpService
}

func NewResolver(uConverter userConverter, lConverter listConverter, tConverter todoConverter, restUrl string, factory urlDecoratorFactory, httpService httpService) *resolver {
	return &resolver{
		uConverter:  uConverter,
		lConverter:  lConverter,
		tConverter:  tConverter,
		restUrl:     restUrl,
		factory:     factory,
		httpService: httpService,
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

	log.C(ctx).Infof("url is %s", url)
	resp, err := r.httpService.GetHttpResponseWithAuthHeader(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response in user resolver, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if err = utils.HandleHttpCode(resp.StatusCode); err != nil {
		log.C(ctx).Errorf("failed to get list where todo is in todo resolver, error %s due to bad response status code", err.Error())
		return nil, err
	}

	var userPage models.UserPage
	if err = json.NewDecoder(resp.Body).Decode(&userPage); err != nil {
		log.C(ctx).Errorf("failed to decode http response body, error %s", err.Error())
		return nil, err
	}

	return r.uConverter.ToUserPageGQL(&userPage), nil
}

func (r *resolver) User(ctx context.Context, id string) (*gql.User, error) {
	log.C(ctx).Infof("getting user with id %s in user resolver", id)

	formattedSuffix := fmt.Sprintf("/%s", id)
	url := r.restUrl + gql_constants.USER_PATH + formattedSuffix

	resp, err := r.httpService.GetHttpResponseWithAuthHeader(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response in user resolver, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		log.C(ctx).Debugf("http status not found received...")
		return nil, nil
	}

	if err = utils.HandleHttpCode(resp.StatusCode); err != nil {
		log.C(ctx).Errorf("failed to get user with id %s user resolver, error %s due to bad response status code", id, err.Error())
		return nil, err
	}

	var user models.User
	if err = json.NewDecoder(resp.Body).Decode(&user); err != nil {
		log.C(ctx).Errorf("failed to decode http response body, error %s", err.Error())
		return nil, err
	}

	return r.uConverter.ToGQL(&user), nil
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

	resp, err := r.httpService.GetHttpResponseWithAuthHeader(ctx, http.MethodDelete, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response in user resolver, error %s", err.Error())
		return &gql.DeleteUserPayload{
			Success: false,
		}, err
	}
	defer resp.Body.Close()

	if err = utils.HandleHttpCode(resp.StatusCode); err != nil {
		log.C(ctx).Errorf("failed to delete user with id %s in user resolver, error %s", id, err.Error())
		return &gql.DeleteUserPayload{
			Success: false,
		}, err
	}

	return r.uConverter.FromGQLToDeleteUserPayload(gqlUser, true), nil
}

func (r *resolver) AssignedTo(ctx context.Context, obj *gql.User, filters *url_filters.TodoFilters) (*gql.TodoPage, error) {
	log.C(ctx).Infof("getting todo assigned to user with id %s", obj.ID)

	formattedSuffix := fmt.Sprintf("/%s", obj.ID)

	decorator := r.factory.CreateUrlDecorator(ctx, gql_constants.USER_PATH+formattedSuffix+gql_constants.TODO_PATH, filters)

	url, err := decorator.DetermineCorrectQueryParams(ctx, r.restUrl)
	if err != nil {
		log.C(ctx).Error("failed to get todos assigned to user, error when trying to build url")
		return nil, err
	}

	resp, err := r.httpService.GetHttpResponseWithAuthHeader(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response in user resolver, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		log.C(ctx).Warn("http status code not found when trying to see to which todos user is assigned...")
		return nil, nil
	}

	if err = utils.HandleHttpCode(resp.StatusCode); err != nil {
		log.C(ctx).Errorf("failed to get lists where user participates, error %s", err.Error())
		return nil, err
	}

	var todoPage models.TodoPage
	if err = json.NewDecoder(resp.Body).Decode(&todoPage); err != nil {
		log.C(ctx).Error("failed to get todos to which user is assigned to , error when trying to decode JSON")
		return nil, err
	}

	return r.tConverter.ToTodoPageGQL(&todoPage), nil
}

func (r *resolver) ParticipateIn(ctx context.Context, obj *gql.User, filters *url_filters.BaseFilters) (*gql.ListPage, error) {
	log.C(ctx).Infof("getting lists shared with user with id %s", obj.ID)

	formattedSuffix := fmt.Sprintf("/%s", obj.ID)

	decorator := r.factory.CreateUrlDecorator(ctx, gql_constants.USER_PATH+formattedSuffix+gql_constants.LISTS_PATH, filters)

	url, err := decorator.DetermineCorrectQueryParams(ctx, r.restUrl)
	if err != nil {
		log.C(ctx).Errorf("failed to get lists shared with user with id %s, error when trying to build url", obj.ID)
		return &gql.ListPage{
			Data:       make([]*gql.List, 0),
			PageInfo:   nil,
			TotalCount: 0,
		}, err
	}

	resp, err := r.httpService.GetHttpResponseWithAuthHeader(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response in user resolver, error %s", err.Error())
		return &gql.ListPage{
			Data:       make([]*gql.List, 0),
			PageInfo:   nil,
			TotalCount: 0,
		}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		log.C(ctx).Warn("http status code not found when trying to see in which lists user participates...")
		return &gql.ListPage{
			Data:       make([]*gql.List, 0),
			PageInfo:   nil,
			TotalCount: 0,
		}, nil
	}

	if err = utils.HandleHttpCode(resp.StatusCode); err != nil {
		log.C(ctx).Errorf("failed to get lists where user participates, error %s", err.Error())
		return &gql.ListPage{
			Data:       make([]*gql.List, 0),
			PageInfo:   nil,
			TotalCount: 0,
		}, err
	}

	var listPage models.ListPage
	if err = json.NewDecoder(resp.Body).Decode(&listPage); err != nil {
		log.C(ctx).Errorf("failed to get lists shared with user with id %s, error when trying to decode JSON", obj.ID)
		return &gql.ListPage{
			Data:       make([]*gql.List, 0),
			PageInfo:   nil,
			TotalCount: 0,
		}, err
	}

	return r.lConverter.ToListPageGQL(&listPage), nil
}

func (r *resolver) DeleteUsers(ctx context.Context) ([]*gql.DeleteUserPayload, error) {
	log.C(ctx).Info("deleting all users in user resolver")

	gqlUsers, err := r.getUsers(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to delete users, error %s when trying to get them", err.Error())
		return []*gql.DeleteUserPayload{
			{
				Success: false,
			},
		}, err
	}

	url := r.restUrl + gql_constants.USER_PATH

	resp, err := r.httpService.GetHttpResponseWithAuthHeader(ctx, http.MethodDelete, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response in user resolver, error %s", err.Error())
		return []*gql.DeleteUserPayload{
			{
				Success: false,
			},
		}, err
	}
	defer resp.Body.Close()

	return r.uConverter.ManyFromGQLToDeleteUserPayload(gqlUsers.Data, true), nil
}

func (r *resolver) getUsers(ctx context.Context) (*gql.UserPage, error) {
	log.C(ctx).Info("getting users without filters in user resolver")

	url := r.restUrl + gql_constants.USER_PATH

	resp, err := r.httpService.GetHttpResponseWithAuthHeader(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response in user resolver, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if err = utils.HandleHttpCode(resp.StatusCode); err != nil {
		log.C(ctx).Errorf("failed to get users in user resolver, error %s due to bad response status code", err.Error())
		return nil, err
	}

	var userPage models.UserPage
	if err = json.NewDecoder(resp.Body).Decode(&userPage); err != nil {
		log.C(ctx).Errorf("failed to decode http response body, error %s", err.Error())
		return nil, err
	}

	return r.uConverter.ToUserPageGQL(&userPage), nil
}
