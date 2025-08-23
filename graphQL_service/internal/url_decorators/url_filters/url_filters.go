package url_filters

import (
	gql "Todo-List/internProject/graphQL_service/graph/model"
	"Todo-List/internProject/graphQL_service/internal/gql_constants"
	"bytes"
)

type statusConverter interface {
	ToStringStatus(status *gql.TodoStatus) string
}

type priorityConverter interface {
	ToStringPriority(priority *gql.Priority) string
}

type overdueConverter interface {
	ToStringType(overdue *gql.TodoType) string
}

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
	TodoFilters       *gql.TodosFilterInput
	statusConverter   statusConverter
	priorityConverter priorityConverter
	overdueConverter  overdueConverter
}

func NewTodoFilters(b BaseFilters, tFilters *gql.TodosFilterInput, statusConverter statusConverter, priorityConverter priorityConverter, overdueConverter overdueConverter) *TodoFilters {
	return &TodoFilters{
		BaseFilters:       b,
		TodoFilters:       tFilters,
		statusConverter:   statusConverter,
		priorityConverter: priorityConverter,
		overdueConverter:  overdueConverter,
	}
}

func (t *TodoFilters) GetFilters() map[string]*string {
	var convertedStatus string
	var convertedPriority string
	var convertedType string
	var name string

	if t.TodoFilters != nil {
		convertedStatus = t.statusConverter.ToStringStatus(t.TodoFilters.Status)
		convertedPriority = t.priorityConverter.ToStringPriority(t.TodoFilters.Priority)
		convertedType = t.overdueConverter.ToStringType(t.TodoFilters.Type)

		if t.TodoFilters.Name != nil {
			name = *t.TodoFilters.Name
		}
	}

	return map[string]*string{
		gql_constants.FIRST:    t.First,
		gql_constants.AFTER:    t.After,
		gql_constants.LAST:     t.Last,
		gql_constants.BEFORE:   t.Before,
		gql_constants.PRIORITY: &convertedPriority,
		gql_constants.STATUS:   &convertedStatus,
		gql_constants.TYPE:     &convertedType,
		gql_constants.NAME:     &name,
	}
}

type ListFilters struct {
	BaseFilters
	ListFilters *gql.ListFilterInput
}

func NewListFilters(b BaseFilters, lFilters *gql.ListFilterInput) *ListFilters {
	return &ListFilters{
		BaseFilters: b,
		ListFilters: lFilters,
	}
}

func (u *ListFilters) GetFilters() map[string]*string {
	var name string
	if u.ListFilters != nil {
		if u.ListFilters.Name != nil {
			name = *u.ListFilters.Name
		}
	}

	return map[string]*string{
		gql_constants.FIRST:  u.First,
		gql_constants.AFTER:  u.After,
		gql_constants.LAST:   u.Last,
		gql_constants.BEFORE: u.Before,
		gql_constants.NAME:   &name,
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
