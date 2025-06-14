package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.73

import (
	"context"
	gql "github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/model"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/url_decorators/url_filters"
)

// Owner is the resolver for the owner field.
func (r *listResolver) Owner(ctx context.Context, obj *gql.List) (*gql.User, error) {
	return r.lResolver.ListOwner(ctx, obj)
}

// Todos is the resolver for the todos field.
func (r *listResolver) Todos(ctx context.Context, obj *gql.List, limit *int32, after *string, criteria *gql.TodosFilterInput) (*gql.TodoPage, error) {
	filters := &url_filters.TodoFilters{
		BaseFilters: url_filters.BaseFilters{
			Limit:  limit,
			Cursor: after,
		},
		TodoFilters: criteria,
	}

	return r.lResolver.Todos(ctx, obj, filters)
}

// Collaborators is the resolver for the collaborators field.
func (r *listResolver) Collaborators(ctx context.Context, obj *gql.List, limit *int32, after *string) (*gql.UserPage, error) {
	filters := &url_filters.BaseFilters{
		Limit:  limit,
		Cursor: after,
	}
	return r.lResolver.Collaborators(ctx, obj, filters)
}

// CreateList is the resolver for the createList field.
func (r *mutationResolver) CreateList(ctx context.Context, input gql.CreateListInput) (*gql.List, error) {
	return r.lResolver.CreateList(ctx, input)
}

// UpdateList is the resolver for the updateList field.
func (r *mutationResolver) UpdateList(ctx context.Context, id string, input gql.UpdateListInput) (*gql.List, error) {
	return r.lResolver.UpdateList(ctx, id, input)
}

// AddListCollaborator is the resolver for the addListCollaborator field.
func (r *mutationResolver) AddListCollaborator(ctx context.Context, input gql.CollaboratorInput) (*gql.CreateCollaboratorPayload, error) {
	return r.lResolver.AddListCollaborator(ctx, input)
}

// DeleteListCollaborator is the resolver for the deleteListCollaborator field.
func (r *mutationResolver) DeleteListCollaborator(ctx context.Context, id string, userID string) (*gql.DeleteCollaboratorPayload, error) {
	return r.lResolver.DeleteListCollaborator(ctx, id, userID)
}

// DeleteList is the resolver for the deleteList field.
func (r *mutationResolver) DeleteList(ctx context.Context, id string) (*gql.DeleteListPayload, error) {
	return r.lResolver.DeleteList(ctx, id)
}

// DeleteLists is the resolver for the deleteLists field.
func (r *mutationResolver) DeleteLists(ctx context.Context) ([]*gql.DeleteListPayload, error) {
	return r.lResolver.DeleteLists(ctx)
}

// CreateTodo is the resolver for the createTodo field.
func (r *mutationResolver) CreateTodo(ctx context.Context, input gql.CreateTodoInput) (*gql.Todo, error) {
	return r.tResolver.CreateTodo(ctx, input)
}

// DeleteTodo is the resolver for the deleteTodo field.
func (r *mutationResolver) DeleteTodo(ctx context.Context, id string) (*gql.DeleteTodoPayload, error) {
	return r.tResolver.DeleteTodo(ctx, id)
}

// DeleteTodos is the resolver for the deleteTodos field.
func (r *mutationResolver) DeleteTodos(ctx context.Context) ([]*gql.DeleteTodoPayload, error) {
	return r.tResolver.DeleteTodos(ctx)
}

// UpdateTodo is the resolver for the updateTodo field.
func (r *mutationResolver) UpdateTodo(ctx context.Context, id string, input gql.UpdateTodoInput) (*gql.Todo, error) {
	return r.tResolver.UpdateTodo(ctx, id, input)
}

// DeleteTodosByListID is the resolver for the deleteTodosByListId field.
func (r *mutationResolver) DeleteTodosByListID(ctx context.Context, id string) ([]*gql.DeleteTodoPayload, error) {
	return r.tResolver.DeleteTodosByListID(ctx, id)
}

