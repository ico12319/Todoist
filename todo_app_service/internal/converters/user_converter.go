package converters

import (
	"Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"Todo-List/internProject/todo_app_service/pkg/handler_models"
	"Todo-List/internProject/todo_app_service/pkg/models"
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

func (u *userConverter) ManyToModel(users []entities.User) []*models.User {
	modelUsers := make([]*models.User, 0, len(users))

	for _, entity := range users {
		model := u.ToModel(&entity)
		modelUsers = append(modelUsers, model)
	}

	return modelUsers
}
