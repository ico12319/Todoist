package gql_converters

import (
	gql "Todo-List/internProject/graphQL_service/graph/model"
	"Todo-List/internProject/todo_app_service/pkg/constants"
)

type overdueConverter struct{}

func NewOverdueConverter() *overdueConverter {
	return &overdueConverter{}
}

func (*overdueConverter) ToStringType(overdue *gql.TodoType) string {
	if overdue == nil {
		return ""
	}
	overdueValue := *overdue

	if overdueValue == constants.EXPIRED {
		return constants.TRUE_VALUE
	}

	return constants.FALSE_VALUE
}
