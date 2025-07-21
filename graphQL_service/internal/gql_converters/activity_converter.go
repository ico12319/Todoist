package gql_converters

import (
	gql "Todo-List/internProject/graphQL_service/graph/model"
	"Todo-List/internProject/todo_app_service/pkg/models"
)

type activityConverter struct{}

func NewActivityConverter() *activityConverter {
	return &activityConverter{}
}

func (a *activityConverter) ToGQL(activity *models.RandomActivity) *gql.RandomActivity {
	return &gql.RandomActivity{
		Activity:     activity.Activity,
		Type:         activity.Type,
		Participants: int32(activity.Participants),
		KidFriendly:  activity.KidFriendly,
	}
}
