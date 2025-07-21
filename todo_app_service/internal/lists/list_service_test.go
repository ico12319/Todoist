package lists

import (
	"Todo-List/internProject/todo_app_service/internal/lists/mocks"
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators"
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators/filters"
	"Todo-List/internProject/todo_app_service/pkg/handler_models"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"context"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestService_DeleteListRecord(t *testing.T) {
	tests := []struct {
		testName     string
		mockListRepo func() *mocks.ListRepo
		listId       string
		err          error
	}{
		{
			testName: "Successfully deleting list record",
			mockListRepo: func() *mocks.ListRepo {
				mRepo := &mocks.ListRepo{}

				mRepo.EXPECT().
					DeleteList(context.TODO(), VALID_LIST_ID).
					Return(nil).Once()

				return mRepo
			},
			listId: VALID_LIST_ID,
		},
		{
			testName: "Failed to delete list record",
			mockListRepo: func() *mocks.ListRepo {
				mRepo := &mocks.ListRepo{}

				mRepo.EXPECT().
					DeleteList(context.TODO(), INVALID_LIST_ID).
					Return(dbError).Once()

				return mRepo
			},
			listId: INVALID_LIST_ID,
			err:    serviceErrorWhenDeletingList,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mRepo := &mocks.ListRepo{}
			if test.mockListRepo != nil {
				mRepo = test.mockListRepo()
			}

			lService := NewService(mRepo, nil, nil,
				nil, nil, nil, nil)

			err := lService.DeleteListRecord(context.TODO(), test.listId)

			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, mRepo)
		})
	}
}

func TestSqlListDB_GetLists(t *testing.T) {
	mockBaseRetriever := sql_query_decorators.NewBaseQuery(mockQuery)

	tests := []struct {
		testName                string
		mockSqlDecoratorFactory func() *mocks.SqlDecoratorFactory
		mockListRepo            func() *mocks.ListRepo
		mockListConverter       func() *mocks.ListConverter
		filters                 *filters.ListFilters
		expectedListModel       []*models.List
		err                     error
	}{
		{
			testName: "Successfully getting lists",
			mockSqlDecoratorFactory: func() *mocks.SqlDecoratorFactory {
				mSqlDecoratorFactory := &mocks.SqlDecoratorFactory{}

				mSqlDecoratorFactory.
					EXPECT().
					CreateSqlDecorator(context.TODO(), &filters.ListFilters{}, mockQuery).
					Return(mockBaseRetriever, nil)

				return mSqlDecoratorFactory
			},
			mockListRepo: func() *mocks.ListRepo {
				mListRepo := &mocks.ListRepo{}

				mListRepo.EXPECT().
					GetLists(context.TODO(), mockBaseRetriever).
					Return(listEntities, nil).Once()

				return mListRepo
			},

			mockListConverter: func() *mocks.ListConverter {
				mListConverter := &mocks.ListConverter{}

				mListConverter.EXPECT().
					ManyToModel(listEntities).
					Return(modelLists).Once()

				return mListConverter
			},

			filters: &filters.ListFilters{},

			expectedListModel: modelLists,
		},
		{
			testName: "Failed to get lists, error when trying to create sql decorator",
			mockSqlDecoratorFactory: func() *mocks.SqlDecoratorFactory {
				mSqlDecoratorFactory := &mocks.SqlDecoratorFactory{}

				mSqlDecoratorFactory.
					EXPECT().
					CreateSqlDecorator(context.TODO(), &filters.ListFilters{BaseFilters: filters.BaseFilters{
						Limit: INVALID_LIMIT,
					}}, mockQuery).
					Return(nil, invalidLimitError).Once()

				return mSqlDecoratorFactory
			},

			filters: &filters.ListFilters{BaseFilters: filters.BaseFilters{
				Limit: INVALID_LIMIT,
			}},

			err: invalidLimitError,
		},

		{
			testName: "Failed to get lists, error when trying to call list repo",
			mockSqlDecoratorFactory: func() *mocks.SqlDecoratorFactory {
				mSqlDecoratorFactory := &mocks.SqlDecoratorFactory{}

				mSqlDecoratorFactory.
					EXPECT().
					CreateSqlDecorator(context.TODO(), &filters.ListFilters{}, mockQuery).
					Return(mockBaseRetriever, nil)

				return mSqlDecoratorFactory
			},

			mockListRepo: func() *mocks.ListRepo {
				mListRepo := &mocks.ListRepo{}

				mListRepo.EXPECT().
					GetLists(context.TODO(), mockBaseRetriever).
					Return(nil, dbError).Once()

				return mListRepo
			},

			filters: &filters.ListFilters{},

			err: dbError,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockDecoratorFactory := &mocks.SqlDecoratorFactory{}
			if test.mockSqlDecoratorFactory != nil {
				mockDecoratorFactory = test.mockSqlDecoratorFactory()
			}

			mockListRepo := &mocks.ListRepo{}
			if test.mockListRepo != nil {
				mockListRepo = test.mockListRepo()
			}

			mockListConverter := &mocks.ListConverter{}
			if test.mockListConverter != nil {
				mockListConverter = test.mockListConverter()
			}

			lService := NewService(mockListRepo, nil, nil,
				mockListConverter, nil, nil, mockDecoratorFactory)

			gotModelLists, err := lService.GetListsRecords(context.TODO(), test.filters)
			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedListModel, gotModelLists)
			mock.AssertExpectationsForObjects(t, mockDecoratorFactory, mockListRepo, mockListConverter)
		})
	}
}