// DeleteUser is the resolver for the deleteUser field.
func (r *mutationResolver) DeleteUser(ctx context.Context, id string) (*gql.DeleteUserPayload, error) {
	return r.uResolver.DeleteUser(ctx, id)
}

// DeleteUsers is the resolver for the deleteUsers field.
func (r *mutationResolver) DeleteUsers(ctx context.Context) ([]*gql.DeleteUserPayload, error) {
	return r.uResolver.DeleteUsers(ctx)
}

// Lists is the resolver for the lists field.
func (r *queryResolver) Lists(ctx context.Context, limit *int32, after *string) (*gql.ListPage, error) {
	filters := &url_filters.BaseFilters{
		Limit:  limit,
		Cursor: after,
	}
	return r.lResolver.Lists(ctx, filters)
}

// List is the resolver for the list field.
func (r *queryResolver) List(ctx context.Context, id string) (*gql.List, error) {
	return r.lResolver.List(ctx, id)
}

// Todos is the resolver for the todos field.
func (r *queryResolver) Todos(ctx context.Context, limit *int32, after *string, criteria *gql.TodosFilterInput) (*gql.TodoPage, error) {
	filters := &url_filters.TodoFilters{
		BaseFilters: url_filters.BaseFilters{
			Limit:  limit,
			Cursor: after,
		},
		TodoFilters: criteria,
	}
	return r.tResolver.Todos(ctx, filters)
}

// Todo is the resolver for the todo field.
func (r *queryResolver) Todo(ctx context.Context, id string) (*gql.Todo, error) {
	return r.tResolver.Todo(ctx, id)
}

// Users is the resolver for the users field.
func (r *queryResolver) Users(ctx context.Context, limit *int32, after *string) (*gql.UserPage, error) {
	filters := &url_filters.BaseFilters{
		Limit:  limit,
		Cursor: after,
	}
	return r.uResolver.Users(ctx, filters)
}

// User is the resolver for the user field.
func (r *queryResolver) User(ctx context.Context, id string) (*gql.User, error) {
	return r.uResolver.User(ctx, id)
}

// List is the resolver for the list field.
func (r *todoResolver) List(ctx context.Context, obj *gql.Todo) (*gql.List, error) {
	return r.tResolver.List(ctx, obj)
}

// AssignedTo is the resolver for the assigned_to field.
func (r *todoResolver) AssignedTo(ctx context.Context, obj *gql.Todo) (*gql.User, error) {
	return r.tResolver.AssignedTo(ctx, obj)
}

// AssignedTo is the resolver for the assigned_to field.
func (r *userResolver) AssignedTo(ctx context.Context, obj *gql.User, limit *int32, after *string) (*gql.TodoPage, error) {
	filters := &url_filters.BaseFilters{
		Limit:  limit,
		Cursor: after,
	}
	return r.uResolver.AssignedTo(ctx, obj, filters)
}

// ParticipateIn is the resolver for the participate_in field.
func (r *userResolver) ParticipateIn(ctx context.Context, obj *gql.User, limit *int32, after *string) (*gql.ListPage, error) {
	filters := &url_filters.BaseFilters{
		Limit:  limit,
		Cursor: after,
	}
	return r.uResolver.ParticipateIn(ctx, obj, filters)
}

// List returns ListResolver implementation.
func (r *Resolver) List() ListResolver { return &listResolver{r} }

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

// Todo returns TodoResolver implementation.
func (r *Resolver) Todo() TodoResolver { return &todoResolver{r} }

// User returns UserResolver implementation.
func (r *Resolver) User() UserResolver { return &userResolver{r} }

type listResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type todoResolver struct{ *Resolver }
type userResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
/*
	func (r *mutationResolver) CreateUser(ctx context.Context, input gql.CreateUserInput) (*gql.User, error) {
	return r.uResolver.CreateUser(ctx, input)
}
func (r *mutationResolver) UpdateUser(ctx context.Context, id string, input gql.UpdateUserInput) (*gql.User, error) {
	return r.uResolver.UpdateUser(ctx, id, input)
}
*/
