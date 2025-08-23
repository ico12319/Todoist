package gql_converters

import (
	gql "Todo-List/internProject/graphQL_service/graph/model"
	"Todo-List/internProject/graphQL_service/internal/gql_constants"
)

type priorityConverter struct{}

func NewPriorityConverter() *priorityConverter {
	return &priorityConverter{}
}

func (*priorityConverter) ToStringPriority(priority *gql.Priority) string {
	if priority == nil {
		return ""
	}

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
