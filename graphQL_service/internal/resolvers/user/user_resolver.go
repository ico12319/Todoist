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

type httpResponseGetter interface {
	GetHttpResponseWithAuthHeader(context.Context, string, string, io.Reader) (*http.Response, error)
}

type urlDecoratorFactory interface {
	CreateUrlDecorator(context.Context, string, url_decorators.UrlFilters) url_decorators.QueryParamsRetrievers
}

type userConverter interface {
	ToGQL(*models.User) *gql.User
	ManyToGQL([]*models.User) []*gql.User
	FromGQLToDeleteUserPayload(*gql.User, bool) *gql.DeleteUserPayload
	ManyFromGQLToDeleteUserPayload([]*gql.User, bool) []*gql.DeleteUserPayload
}

type todoConverter interface {
	ManyToGQL([]*models.Todo) []*gql.Todo
}

type listConverter interface {
	ManyToGQL([]*models.List) []*gql.List
}

type resolver struct {
	uConverter     userConverter
	lConverter     listConverter
	tConverter     todoConverter
	factory        urlDecoratorFactory
	restUrl        string
	responseGetter httpResponseGetter
}

func NewResolver(uConverter userConverter, lConverter listConverter, tConverter todoConverter, restUrl string, factory urlDecoratorFactory, responseGetter httpResponseGetter) *resolver {
	return &resolver{
		uConverter:     uConverter,
		lConverter:     lConverter,
		tConverter:     tConverter,
		restUrl:        restUrl,
		factory:        factory,
		responseGetter: responseGetter,
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

	resp, err := r.responseGetter.GetHttpResponseWithAuthHeader(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response in user resolver, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if err = utils.HandleHttpCode(resp.StatusCode); err != nil {
		log.C(ctx).Errorf("failed to get list where todo is in todo resolver, error %s due to bad response status code", err.Error())
		return nil, err
	}

	var users []*models.User
	if err = json.NewDecoder(resp.Body).Decode(&users); err != nil {
		log.C(ctx).Errorf("failed to decode http response body, error %s", err.Error())
		return nil, err
	}

	gqlUsers := r.uConverter.ManyToGQL(users)

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

	resp, err := r.responseGetter.GetHttpResponseWithAuthHeader(ctx, http.MethodGet, url, nil)
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

	resp, err := r.responseGetter.GetHttpResponseWithAuthHeader(ctx, http.MethodDelete, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response in user resolver, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	return r.uConverter.FromGQLToDeleteUserPayload(gqlUser, true), nil
}

func (r *resolver) AssignedTo(ctx context.Context, obj *gql.User, baseFilters *url_filters.TodoFilters) (*gql.TodoPage, error) {
	log.C(ctx).Infof("getting todo assigned to user with id %s", obj.ID)

	formattedSuffix := fmt.Sprintf("/%s", obj.ID)

	decorator := r.factory.CreateUrlDecorator(ctx, gql_constants.USER_PATH+formattedSuffix+gql_constants.TODO_PATH, baseFilters)

	url, err := decorator.DetermineCorrectQueryParams(ctx, r.restUrl)
	if err != nil {
		log.C(ctx).Error("failed to get todos assigned to user, error when trying to build url")
		return nil, err
	}

	resp, err := r.responseGetter.GetHttpResponseWithAuthHeader(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response in user resolver, error %s", err.Error())
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

func (r *resolver) ParticipateIn(ctx context.Context, obj *gql.User, filters *url_filters.UserFilters) (*gql.ListPage, error) {
	log.C(ctx).Infof("getting lists shared with user with id %s", obj.ID)

	formattedSuffix := fmt.Sprintf("/%s", obj.ID)

	decorator := r.factory.CreateUrlDecorator(ctx, gql_constants.USER_PATH+formattedSuffix+gql_constants.LISTS_PATH, filters)

	url, err := decorator.DetermineCorrectQueryParams(ctx, r.restUrl)
	if err != nil {
		log.C(ctx).Errorf("failed to get lists shared with user with id %s, error when trying to build url", obj.ID)
		return nil, err
	}

	resp, err := r.responseGetter.GetHttpResponseWithAuthHeader(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response in user resolver, error %s", err.Error())
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

	gqlUsers, err := r.getUsers(ctx)
	if err != nil {
		log.C(ctx).Errorf("failed to delete users, error %s when trying to get them", err.Error())
		return nil, err
	}

	url := r.restUrl + gql_constants.USER_PATH

	resp, err := r.responseGetter.GetHttpResponseWithAuthHeader(ctx, http.MethodDelete, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response in user resolver, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	return r.uConverter.ManyFromGQLToDeleteUserPayload(gqlUsers, true), nil
}

func (r *resolver) getUsers(ctx context.Context) ([]*gql.User, error) {
	log.C(ctx).Info("getting users without filters in user resolver")

	url := r.restUrl + gql_constants.USER_PATH

	resp, err := r.responseGetter.GetHttpResponseWithAuthHeader(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.C(ctx).Errorf("failed to get http response in user resolver, error %s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if err = utils.HandleHttpCode(resp.StatusCode); err != nil {
		log.C(ctx).Errorf("failed to get users in user resolver, error %s due to bad response status code", err.Error())
		return nil, err
	}

	var users []*models.User
	if err = json.NewDecoder(resp.Body).Decode(&users); err != nil {
		log.C(ctx).Errorf("failed to decode http response body, error %s", err.Error())
		return nil, err
	}

	return r.uConverter.ManyToGQL(users), nil
}
