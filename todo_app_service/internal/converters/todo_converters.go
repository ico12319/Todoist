package converters

import (
	"Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/internal/utils"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"Todo-List/internProject/todo_app_service/pkg/handler_models"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"Todo-List/internProject/todo_app_service/pkg/pagination"
	"github.com/gofrs/uuid"
)

type todoConverter struct{}

func NewTodoConverter() *todoConverter {
	return &todoConverter{}
}

func (t *todoConverter) ToModel(todo *entities.Todo) *models.Todo {
	dueDate := utils.ExtractDueDateValueFromSQLNull(todo)
	assignedTo := utils.ConvertFromNullUuidToStringPtr(todo.AssignedTo)

	return &models.Todo{
		Id:          todo.Id.String(),
		Name:        todo.Name,
		Description: todo.Description,
		ListId:      todo.ListId.String(),
		Status:      constants.TodoStatus(todo.Status),
		CreatedAt:   todo.CreatedAt,
		LastUpdated: todo.LastUpdated,
		Priority:    constants.Priority(todo.Priority),
		AssignedTo:  assignedTo,
		DueDate:     dueDate,
	}
}

func (t *todoConverter) ToEntity(todo *models.Todo) *entities.Todo {
	dueDate := utils.ConvertFromPointerToSQLNullTime(todo.DueDate)
	assignedTo := utils.ConvertFromPointerToNullUUID(todo.AssignedTo)

	return &entities.Todo{
		Id:          uuid.FromStringOrNil(todo.Id),
		Name:        todo.Name,
		Description: todo.Description,
		ListId:      uuid.FromStringOrNil(todo.ListId),
		Status:      string(todo.Status),
		CreatedAt:   todo.CreatedAt,
		LastUpdated: todo.LastUpdated,
		AssignedTo:  assignedTo,
		DueDate:     dueDate,
		Priority:    string(todo.Priority),
	}
}

func (t *todoConverter) ConvertFromUpdateHandlerModelToModel(todo *handler_models.UpdateTodo) *models.Todo {
	var modelTodo models.Todo

	if todo.Name != nil {
		modelTodo.Name = *todo.Name
	}
	if todo.Description != nil {
		modelTodo.Description = *todo.Description
	}
	if todo.Status != nil {
		modelTodo.Status = *todo.Status
	}
	if todo.Priority != nil {
		modelTodo.Priority = *todo.Priority
	}

	modelTodo.AssignedTo = todo.AssignedTo
	modelTodo.DueDate = todo.DueDate

	return &modelTodo
}

func (*todoConverter) ConvertFromCreateHandlerModelToModel(todo *handler_models.CreateTodo) *models.Todo {
	return &models.Todo{
		Name:        todo.Name,
		Description: todo.Description,
		ListId:      todo.ListId,
		Priority:    todo.Priority,
		AssignedTo:  todo.AssignedTo,
		DueDate:     todo.DueDate,
	}
}

func (t *todoConverter) ManyToPage(todos []entities.Todo, pageInfo *entities.PaginationInfo) *models.TodoPage {
	if len(todos) == 0 || pageInfo == nil || !pageInfo.LastID.Valid || !pageInfo.FirstID.Valid {
		return &models.TodoPage{
			Data: make([]*models.Todo, 0),
			PageInfo: &pagination.Page{
				HasNextPage: false,
				HasPrevPage: false,
			},
			TotalCount: 0,
		}
	}

	modelsTodos := make([]*models.Todo, len(todos))
	for index, entity := range todos {
		model := t.ToModel(&entity)
		modelsTodos[index] = model
	}

	startCursor := modelsTodos[0].Id
	endCursor := modelsTodos[len(todos)-1].Id

	lastEntityID := pageInfo.LastID.UUID.String()
	firstEntityID := pageInfo.FirstID.UUID.String()

	return &models.TodoPage{
		Data:       modelsTodos,
		TotalCount: pageInfo.TotalCount,
		PageInfo: &pagination.Page{
			StartCursor: startCursor,
			EndCursor:   endCursor,
			HasNextPage: lastEntityID != endCursor,
			HasPrevPage: firstEntityID != startCursor,
		},
	}
}
