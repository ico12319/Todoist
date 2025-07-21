package graph

import (
	gql "Todo-List/internProject/graphQL_service/graph/model"
	"Todo-List/internProject/graphQL_service/internal/url_decorators/url_filters"
	"context"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type lResolver interface {
	Lists(ctx context.Context, filter *url_filters.BaseFilters) (*gql.ListPage, error)
	List(ctx context.Context, id string) (*gql.List, error)
	ListOwner(ctx context.Context, obj *gql.List) (*gql.User, error)
	Todos(ctx context.Context, obj *gql.List, filters *url_filters.TodoFilters) (*gql.TodoPage, error)
	DeleteList(ctx context.Context, id string) (*gql.DeleteListPayload, error)
	DeleteLists(ctx context.Context) ([]*gql.DeleteListPayload, error)
	UpdateList(ctx context.Context, id string, input gql.UpdateListInput) (*gql.List, error)
	AddListCollaborator(ctx context.Context, input gql.CollaboratorInput) (*gql.CreateCollaboratorPayload, error)
	DeleteListCollaborator(ctx context.Context, id string, userID string) (*gql.DeleteCollaboratorPayload, error)
	Collaborators(ctx context.Context, obj *gql.List, filters *url_filters.BaseFilters) (*gql.UserPage, error)
	CreateList(ctx context.Context, input gql.CreateListInput) (*gql.List, error)
}

type tResolver interface {
	Todos(ctx context.Context, filter *url_filters.TodoFilters) (*gql.TodoPage, error)
	Todo(ctx context.Context, id string) (*gql.Todo, error)
	DeleteTodosByListID(ctx context.Context, id string) ([]*gql.DeleteTodoPayload, error)
	DeleteTodos(ctx context.Context) ([]*gql.DeleteTodoPayload, error)
	DeleteTodo(ctx context.Context, id string) (*gql.DeleteTodoPayload, error)
	CreateTodo(ctx context.Context, input gql.CreateTodoInput) (*gql.Todo, error)
	UpdateTodo(ctx context.Context, id string, input gql.UpdateTodoInput) (*gql.Todo, error)
	AssignedTo(ctx context.Context, obj *gql.Todo) (*gql.User, error)
	List(ctx context.Context, obj *gql.Todo) (*gql.List, error)
}

type uResolver interface {
	Users(ctx context.Context, filters *url_filters.BaseFilters) (*gql.UserPage, error)
	User(ctx context.Context, id string) (*gql.User, error)
	DeleteUser(ctx context.Context, id string) (*gql.DeleteUserPayload, error)
	DeleteUsers(ctx context.Context) ([]*gql.DeleteUserPayload, error)
	AssignedTo(ctx context.Context, obj *gql.User, baseFilters *url_filters.TodoFilters) (*gql.TodoPage, error)
	ParticipateIn(ctx context.Context, obj *gql.User, filters *url_filters.UserFilters) (*gql.ListPage, error)
}

type aResolver interface {
	ExchangeRefreshToken(ctx context.Context, input gql.RefreshTokenInput) (*gql.Access, error)
}

type activityResolver interface {
	RandomActivity(ctx context.Context) (*gql.RandomActivity, error)
}

type Resolver struct {
	lResolver        lResolver
	tResolver        tResolver
	uResolver        uResolver
	aResolver        aResolver
	activityResolver activityResolver
}

func NewResolver(lResolver lResolver, tResolver tResolver, uResolver uResolver, aResolver aResolver, activityResolver activityResolver) *Resolver {
	return &Resolver{
		lResolver:        lResolver,
		tResolver:        tResolver,
		uResolver:        uResolver,
		aResolver:        aResolver,
		activityResolver: activityResolver,
	}
}