func TestService_CreateListRecord(t *testing.T) {

	tests := []struct {
		testName           string
		ownerId            string
		mockListConverter  func() *mocks.ListConverter
		mockRepo           func() *mocks.ListRepo
		mockTimeGen        func() *mocks.TimeGenerator
		mockUuidGen        func() *mocks.UuidGenerator
		passedHandlerModel *handler_models.CreateList
		expectedModelList  *models.List
		err                error
	}{
		{
			testName: "Successfully creating list",
			ownerId:  VALID_OWNER_ID1,

			mockUuidGen: func() *mocks.UuidGenerator {
				mUuidGen := &mocks.UuidGenerator{}

				mUuidGen.EXPECT().
					Generate().
					Return(returnedListUuid).Once()

				return mUuidGen
			},

			mockTimeGen: func() *mocks.TimeGenerator {
				mTimeGen := &mocks.TimeGenerator{}

				mTimeGen.EXPECT().
					Now().
					Return(createdAt).Once()

				mTimeGen.EXPECT().
					Now().
					Return(lastUpdate).Once()

				return mTimeGen
			},

			mockListConverter: func() *mocks.ListConverter {
				mListConverter := &mocks.ListConverter{}

				mListConverter.EXPECT().
					FromCreateHandlerModelToModel(validHandlerModel).
					Return(convertedModel).Once()

				mListConverter.EXPECT().
					ConvertFromModelToDBEntity(constructedModelListByService).
					Return(&returnedEntityListByListConverterFromModel).Once()

				return mListConverter
			},

			mockRepo: func() *mocks.ListRepo {
				mRepo := &mocks.ListRepo{}

				mRepo.EXPECT().
					CreateList(context.TODO(), &returnedEntityListByListConverterFromModel).
					Return(&returnedEntityListByListConverterFromModel, nil).Once()

				return mRepo
			},

			expectedModelList: constructedModelListByService,

			passedHandlerModel: validHandlerModel,
		},

		{
			testName: "Failed to create list, invalid owner id provided in list",

			ownerId: INVALID_OWNER_ID,

			mockUuidGen: func() *mocks.UuidGenerator {
				return initUuidGeneratorMock()
			},

			mockTimeGen: func() *mocks.TimeGenerator {
				return initTimeGenMock()
			},

			mockListConverter: func() *mocks.ListConverter {
				mListConverter := &mocks.ListConverter{}

				mListConverter.EXPECT().
					FromCreateHandlerModelToModel(validHandlerModel).
					Return(convertedModel).Once()

				mListConverter.EXPECT().
					ConvertFromModelToDBEntity(constructedModelListByServiceWithInvalidOwnerId).
					Return(&returnedEntityListByListConverterFromModelWithInvalidOwnerID).Once()

				return mListConverter
			},

			mockRepo: func() *mocks.ListRepo {
				mRepo := &mocks.ListRepo{}

				mRepo.EXPECT().
					CreateList(context.TODO(), &returnedEntityListByListConverterFromModelWithInvalidOwnerID).
					Return(nil, errorWhenTryingToConstructModelListWithInvalidOwnerId).Once()

				return mRepo
			},

			passedHandlerModel: validHandlerModel,

			err: errorWhenTryingToConstructModelListWithInvalidOwnerId,
		},

		{
			testName: "Failed to create list, trying to create list with already existing name",

			ownerId: VALID_OWNER_ID1,

			mockUuidGen: func() *mocks.UuidGenerator {
				return initUuidGeneratorMock()
			},

			mockTimeGen: func() *mocks.TimeGenerator {
				return initTimeGenMock()
			},

			mockListConverter: func() *mocks.ListConverter {
				mListConverter := &mocks.ListConverter{}

				mListConverter.EXPECT().
					FromCreateHandlerModelToModel(validHandlerModel).
					Return(convertedModel).Once()

				mListConverter.EXPECT().
					ConvertFromModelToDBEntity(constructedModelListByService).
					Return(&returnedEntityListByListConverterFromModel).Once()

				return mListConverter
			},

			mockRepo: func() *mocks.ListRepo {
				mRepo := &mocks.ListRepo{}

				mRepo.EXPECT().
					CreateList(context.TODO(), &returnedEntityListByListConverterFromModel).
					Return(nil, errorWnTryingToConstructModelListWithAlreadyExistingName).Once()

				return mRepo
			},

			err: errorWnTryingToConstructModelListWithAlreadyExistingName,

			passedHandlerModel: validHandlerModel,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockUuidGen := &mocks.UuidGenerator{}
			if test.mockUuidGen != nil {
				mockUuidGen = test.mockUuidGen()
			}

			mockTimeGen := &mocks.TimeGenerator{}
			if test.mockTimeGen != nil {
				mockTimeGen = test.mockTimeGen()
			}

			mockListConverter := &mocks.ListConverter{}
			if test.mockListConverter != nil {
				mockListConverter = test.mockListConverter()
			}

			mockRepo := &mocks.ListRepo{}
			if test.mockRepo != nil {
				mockRepo = test.mockRepo()
			}

			lService := NewService(mockRepo, mockUuidGen, mockTimeGen,
				mockListConverter, nil, nil, nil)

			gotModelList, err := lService.CreateListRecord(context.TODO(), test.passedHandlerModel, test.ownerId)
			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedModelList, gotModelList)
			mock.AssertExpectationsForObjects(t, mockUuidGen, mockTimeGen, mockListConverter, mockRepo)
		})
	}
}
