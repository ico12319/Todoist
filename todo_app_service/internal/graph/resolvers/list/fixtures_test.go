package list

import (
	"fmt"
	gql "github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/model"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/models"
	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"time"
)

var (
	listName                                 = "name"
	listDescription                          = "description"
	listId                                   = uuid.New()
	invalidListId                            = uuid.New()
	createdAt                                = time.Date(2021, time.January, 15, 10, 30, 0, 0, time.UTC)
	lastUpdate                               = time.Date(2025, time.January, 15, 17, 30, 0, 0, time.UTC)
	expectedErrorWhenMakingResponseFailsList = fmt.Errorf("failed to fetch list response")
	listOwner                                = "owner"
	updatedListName                          = "updatedListName"
	updatedDescription                       = "updated description"
	gqlInternalServerError                   = &gqlerror.Error{
		Message:    "Internal error, please try again later.",
		Extensions: map[string]interface{}{"code": "INTERNAL_SERVER_ERROR"},
	}
	gqlBadRequestError = &gqlerror.Error{
		Message:    "Invalid Request",
		Extensions: map[string]interface{}{"code": "BAD_REQUEST"},
	}
	userId = uuid.New()
)

func initModelList() *models.List {
	return &models.List{
		Id:          listId.String(),
		Name:        listName,
		Description: listDescription,
		CreatedAt:   createdAt,
		LastUpdated: lastUpdate,
		Owner:       listOwner,
	}
}

func initGqlModel() *gql.List {
	return &gql.List{
		ID:          listId.String(),
		Name:        listName,
		Description: listDescription,
		CreatedAt:   createdAt,
		LastUpdated: lastUpdate,
	}
}

func initSuccessfulGqlDeleteListPayload() *gql.DeleteListPayload {
	return &gql.DeleteListPayload{
		ID:          listId.String(),
		Name:        &listName,
		Description: &listDescription,
		CreatedAt:   &createdAt,
		Success:     true,
	}
}
