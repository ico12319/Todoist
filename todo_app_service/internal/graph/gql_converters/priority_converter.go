package gql_converters

import (
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/gql_constants"
	gql "github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/model"
)

type priorityConverter struct{}

func NewPriorityConverter() *priorityConverter {
	return &priorityConverter{}
}

func (*priorityConverter) ToStringPriority(priority *gql.Priority) string {
	ptrValue := *priority

	if ptrValue == gql.PriorityVeryLow {
		return gql_constants.VERY_LOW_PRIORITY_LOWERCASE
	} else if ptrValue == gql.PriorityLow {
		return gql_constants.LOW_PRIORITY_LOWERCASE
	} else if ptrValue == gql.PriorityMedium {
		return gql_constants.MEDIUM_PRIORITY_LOWERCASE
	} else if ptrValue == gql.PriorityHigh {
		return gql_constants.HIGH_PRIORITY_LOWERCASE
	}

	return gql_constants.VERY_HIGH_PRIORITY_LOWERCASE
}
