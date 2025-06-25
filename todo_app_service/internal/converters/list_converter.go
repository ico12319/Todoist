package converters

import (
	"Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/pkg/handler_models"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"github.com/gofrs/uuid"
)

type listConverter struct{}

func NewListConverter() *listConverter {
	return &listConverter{}
}

func (*listConverter) ConvertFromDBEntityToModel(list *entities.List) *models.List {
	return &models.List{
		Id:          list.Id.String(),
		Name:        list.Name,
		CreatedAt:   list.CreatedAt,
		LastUpdated: list.LastUpdated,
		Owner:       list.Owner.String(),
		Description: list.Description,
	}
}

func (*listConverter) ConvertFromModelToDBEntity(list *models.List) *entities.List {
	return &entities.List{
		Id:          uuid.FromStringOrNil(list.Id),
		Name:        list.Name,
		CreatedAt:   list.CreatedAt,
		LastUpdated: list.LastUpdated,
		Owner:       uuid.FromStringOrNil(list.Owner),
		Description: list.Description,
	}
}

func (l *listConverter) ManyToModel(lists []entities.List) []*models.List {
	modelsLists := make([]*models.List, 0, len(lists))

	for _, entity := range lists {
		model := l.ConvertFromDBEntityToModel(&entity)
		modelsLists = append(modelsLists, model)
	}

	return modelsLists
}

func (*listConverter) FromCreateHandlerModelToModel(list *handler_models.CreateList) *models.List {
	return &models.List{
		Name:        list.Name,
		Description: list.Description,
	}
}

func (*listConverter) FromUpdateHandlerModelToModel(list *handler_models.UpdateList) *models.List {
	var modelList models.List

	if list.Name != nil {
		modelList.Name = *list.Name
	}

	if list.Description != nil {
		modelList.Description = *list.Description
	}

	return &modelList
}
