package url_filters

import (
	gql "Todo-List/internProject/graphQL_service/graph/model"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"bytes"
)

type BaseFilters struct {
	Limit  *string
	Cursor *string
}

func (b *BaseFilters) GetFilters() map[string]*string {
	return map[string]*string{
		constants.LIMIT:  b.Limit,
		constants.CURSOR: b.Cursor,
	}
}

type TodoFilters struct {
	BaseFilters
	TodoFilters *gql.TodosFilterInput
}

func (t *TodoFilters) GetFilters() map[string]*string {
	pr, st := extractPriorityAndStatus(t.TodoFilters)

	return map[string]*string{
		constants.LIMIT:    t.BaseFilters.Limit,
		constants.CURSOR:   t.BaseFilters.Cursor,
		constants.PRIORITY: pr,
		constants.STATUS:   st,
	}
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
