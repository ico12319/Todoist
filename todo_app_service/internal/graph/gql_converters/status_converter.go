package gql_converters

import (
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/gql_constants"
	gql "github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/model"
)

type statusConverter struct{}

func NewStatusConverter() *statusConverter {
	return &statusConverter{}
}

func (*statusConverter) ToStringStatus(status *gql.TodoStatus) string {
	ptrValue := *status

	if ptrValue == gql.TodoStatusOpen {
		return gql_constants.OPEN_LOWERCASE
	} else if ptrValue == gql.TodoStatusInProgress {
		return gql_constants.IN_PROGRESS_LOWERCASE
	}

	return gql_constants.DONE_LOWERCASE
}
