package converters

import (
	"Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"Todo-List/internProject/todo_app_service/pkg/handler_models"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"Todo-List/internProject/todo_app_service/pkg/pagination"
	"github.com/gofrs/uuid"
)

type userConverter struct{}

func NewUserConverter() *userConverter {
	return &userConverter{}
}

func (*userConverter) ToModel(user *entities.User) *models.User {
	if user == nil {
		return nil
	}

	return &models.User{
		Id:    user.Id.String(),
		Email: user.Email,
		Role:  constants.UserRole(user.Role),
	}
}

func (*userConverter) ToEntity(user *models.User) *entities.User {
	return &entities.User{
		Id:    uuid.FromStringOrNil(user.Id),
		Email: user.Email,
		Role:  string(user.Role),
	}
}

func (*userConverter) ConvertFromUpdateModelToModel(user *handler_models.UpdateUser) *models.User {
	var modelUser models.User

	if user.Email != nil {
		modelUser.Email = *user.Email
	}

	if user.Role != nil {
		modelUser.Role = constants.UserRole(*user.Role)
	}

	return &modelUser
}

func (*userConverter) ConvertFromCreateHandlerModelToModel(user *handler_models.CreateUser) *models.User {
	return &models.User{
		Email: user.Email,
		Role:  constants.UserRole(user.Role),
	}
}

func (u *userConverter) ManyToPage(users []entities.User, pageInfo *entities.PaginationInfo) *models.UserPage {
	if len(users) == 0 || pageInfo == nil || !pageInfo.LastID.Valid || !pageInfo.FirstID.Valid {
		return &models.UserPage{
			Data: make([]*models.User, 0),
			PageInfo: &pagination.Page{
				HasNextPage: false,
				HasPrevPage: false,
			},
			TotalCount: 0,
		}
	}

	modelUsers := make([]*models.User, 0, len(users))
	for _, entity := range users {
		model := u.ToModel(&entity)
		modelUsers = append(modelUsers, model)
	}

	startCursor := modelUsers[0].Id
	endCursor := modelUsers[len(users)-1].Id

	lastEntityID := pageInfo.LastID.UUID.String()
	firstEntityID := pageInfo.FirstID.UUID.String()

	return &models.UserPage{
		Data:       modelUsers,
		TotalCount: pageInfo.TotalCount,
		PageInfo: &pagination.Page{
			StartCursor: startCursor,
			EndCursor:   endCursor,
			HasNextPage: endCursor != lastEntityID,
			HasPrevPage: startCursor != firstEntityID,
		},
	}
}
