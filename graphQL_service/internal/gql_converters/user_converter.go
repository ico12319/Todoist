package gql_converters

import (
	gql "Todo-List/internProject/graphQL_service/graph/model"
	"Todo-List/internProject/todo_app_service/pkg/handler_models"
	"Todo-List/internProject/todo_app_service/pkg/models"
)

type userConverter struct {
	rConverter *roleConverter
}

func NewUserConverter(rConverter *roleConverter) *userConverter {
	return &userConverter{rConverter: rConverter}
}

func (*userConverter) ToGQL(user *models.User) *gql.User {
	if user == nil {
		return nil
	}
	role := gql.UserRole(user.Role)

	return &gql.User{
		ID:    user.Id,
		Email: user.Email,
		Role:  &role,
	}
}

func (u *userConverter) ManyToGQL(users []*models.User) []*gql.User {
	gqlUsers := make([]*gql.User, len(users))

	for index, user := range users {
		gqlUser := u.ToGQL(user)
		gqlUsers[index] = gqlUser
	}
	return gqlUsers
}

func (*userConverter) FromCollaboratorInputToAddCollaboratorHandlerModel(user *gql.CollaboratorInput) *handler_models.AddCollaborator {
	if user == nil {
		return nil
	}

	return &handler_models.AddCollaborator{
		Id: user.UserID,
	}
}

func (*userConverter) FromGQLToDeleteUserPayload(user *gql.User, success bool) *gql.DeleteUserPayload {
	if user == nil {
		return nil
	}

	return &gql.DeleteUserPayload{
		Success: success,
		ID:      user.ID,
		Email:   &user.Email,
		Role:    user.Role,
	}
}

func (u *userConverter) ManyFromGQLToDeleteUserPayload(users []*gql.User, success bool) []*gql.DeleteUserPayload {
	if users == nil {
		return nil
	}

	payloads := make([]*gql.DeleteUserPayload, 0, len(users))
	for _, user := range users {
		payload := u.FromGQLToDeleteUserPayload(user, success)
		payloads = append(payloads, payload)
	}

	return payloads
}
