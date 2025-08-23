package gql_converters

import (
	gql "Todo-List/internProject/graphQL_service/graph/model"
	"Todo-List/internProject/graphQL_service/internal/gql_constants"
)

type statusConverter struct{}

func NewStatusConverter() *statusConverter {
	return &statusConverter{}
}

func (*statusConverter) ToStringStatus(status *gql.TodoStatus) string {
	if status == nil {
		return ""
	}

	ptrValue := *status

	if ptrValue == gql.TodoStatusOpen {
		return gql_constants.OPEN_LOWERCASE
	} else if ptrValue == gql.TodoStatusInProgress {
		return gql_constants.IN_PROGRESS_LOWERCASE
	}

	return gql_constants.DONE_LOWERCASE
}
