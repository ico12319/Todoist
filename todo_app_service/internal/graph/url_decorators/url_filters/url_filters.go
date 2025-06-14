package url_filters

import (
	gql "github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/model"
)

type BaseFilters struct {
	Limit  *int32
	Cursor *string
}

type TodoFilters struct {
	BaseFilters
	TodoFilters *gql.TodosFilterInput
}
