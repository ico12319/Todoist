package utils

import (
	gql "Todo-List/internProject/graphQL_service/graph/model"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"net/http"
)

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
			Extensions: map[string]interface{}{"code": "UNAUTHORIZED"},
		}
	}

	return nil
}

func InitEmptyTodoPage() *gql.TodoPage {
	return &gql.TodoPage{
		Data:       make([]*gql.Todo, 0),
		PageInfo:   nil,
		TotalCount: 0,
	}
}
