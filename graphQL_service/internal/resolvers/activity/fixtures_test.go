package activity

import (
	gql "Todo-List/internProject/graphQL_service/graph/model"
	"Todo-List/internProject/graphQL_service/internal/gql_constants"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"errors"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

var (
	restUrl      = "/rest/url"
	url          = restUrl + gql_constants.ACTIVITIES_PATH + gql_constants.RANDOM_PATH
	activity     = "activity"
	activityType = "type"
	participants = 10
	kidFriendly  = true
)

var (
	errByHttpService                 = errors.New("error when trying to get http response")
	errHttpStatusInternalServerError = &gqlerror.Error{
		Message:    "Internal error, please try again later.",
		Extensions: map[string]interface{}{"code": "INTERNAL_SERVER_ERROR"},
	}
	errWhenDecodingJSON = errors.New("EOF")
)

func initExpectedModelRandomActivity() *models.RandomActivity {
	return &models.RandomActivity{
		Activity:     activity,
		Type:         activityType,
		Participants: participants,
		KidFriendly:  kidFriendly,
	}
}

func initExpectedGqlRandomActivity() *gql.RandomActivity {
	return &gql.RandomActivity{
		Activity:     activity,
		Type:         activityType,
		Participants: int32(participants),
		KidFriendly:  kidFriendly,
	}
}

func initEmptyGqlRandomActivity() *gql.RandomActivity {
	return &gql.RandomActivity{
		Activity:     "",
		Type:         "",
		Participants: 0,
		KidFriendly:  false,
	}
}
