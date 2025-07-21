package utils

import (
	gql "Todo-List/internProject/graphQL_service/graph/model"
	"Todo-List/internProject/todo_app_service/pkg/handler_models"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"encoding/json"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"net/http"
)

type todoConverter interface {
	ToGQL(todo *models.Todo) *gql.Todo
	ManyToGQL(todos []*models.Todo) []*gql.Todo
	ToHandlerModel(todoInput *gql.UpdateTodoInput) *handler_models.UpdateTodo
	CreateTodoInputToModel(todoInput *gql.CreateTodoInput) *handler_models.CreateTodo
	FromGQLModelToDeleteTodoPayload(todo *gql.Todo, success bool) *gql.DeleteTodoPayload
}

func DecodeJsonResponse[T any](resp *http.Response) (T, error) {
	var decodeModel T
	if err := json.NewDecoder(resp.Body).Decode(&decodeModel); err != nil {
		return decodeModel, err
	}

	return decodeModel, nil
}

func InitPageInfo[T *gql.Todo | *gql.List | *gql.User](gqlModels []T, extractIdFunc func(T) string) *gql.PageInfo {
	if len(gqlModels) == 0 {
		return nil
	}

	pageInfo := &gql.PageInfo{
		StartCursor: extractIdFunc(gqlModels[0]),
		EndCursor:   extractIdFunc(gqlModels[len(gqlModels)-1]),
	}
	return pageInfo
}

func HandleHttpCode(statusCode int) error {

	if statusCode == http.StatusInternalServerError {
		return &gqlerror.Error{
			Message:    "Internal error, please try again later.",
			Extensions: map[string]interface{}{"code": "INTERNAL_SERVER_ERROR"},
		}
	} else if statusCode == http.StatusNotFound {
		return nil
	} else if statusCode == http.StatusBadRequest {
		return &gqlerror.Error{
			Message:    "Invalid Request",
			Extensions: map[string]interface{}{"code": "BAD_REQUEST"},
		}
	} else if statusCode == http.StatusForbidden {
		return &gqlerror.Error{
			Message:    "Don't have permission to perform this action",
			Extensions: map[string]interface{}{"code": "FORBIDDEN"},
		}
	} else if statusCode == http.StatusUnauthorized {
		return &gqlerror.Error{
			Message:    "Unauthorized user",
			Extensions: map[string]interface{}{"code": "Unauthorized"},
		}
	}

	return nil
}

func InitEmptyListPage() *gql.ListPage {
	return &gql.ListPage{
		Data:       make([]*gql.List, 0),
		PageInfo:   nil,
		TotalCount: 0,
	}
}

func InitEmptyTodoPage() *gql.TodoPage {
	return &gql.TodoPage{
		Data:       make([]*gql.Todo, 0),
		PageInfo:   nil,
		TotalCount: 0,
	}
}
