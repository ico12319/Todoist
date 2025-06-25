package converters

import (
	"Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"github.com/gofrs/uuid"
)

type refreshConverter struct{}

func NewRefreshConverter() *refreshConverter {
	return &refreshConverter{}
}

func (r *refreshConverter) ToModel(refresh *entities.Refresh) *models.Refresh {
	return &models.Refresh{
		UserId:       refresh.UserId.String(),
		RefreshToken: refresh.RefreshToken,
	}
}

func (r *refreshConverter) ToEntity(refresh *models.Refresh) *entities.Refresh {
	return &entities.Refresh{
		UserId:       uuid.FromStringOrNil(refresh.UserId),
		RefreshToken: refresh.RefreshToken,
	}
}
