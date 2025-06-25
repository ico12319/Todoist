package gql_converters

import (
	gql "Todo-List/internProject/graphQL_service/graph/model"
	"Todo-List/internProject/todo_app_service/pkg/handler_models"
	"Todo-List/internProject/todo_app_service/pkg/models"
)

type accessConverter struct{}

func NewAccessConverter() *accessConverter {
	return &accessConverter{}
}

func (a *accessConverter) ToGQL(refresh *models.CallbackResponse) *gql.Access {
	if refresh == nil {
		return nil
	}

	return &gql.Access{
		JwtToken:     refresh.JwtToken,
		RefreshToken: refresh.RefreshToken,
	}
}

func (a *accessConverter) ToHandlerModelRefresh(refreshInput *gql.RefreshTokenInput) *handler_models.Refresh {
	if refreshInput == nil {
		return nil
	}

	return &handler_models.Refresh{
		RefreshToken: refreshInput.RefreshToken,
	}
}
