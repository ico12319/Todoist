package lists

import (
	"Todo-List/internProject/todo_app_service/internal/application_errors"
	"Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/internal/lists/mocks"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"Todo-List/internProject/todo_app_service/pkg/handler_models"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"time"
)

const (
	VALID_LIST_ID    = "valid id"
	INVALID_LIST_ID  = "invalid id"
	INVALID_LIMIT    = "limit"
	VALID_NAME       = "valid name"
	DESCRIPTION      = "description"
	VALID_LIST_ID2   = "valid id2"
	VALID_OWNER_ID1  = "valid owner id1"
	VALID_OWNER_ID2  = "valid owner id2"
	INVALID_OWNER_ID = "invalid owner id"
)

var (
	createdAt                    = time.Date(2021, time.January, 15, 10, 30, 0, 0, time.UTC)
	lastUpdate                   = time.Date(2025, time.January, 15, 17, 30, 0, 0, time.UTC)
	dbError                      = errors.New("db error")
	serviceErrorWhenDeletingList = fmt.Errorf("failed to delete list with id %s", INVALID_LIST_ID)
	mockQuery                    = `SELECT id,name,created_at,last_updated,owner,description FROM (SELECT * FROM lists ORDER BY id)`
	listEntities                 = []entities.List{
		initListEntity(VALID_LIST_ID, "name1", "description1", createdAt, lastUpdate, VALID_OWNER_ID1),
		initListEntity(VALID_LIST_ID2, "name2", "description2", createdAt, lastUpdate, VALID_OWNER_ID2),
	}

	modelLists = []*models.List{
		initModelList(VALID_LIST_ID, "name1", "description1", createdAt, lastUpdate, VALID_OWNER_ID1),
		initModelList(VALID_LIST_ID2, "name2", "description2", createdAt, lastUpdate, VALID_OWNER_ID2),
	}

	invalidLimitError = fmt.Errorf("invalid limit provided %s", INVALID_LIMIT)

	validHandlerModel = initHandlerModel(VALID_NAME, DESCRIPTION)

	convertedModel = &models.List{
		Name:        VALID_NAME,
		Description: DESCRIPTION,
	}

	returnedListUuid = "returned uuid"

	constructedModelListByService = initModelList(returnedListUuid, VALID_NAME, DESCRIPTION, createdAt, lastUpdate, VALID_OWNER_ID1)

	returnedEntityListByListConverterFromModel = initListEntity(returnedListUuid, VALID_NAME, DESCRIPTION, createdAt, lastUpdate, VALID_OWNER_ID1)

	constructedModelListByServiceWithInvalidOwnerId = initModelList(returnedListUuid, VALID_NAME, DESCRIPTION, createdAt, lastUpdate, INVALID_OWNER_ID)

	returnedEntityListByListConverterFromModelWithInvalidOwnerID = initListEntity(returnedListUuid, VALID_NAME, DESCRIPTION, createdAt, lastUpdate, INVALID_OWNER_ID)

	errorWhenTryingToConstructModelListWithInvalidOwnerId = application_errors.NewNotFoundError(constants.USER_TARGET, INVALID_OWNER_ID)

	errorWnTryingToConstructModelListWithAlreadyExistingName = application_errors.NewAlreadyExistError(constants.LIST_TARGET, VALID_NAME)
)

func initModelList(id string, name string, description string, createdAt time.Time, lastUpdate time.Time, ownerId string) *models.List {
	return &models.List{
		Id:          id,
		Name:        name,
		Description: description,
		CreatedAt:   createdAt,
		LastUpdated: lastUpdate,
		Owner:       ownerId,
	}
}

func initListEntity(id string, name string, description string, createdAt time.Time, lastUpdate time.Time, ownerId string) entities.List {
	return entities.List{
		Id:          uuid.FromStringOrNil(id),
		Name:        name,
		Description: description,
		CreatedAt:   createdAt,
		LastUpdated: lastUpdate,
		Owner:       uuid.FromStringOrNil(ownerId),
	}
}

func initHandlerModel(name string, description string) *handler_models.CreateList {
	return &handler_models.CreateList{
		Name:        name,
		Description: description,
	}
}

func initUuidGeneratorMock() *mocks.UuidGenerator {
	mUuidGen := &mocks.UuidGenerator{}

	mUuidGen.EXPECT().
		Generate().
		Return(returnedListUuid).Once()

	return mUuidGen
}

func initTimeGenMock() *mocks.TimeGenerator {
	mTimeGen := &mocks.TimeGenerator{}

	mTimeGen.EXPECT().
		Now().
		Return(createdAt).Once()

	mTimeGen.EXPECT().
		Now().
		Return(lastUpdate).Once()

	return mTimeGen
}
