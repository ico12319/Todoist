package url_filters

import (
	gql "Todo-List/internProject/graphQL_service/graph/model"
	"Todo-List/internProject/graphQL_service/internal/gql_constants"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"bytes"
)

type BaseFilters struct {
	First  *string
	After  *string
	Last   *string
	Before *string
}

func (b *BaseFilters) GetFilters() map[string]*string {
	return map[string]*string{
		gql_constants.FIRST:  b.First,
		gql_constants.AFTER:  b.After,
		gql_constants.LAST:   b.Last,
		gql_constants.BEFORE: b.Before,
	}
}

type TodoFilters struct {
	BaseFilters
	TodoFilters *gql.TodosFilterInput
}

func (t *TodoFilters) GetFilters() map[string]*string {
	pr, st := extractPriorityAndStatus(t.TodoFilters)
	todoType := extractType(t.TodoFilters)

	return map[string]*string{
		gql_constants.FIRST:    t.First,
		gql_constants.AFTER:    t.After,
		gql_constants.LAST:     t.Last,
		gql_constants.BEFORE:   t.Before,
		gql_constants.PRIORITY: pr,
		gql_constants.STATUS:   st,
		gql_constants.TYPE:     todoType,
	}
}

type UserFilters struct {
	BaseFilters
	UserFilters *gql.UserRoleFilter
}

func (u *UserFilters) GetFilters() map[string]*string {
	role := extractUserRole(u.UserFilters)

	return map[string]*string{
		gql_constants.FIRST:  u.First,
		gql_constants.AFTER:  u.After,
		gql_constants.LAST:   u.Last,
		gql_constants.BEFORE: u.Before,
		gql_constants.ROLE:   role,
	}
}

func extractUserRole(filters *gql.UserRoleFilter) *string {
	var role *string

	if filters != nil {
		if filters.Role != nil {
			role = fromStringPointerToLowerStringPointer((*string)(filters.Role))
		}
	}

	return role
}

func fromStringPointerToLowerStringPointer(ptr *string) *string {
	value := bytes.ToLower([]byte(*ptr))
	stringValue := string(value)

	return &stringValue
}

func extractPriorityAndStatus(todoFilters *gql.TodosFilterInput) (*string, *string) {
	var pr *string
	var st *string

	if todoFilters != nil {
		if todoFilters.Priority != nil {
			pr = fromStringPointerToLowerStringPointer((*string)(todoFilters.Priority))
		}

		if todoFilters.Status != nil {
			st = fromStringPointerToLowerStringPointer((*string)(todoFilters.Status))
		}
	}

	return pr, st
}

func extractType(todoFilters *gql.TodosFilterInput) *string {
	trueValue := constants.TRUE_VALUE
	falseValue := constants.FALSE_VALUE

	var todoType *string
	if todoFilters != nil {
		if todoFilters.Type != nil {
			if string(*todoFilters.Type) == constants.EXPIRED {
				todoType = &trueValue
			} else {
				todoType = &falseValue
			}
		}
	}
	return todoType
}
