package gql_converters

import (
	gql "Todo-List/internProject/graphQL_service/graph/model"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"Todo-List/internProject/todo_app_service/pkg/handler_models"
	"Todo-List/internProject/todo_app_service/pkg/models"
)

type iPriorityConverter interface {
	ToStringPriority(priority *gql.Priority) string
}

type iStatusConverter interface {
	ToStringStatus(status *gql.TodoStatus) string
}
type todoConverter struct {
	pConverter iPriorityConverter
	sConverter iStatusConverter
}

func NewTodoConverter(pConverter iPriorityConverter, sConverter iStatusConverter) *todoConverter {
	return &todoConverter{pConverter: pConverter, sConverter: sConverter}
}

func (*todoConverter) ToGQL(todo *models.Todo) *gql.Todo {
	return &gql.Todo{
		ID:          todo.Id,
		Name:        todo.Name,
		Description: todo.Description,
		Status:      gql.TodoStatus(todo.Status),
		CreatedAt:   todo.CreatedAt,
		LastUpdated: todo.LastUpdated,
		Priority:    gql.Priority(todo.Priority),
		DueDate:     todo.DueDate,
	}
}

func (t *todoConverter) ToTodoPageGQL(todoPage *models.TodoPage) *gql.TodoPage {
	if todoPage == nil {
		return &gql.TodoPage{
			Data:       make([]*gql.Todo, 0),
			PageInfo:   nil,
			TotalCount: 0,
		}
	}

	todos := todoPage.Data
	gqlTodos := make([]*gql.Todo, len(todos))

	for index, todo := range todos {
		gqlTodo := t.ToGQL(todo)
		gqlTodos[index] = gqlTodo
	}

	return &gql.TodoPage{
		Data: gqlTodos,
		PageInfo: &gql.PageInfo{
			HasPrevPage: todoPage.PageInfo.HasPrevPage,
			HasNextPage: todoPage.PageInfo.HasNextPage,
			StartCursor: todoPage.PageInfo.StartCursor,
			EndCursor:   todoPage.PageInfo.EndCursor,
		},
		TotalCount: int32(todoPage.TotalCount),
	}
}

func (t *todoConverter) ToHandlerModel(todoInput *gql.UpdateTodoInput) *handler_models.UpdateTodo {
	var status *constants.TodoStatus
	if todoInput.Status != nil {
		s := constants.TodoStatus(t.sConverter.ToStringStatus(todoInput.Status))
		status = &s
	}

	var priority *constants.Priority
	if todoInput.Priority != nil {
		p := constants.Priority(t.pConverter.ToStringPriority(todoInput.Priority))
		priority = &p
	}

	return &handler_models.UpdateTodo{
		Name:        todoInput.Name,
		Description: todoInput.Description,
		Status:      status,
		Priority:    priority,
		AssignedTo:  todoInput.AssignedTo,
		DueDate:     todoInput.DueDate,
	}
}

func (t *todoConverter) CreateTodoInputToModel(todoInput *gql.CreateTodoInput) *handler_models.CreateTodo {
	return &handler_models.CreateTodo{
		Name:        todoInput.Name,
		Description: todoInput.Description,
		ListId:      todoInput.ListID,
		Priority:    constants.Priority(t.pConverter.ToStringPriority(&todoInput.Priority)),
		AssignedTo:  todoInput.AssignedTo,
		DueDate:     todoInput.DueDate,
	}
}

func (*todoConverter) FromGQLModelToDeleteTodoPayload(todo *gql.Todo, success bool) *gql.DeleteTodoPayload {
	return &gql.DeleteTodoPayload{
		Success:     success,
		ID:          todo.ID,
		Name:        &todo.Name,
		Description: &todo.Description,
		Status:      &todo.Status,
		Priority:    &todo.Priority,
		CreatedAt:   &todo.CreatedAt,
		LastUpdated: &todo.LastUpdated,
		DueDate:     todo.DueDate,
	}
}

func (t *todoConverter) ManyToDeleteTodoPayload(todos []*gql.Todo, success bool) []*gql.DeleteTodoPayload {
	deleteTodoPayloads := make([]*gql.DeleteTodoPayload, 0, len(todos))
	for _, todo := range todos {
		converted := t.FromGQLModelToDeleteTodoPayload(todo, success)
		deleteTodoPayloads = append(deleteTodoPayloads, converted)
	}
	return deleteTodoPayloads
}
