package lists

import (
	application_errors2 "Todo-List/internProject/todo_app_service/internal/application_errors"
	entities2 "Todo-List/internProject/todo_app_service/internal/entities"
	mocks2 "Todo-List/internProject/todo_app_service/internal/lists/mocks"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	models2 "Todo-List/internProject/todo_app_service/pkg/models"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestService_GetListOwnerRecord(t *testing.T) {
	entityUser := initEntityUser(userId, userEmail, adminRole)
	modelUser := initModelUser(userId.String(), userEmail, adminRole)

	tests := []struct {
		testName          string
		listId            string
		mockRepo          func() *mocks2.ListRepo
		mockUserConverter func() *mocks2.IUserConverter
		expectedUserModel *models2.User
		expectError       bool
		err               error
	}{
		{
			testName: "Successfully getting list owner",
			listId:   existingListId.String(),
			mockRepo: func() *mocks2.ListRepo {
				mock := &mocks2.ListRepo{}
				mock.EXPECT().GetListOwner(existingListId.String()).
					Return(entityUser, nil).Once()
				return mock
			},
			mockUserConverter: func() *mocks2.IUserConverter {
				mock := &mocks2.IUserConverter{}
				mock.EXPECT().ConvertFromDBEntityToModel(entityUser).
					Return(modelUser).Once()
				return mock
			},
			expectedUserModel: modelUser,
			expectError:       false,
		},
		{
			testName: "Unable to get list owner because of an error in list repo caused by invalid list_id",
			listId:   nonExistingListId.String(),
			mockRepo: func() *mocks2.ListRepo {
				mock := &mocks2.ListRepo{}
				mock.EXPECT().GetListOwner(nonExistingListId.String()).
					Return(nil, application_errors2.NewNotFoundError(constants.LIST_TARGET, nonExistingListId.String())).Once()
				return mock
			},
			expectError: true,
			err:         application_errors2.NewNotFoundError(constants.LIST_TARGET, nonExistingListId.String()),
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockRepo := &mocks2.ListRepo{}
			if test.mockRepo != nil {
				mockRepo = test.mockRepo()
			}

			mockUserConverter := &mocks2.IUserConverter{}
			if test.mockUserConverter != nil {
				mockUserConverter = test.mockUserConverter()
			}

			listService := NewService(mockRepo, nil,
				nil, nil, nil, mockUserConverter)

			gotUserModel, err := listService.GetListOwnerRecord(test.listId)
			if test.expectError {
				require.ErrorIs(t, err, test.err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.expectedUserModel, gotUserModel)

			mockUserConverter.AssertExpectations(t)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetListRecord(t *testing.T) {

	entityList := initEntityList(existingListId, listName, testDate, testDate, ownerId)
	modelList := initModelList(existingListId.String(), listName, testDate, testDate, ownerId.String())

	tests := []struct {
		testName          string
		listId            string
		mockRepo          func() *mocks2.ListRepo
		mockListConverter func() *mocks2.IListConverter
		expectedListModel *models2.List
		expectError       bool
		err               error
	}{
		{
			testName: "Successfully getting list record",
			listId:   existingListId.String(),
			mockRepo: func() *mocks2.ListRepo {
				mock := &mocks2.ListRepo{}
				mock.EXPECT().GetList(existingListId.String()).Return(entityList, nil).Once()
				return mock
			},
			mockListConverter: func() *mocks2.IListConverter {
				mock := &mocks2.IListConverter{}
				mock.EXPECT().ConvertFromDBEntityToModel(entityList).Return(modelList).Once()
				return mock
			},
			expectedListModel: modelList,
			expectError:       false,
		},
		{
			testName: "Unable to get list record because of an error in list repo caused by invalid list_id",
			listId:   nonExistingListId.String(),
			mockRepo: func() *mocks2.ListRepo {
				mock := &mocks2.ListRepo{}
				mock.EXPECT().GetList(nonExistingListId.String()).Return(nil,
					application_errors2.NewNotFoundError(constants.LIST_TARGET, nonExistingListId.String())).Once()
				return mock
			},
			expectError: true,
			err:         application_errors2.NewNotFoundError(constants.LIST_TARGET, nonExistingListId.String()),
		},
		{
			testName: "Unable to get list record because of an error in list repo caused by database error",
			listId:   nonExistingListId.String(),
			mockRepo: func() *mocks2.ListRepo {
				mock := &mocks2.ListRepo{}
				mock.EXPECT().GetList(nonExistingListId.String()).Return(nil,
					databaseError).Once()
				return mock
			},
			expectError: true,
			err:         databaseError,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockRepo := &mocks2.ListRepo{}
			if test.mockRepo != nil {
				mockRepo = test.mockRepo()
			}

			mockListConverter := &mocks2.IListConverter{}
			if test.mockListConverter != nil {
				mockListConverter = test.mockListConverter()
			}

			listService := NewService(mockRepo, nil, nil,
				nil, mockListConverter, nil)

			gotModelList, err := listService.GetListRecord(test.listId)
			if test.expectError {
				require.ErrorIs(t, err, test.err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedListModel, gotModelList)
			mockRepo.AssertExpectations(t)
			mockListConverter.AssertExpectations(t)
		})
	}
}

func TestService_DeleteListRecord(t *testing.T) {
	tests := []struct {
		testName    string
		listId      string
		mockRepo    func() *mocks2.ListRepo
		expectError bool
		err         error
	}{
		{
			testName: "Successfully deleting list record",
			listId:   existingListId.String(),
			mockRepo: func() *mocks2.ListRepo {
				mock := &mocks2.ListRepo{}
				mock.EXPECT().DeleteList(existingListId.String()).Return(nil).Once()
				return mock
			},
			expectError: false,
			err:         nil,
		},
		{
			testName: "Unable to delete list record because of an error in the list repo layer caused by database error",
			listId:   existingListId.String(),
			mockRepo: func() *mocks2.ListRepo {
				mock := &mocks2.ListRepo{}
				mock.EXPECT().DeleteList(existingListId.String()).Return(databaseError).Once()
				return mock
			},
			expectError: true,
			err:         databaseError,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockRepo := &mocks2.ListRepo{}
			if test.mockRepo != nil {
				mockRepo = test.mockRepo()
			}

			listService := NewService(mockRepo, nil,
				nil, nil, nil, nil)

			err := listService.DeleteListRecord(test.listId)
			if test.expectError {
				require.ErrorIs(t, err, test.err)
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetCollaborators(t *testing.T) {

	userEntity1 := initEntityUser(userId, userEmail, adminRole)
	userEntity2 := initEntityUser(userId2, userEmail2, readerRole)
	userEntity3 := initEntityUser(userId3, userEmail3, writerRole)

	userModel1 := initModelUser(userId.String(), userEmail, adminRole)
	userModel2 := initModelUser(userId2.String(), userEmail2, readerRole)
	userModel3 := initModelUser(userId3.String(), userEmail3, writerRole)

	userEntities := []entities2.User{
		*userEntity1,
		*userEntity2,
		*userEntity3,
	}
	userModels := []*models2.User{
		userModel1,
		userModel2,
		userModel3,
	}

	listEntity := initEntityList(existingListId, listName, testDate, testDate, ownerId)

	tests := []struct {
		testName          string
		listId            string
		mockRepo          func() *mocks2.ListRepo
		mockUserConverter func() *mocks2.IUserConverter
		expectUserModels  []*models2.User
		expectError       bool
		err               error
	}{
		{
			testName: "Successfully getting list's collaborators",
			listId:   existingListId.String(),
			mockRepo: func() *mocks2.ListRepo {
				mock := &mocks2.ListRepo{}
				mock.EXPECT().GetList(existingListId.String()).Return(listEntity, nil).Once()
				mock.EXPECT().GetListCollaborators(existingListId.String()).Return(userEntities, nil).Once()
				return mock
			},
			mockUserConverter: func() *mocks2.IUserConverter {
				mock := &mocks2.IUserConverter{}
				for index, entity := range userEntities {
					mock.EXPECT().ConvertFromDBEntityToModel(&entity).Return(userModels[index]).Once()
				}
				return mock
			},
			expectUserModels: userModels,
			expectError:      false,
			err:              nil,
		},
		{
			testName: "Unable to get list's collaborators because of an error in list repo layer caused by invalid list_id",
			listId:   nonExistingListId.String(),
			mockRepo: func() *mocks2.ListRepo {
				mock := &mocks2.ListRepo{}
				mock.EXPECT().GetList(nonExistingListId.String()).Return(nil,
					application_errors2.NewNotFoundError(constants.LIST_TARGET, nonExistingListId.String())).Once()
				return mock
			},
			expectUserModels: nil,
			expectError:      true,
			err:              application_errors2.NewNotFoundError(constants.LIST_TARGET, nonExistingListId.String()),
		},
		{
			testName: "Unable to get list's collaborators because of an error in the list repo layer caused by database error when trying to get list",
			listId:   existingListId.String(),
			mockRepo: func() *mocks2.ListRepo {
				mock := &mocks2.ListRepo{}
				mock.EXPECT().GetList(existingListId.String()).Return(nil, databaseError).Once()
				return mock
			},
			expectUserModels: nil,
			expectError:      true,
			err:              databaseError,
		},
		{
			testName: "Unable to get list's collaborators because of an error in the list repo layer when trying to get list's collaborators",
			listId:   existingListId.String(),
			mockRepo: func() *mocks2.ListRepo {
				mock := &mocks2.ListRepo{}
				mock.EXPECT().GetList(existingListId.String()).Return(listEntity, nil).Once()
				mock.EXPECT().GetListCollaborators(existingListId.String()).Return(nil, databaseError).Once()
				return mock
			},
			expectUserModels: nil,
			expectError:      true,
			err:              databaseError,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockRepo := &mocks2.ListRepo{}
			if test.mockRepo != nil {
				mockRepo = test.mockRepo()
			}

			mockUserConverter := &mocks2.IUserConverter{}
			if test.mockUserConverter != nil {
				mockUserConverter = test.mockUserConverter()
			}

			listService := NewService(mockRepo, nil, nil,
				nil, nil, mockUserConverter)

			gotUserModels, err := listService.GetCollaborators(test.listId)

			if test.expectError {
				require.ErrorIs(t, err, test.err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.expectUserModels, gotUserModels)

			mockRepo.AssertExpectations(t)
			mockUserConverter.AssertExpectations(t)
		})
	}
}

func TestService_GetListsRecords(t *testing.T) {
	listEntities := []entities2.List{
		*initEntityList(existingListId, listName, testDate, testDate, ownerId),
		*initEntityList(listId2, listName, testDate, testDate, ownerId),
		*initEntityList(listId3, listName, testDate, testDate, ownerId),
	}

	listModels := []*models2.List{
		initModelList(existingListId.String(), listName, testDate, testDate, ownerId.String()),
		initModelList(listId2.String(), listName, testDate, testDate, ownerId.String()),
		initModelList(listId3.String(), listName, testDate, testDate, ownerId.String()),
	}

	tests := []struct {
		testName           string
		mockRepo           func() *mocks2.ListRepo
		mockListConverter  func() *mocks2.IListConverter
		expectedListModels []*models2.List
		expectError        bool
		err                error
	}{
		{
			testName: "Successfully getting list records",
			mockRepo: func() *mocks2.ListRepo {
				mock := &mocks2.ListRepo{}
				mock.EXPECT().GetLists().Return(listEntities, nil).Once()
				return mock
			},
			mockListConverter: func() *mocks2.IListConverter {
				mock := &mocks2.IListConverter{}

				for index, entity := range listEntities {
					mock.EXPECT().ConvertFromDBEntityToModel(&entity).Return(listModels[index]).Once()
				}
				return mock
			},
			expectedListModels: listModels,
			expectError:        false,
			err:                nil,
		},
		{
			testName: "Unable to get list records because of an error in list repo layer caused by database error",
			mockRepo: func() *mocks2.ListRepo {
				mock := &mocks2.ListRepo{}
				mock.EXPECT().GetLists().Return(nil, databaseError).Once()
				return mock
			},
			expectedListModels: nil,
			expectError:        true,
			err:                databaseError,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockRepo := &mocks2.ListRepo{}
			if test.mockRepo != nil {
				mockRepo = test.mockRepo()
			}

			mockListConverter := &mocks2.IListConverter{}
			if test.mockListConverter != nil {
				mockListConverter = test.mockListConverter()
			}

			listService := NewService(mockRepo, nil, nil,
				nil, mockListConverter, nil)

			gotListModels, err := listService.GetListsRecords()

			if test.expectError {
				require.ErrorIs(t, err, test.err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.expectedListModels, gotListModels)
			mockRepo.AssertExpectations(t)
			mockListConverter.AssertExpectations(t)
		})
	}
}

func TestService_ChangelistName(t *testing.T) {
	tests := []struct {
		testName    string
		listId      string
		newName     string
		mockRepo    func() *mocks2.ListRepo
		expectError bool
		err         error
	}{
		{
			testName: "Successfully changing list  name",
			listId:   existingListId.String(),
			newName:  newListName,
			mockRepo: func() *mocks2.ListRepo {
				mock := &mocks2.ListRepo{}
				mock.EXPECT().UpdateListName(existingListId.String(), newListName).
					Return(nil).Once()
				return mock
			},
			expectError: false,
			err:         nil,
		},
		{
			testName: "Unable to change list name because of an error in list repo caused by invalid list_id",
			listId:   nonExistingListId.String(),
			newName:  newListName,
			mockRepo: func() *mocks2.ListRepo {
				mock := &mocks2.ListRepo{}
				mock.EXPECT().UpdateListName(nonExistingListId.String(), newListName).
					Return(application_errors2.NewNotFoundError(constants.LIST_TARGET, nonExistingListId.String())).Once()
				return mock
			},
			expectError: true,
			err:         application_errors2.NewNotFoundError(constants.LIST_TARGET, nonExistingListId.String()),
		},
		{
			testName: "Unable to change list name because of an error in list repo caused by database error",
			listId:   existingListId.String(),
			newName:  newListName,
			mockRepo: func() *mocks2.ListRepo {
				mock := &mocks2.ListRepo{}
				mock.EXPECT().UpdateListName(existingListId.String(), newListName).
					Return(databaseError).Once()
				return mock
			},
			expectError: true,
			err:         databaseError,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockRepo := &mocks2.ListRepo{}
			if test.mockRepo != nil {
				mockRepo = test.mockRepo()
			}

			listService := NewService(mockRepo, nil,
				nil, nil, nil, nil)

			err := listService.ChangelistName(test.listId, test.newName)

			if test.expectError {
				require.ErrorIs(t, err, test.err)
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_CreateListRecord(t *testing.T) {
	listEntity := initEntityList(existingListId, listName, testDate, testDate, ownerId)
	listModel := initModelList(existingListId.String(), listName, testDate, testDate, ownerId.String())

	tests := []struct {
		testName          string
		listName          string
		owner             string
		mockRepo          func() *mocks2.ListRepo
		mockUuidGenerator func() *mocks2.IUuidGenerator
		mockTimeGenerator func() *mocks2.ITimeGenerator
		mockListConverter func() *mocks2.IListConverter
		expectedListModel *models2.List
		expectError       bool
		err               error
	}{
		{
			testName: "Successfully creating list record",
			listName: listName,
			owner:    ownerId.String(),
			mockRepo: func() *mocks2.ListRepo {
				mock := &mocks2.ListRepo{}
				mock.EXPECT().CreateList(listEntity).
					Return(listEntity, nil).Once()
				return mock
			},
			mockUuidGenerator: func() *mocks2.IUuidGenerator {
				mock := &mocks2.IUuidGenerator{}
				mock.EXPECT().Generate().
					Return(existingListId.String()).Once()
				return mock
			},
			mockTimeGenerator: func() *mocks2.ITimeGenerator {
				mock := &mocks2.ITimeGenerator{}
				mock.EXPECT().Now().
					Return(testDate).Times(2)
				return mock
			},
			mockListConverter: func() *mocks2.IListConverter {
				mock := &mocks2.IListConverter{}
				mock.EXPECT().ConvertFromModelToDBEntity(listModel).
					Return(listEntity).Once()
				mock.EXPECT().ConvertFromDBEntityToModel(listEntity).
					Return(listModel).Once()
				return mock
			},
			expectedListModel: listModel,
			expectError:       false,
			err:               nil,
		},
		{
			testName: "Unable to create list because of an error in the list repo layer caused by already existing list name",
			listName: listName,
			owner:    ownerId.String(),
			mockRepo: func() *mocks2.ListRepo {
				mock := &mocks2.ListRepo{}
				mock.EXPECT().CreateList(listEntity).
					Return(nil, application_errors2.NewAlreadyExistError(constants.LIST_TARGET, listName)).Once()
				return mock
			},
			mockUuidGenerator: func() *mocks2.IUuidGenerator {
				mock := &mocks2.IUuidGenerator{}
				mock.EXPECT().Generate().
					Return(existingListId.String()).Once()
				return mock
			},
			mockTimeGenerator: func() *mocks2.ITimeGenerator {
				mock := &mocks2.ITimeGenerator{}
				mock.EXPECT().Now().
					Return(testDate).Times(2)
				return mock
			},
			mockListConverter: func() *mocks2.IListConverter {
				mock := &mocks2.IListConverter{}
				mock.EXPECT().ConvertFromModelToDBEntity(listModel).
					Return(listEntity).Once()
				return mock
			},
			expectedListModel: nil,
			expectError:       true,
			err:               application_errors2.NewAlreadyExistError(constants.LIST_TARGET, listName),
		},
		{
			testName: "Unable to create list because of an error in the list repo layer caused by invalid user_id",
			listName: listName,
			owner:    ownerId.String(),
			mockRepo: func() *mocks2.ListRepo {
				mock := &mocks2.ListRepo{}
				mock.EXPECT().CreateList(listEntity).
					Return(nil, application_errors2.NewNotFoundError(constants.USER_TARGET, ownerId.String())).Once()
				return mock
			},
			mockUuidGenerator: func() *mocks2.IUuidGenerator {
				mock := &mocks2.IUuidGenerator{}
				mock.EXPECT().Generate().
					Return(existingListId.String()).Once()
				return mock
			},
			mockTimeGenerator: func() *mocks2.ITimeGenerator {
				mock := &mocks2.ITimeGenerator{}
				mock.EXPECT().Now().
					Return(testDate).Times(2)
				return mock
			},
			mockListConverter: func() *mocks2.IListConverter {
				mock := &mocks2.IListConverter{}
				mock.EXPECT().ConvertFromModelToDBEntity(listModel).
					Return(listEntity).Once()
				return mock
			},
			expectedListModel: nil,
			expectError:       true,
			err:               application_errors2.NewNotFoundError(constants.USER_TARGET, ownerId.String()),
		},
		{
			testName: "Unable to create list because of an error in the list repo layer caused by database error",
			listName: listName,
			owner:    ownerId.String(),
			mockRepo: func() *mocks2.ListRepo {
				mock := &mocks2.ListRepo{}
				mock.EXPECT().CreateList(listEntity).
					Return(nil, databaseError).Once()
				return mock
			},
			mockUuidGenerator: func() *mocks2.IUuidGenerator {
				mock := &mocks2.IUuidGenerator{}
				mock.EXPECT().Generate().
					Return(existingListId.String()).Once()
				return mock
			},
			mockTimeGenerator: func() *mocks2.ITimeGenerator {
				mock := &mocks2.ITimeGenerator{}
				mock.EXPECT().Now().
					Return(testDate).Times(2)
				return mock
			},
			mockListConverter: func() *mocks2.IListConverter {
				mock := &mocks2.IListConverter{}
				mock.EXPECT().ConvertFromModelToDBEntity(listModel).
					Return(listEntity).Once()
				return mock
			},
			expectedListModel: nil,
			expectError:       true,
			err:               databaseError,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockRepo := &mocks2.ListRepo{}
			if test.mockRepo != nil {
				mockRepo = test.mockRepo()
			}

			mockUuidGenerator := &mocks2.IUuidGenerator{}
			if test.mockUuidGenerator != nil {
				mockUuidGenerator = test.mockUuidGenerator()
			}

			mockTimeGenerator := &mocks2.ITimeGenerator{}
			if test.mockTimeGenerator != nil {
				mockTimeGenerator = test.mockTimeGenerator()
			}

			mockListConverter := &mocks2.IListConverter{}
			if test.mockListConverter != nil {
				mockListConverter = test.mockListConverter()
			}

			listService := NewService(mockRepo, nil,
				mockUuidGenerator, mockTimeGenerator, mockListConverter, nil)

			gotModelList, err := listService.CreateListRecord(test.listName, test.owner)
			if test.expectError {
				require.ErrorIs(t, err, test.err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.expectedListModel, gotModelList)

			mockRepo.AssertExpectations(t)
			mockUuidGenerator.AssertExpectations(t)
			mockTimeGenerator.AssertExpectations(t)
			mockListConverter.AssertExpectations(t)
		})
	}
}

func TestService_AddCollaborator(t *testing.T) {
	tests := []struct {
		testName     string
		listId       string
		userEmail    string
		mockUserRepo func() *mocks2.UserRepo
		mockListRepo func() *mocks2.ListRepo
		expectError  bool
		err          error
	}{
		{
			testName:  "Successfully adding a collaborator to the list",
			listId:    existingListId.String(),
			userEmail: userEmail,
			mockUserRepo: func() *mocks2.UserRepo {
				mock := &mocks2.UserRepo{}
				mock.EXPECT().GetUserIdByEmail(userEmail).
					Return(userId, nil).Once()
				return mock
			},
			mockListRepo: func() *mocks2.ListRepo {
				mock := &mocks2.ListRepo{}
				mock.EXPECT().UpdateListSharedWith(existingListId.String(), userId.String()).
					Return(nil).Once()
				return mock
			},
			expectError: false,
			err:         nil,
		},
		{
			testName:  "Unable to add collaborator because of an error in the list repo caused by invalid user email error",
			listId:    existingListId.String(),
			userEmail: userEmail,
			mockUserRepo: func() *mocks2.UserRepo {
				mock := &mocks2.UserRepo{}
				mock.EXPECT().GetUserIdByEmail(userEmail).
					Return(uuid.UUID{}, application_errors2.NewAlreadyExistError(constants.USER_TARGET, userEmail)).Once()
				return mock
			},
			expectError: true,
			err:         application_errors2.NewAlreadyExistError(constants.USER_TARGET, userEmail),
		},
		{
			testName:  "Unable to add collaborator because of an error in the list repo caused by invalid list_id error",
			listId:    nonExistingListId.String(),
			userEmail: userEmail,
			mockUserRepo: func() *mocks2.UserRepo {
				mock := &mocks2.UserRepo{}
				mock.EXPECT().GetUserIdByEmail(userEmail).
					Return(userId, nil).Once()
				return mock
			},
			mockListRepo: func() *mocks2.ListRepo {
				mock := &mocks2.ListRepo{}
				mock.EXPECT().UpdateListSharedWith(nonExistingListId.String(), userId.String()).
					Return(application_errors2.NewNotFoundError(constants.LIST_TARGET, nonExistingListId.String())).Once()
				return mock
			},
			expectError: true,
			err:         application_errors2.NewNotFoundError(constants.LIST_TARGET, nonExistingListId.String()),
		},
		{
			testName:  "Unable to add collaborator because of an error in the list repo caused by database error",
			listId:    existingListId.String(),
			userEmail: userEmail,
			mockUserRepo: func() *mocks2.UserRepo {
				mock := &mocks2.UserRepo{}
				mock.EXPECT().GetUserIdByEmail(userEmail).
					Return(userId, nil).Once()
				return mock
			},
			mockListRepo: func() *mocks2.ListRepo {
				mock := &mocks2.ListRepo{}
				mock.EXPECT().UpdateListSharedWith(existingListId.String(), userId.String()).
					Return(databaseError).Once()
				return mock
			},
			expectError: true,
			err:         databaseError,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockUserRepo := &mocks2.UserRepo{}
			if test.mockUserRepo != nil {
				mockUserRepo = test.mockUserRepo()
			}

			mockListRepo := &mocks2.ListRepo{}
			if test.mockListRepo != nil {
				mockListRepo = test.mockListRepo()
			}

			listService := NewService(mockListRepo, mockUserRepo, nil,
				nil, nil, nil)

			err := listService.AddCollaborator(test.listId, test.userEmail)
			if test.expectError {
				require.ErrorIs(t, err, test.err)
			} else {
				require.NoError(t, err)
			}

			mockUserRepo.AssertExpectations(t)
			mockListRepo.AssertExpectations(t)
		})
	}
}
