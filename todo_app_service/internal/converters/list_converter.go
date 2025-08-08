package converters

import (
	"Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/pkg/handler_models"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"Todo-List/internProject/todo_app_service/pkg/pagination"
	"github.com/gofrs/uuid"
)

type listConverter struct{}

func NewListConverter() *listConverter {
	return &listConverter{}
}

func (*listConverter) ToModel(list *entities.List) *models.List {
	return &models.List{
		Id:          list.Id.String(),
		Name:        list.Name,
		CreatedAt:   list.CreatedAt,
		LastUpdated: list.LastUpdated,
		Owner:       list.Owner.String(),
		Description: list.Description,
	}
}

func (*listConverter) ToEntity(list *models.List) *entities.List {
	return &entities.List{
		Id:          uuid.FromStringOrNil(list.Id),
		Name:        list.Name,
		CreatedAt:   list.CreatedAt,
		LastUpdated: list.LastUpdated,
		Owner:       uuid.FromStringOrNil(list.Owner),
		Description: list.Description,
	}
}

func (l *listConverter) ManyToPage(lists []entities.List, pageInfo *entities.PaginationInfo) *models.ListPage {
	if len(lists) == 0 || pageInfo == nil || !pageInfo.FirstID.Valid || !pageInfo.LastID.Valid {
		return &models.ListPage{
			Data: make([]*models.List, 0),
			PageInfo: &pagination.Page{
				HasNextPage: false,
				HasPrevPage: false,
			},
			TotalCount: 0,
		}
	}

	modelsLists := make([]*models.List, 0, len(lists))
	for _, entity := range lists {
		model := l.ToModel(&entity)
		modelsLists = append(modelsLists, model)
	}

	startCursor := modelsLists[0].Id
	endCursor := modelsLists[len(modelsLists)-1].Id

	lastEntityID := pageInfo.LastID.UUID.String()
	firstEntityID := pageInfo.FirstID.UUID.String()

	return &models.ListPage{
		Data:       modelsLists,
		TotalCount: pageInfo.TotalCount,
		PageInfo: &pagination.Page{
			StartCursor: startCursor,
			EndCursor:   endCursor,
			HasNextPage: lastEntityID != endCursor,
			HasPrevPage: firstEntityID != startCursor,
		},
	}
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
