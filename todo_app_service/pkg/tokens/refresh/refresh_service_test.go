package refresh

import (
	"Todo-List/internProject/todo_app_service/internal/entities"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"Todo-List/internProject/todo_app_service/pkg/tokens/refresh/mocks"
	"context"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestService_CreateRefreshToken(t *testing.T) {
	validUser := initUser(validUserId.String(), VALID_USER_EMAIL)

	tests := []struct {
		testName        string
		email           string
		refreshToken    string
		mockUserService func() *mocks.UserService
		mockConverter   func() *mocks.Converter
		mockRefreshRepo func() *mocks.RefreshRepository
		expectedOutput  *models.Refresh
		err             error
	}{
		{
			testName: "Successfully creating refresh token",

			email: VALID_USER_EMAIL,

			refreshToken: VALID_REFRESH_TOKEN,

			mockUserService: func() *mocks.UserService {
				mUserService := &mocks.UserService{}

				mUserService.EXPECT().
					GetUserRecordByEmail(context.TODO(), VALID_USER_EMAIL).
					Return(validUser, nil).Once()

				return mUserService
			},
			mockConverter: func() *mocks.Converter {
				mConverter := &mocks.Converter{}

				refresh := initRefresh(VALID_REFRESH_TOKEN, validUserId.String())
				refreshEntity := initRefreshEntityFromModel(refresh)

				mConverter.EXPECT().
					ToEntity(refresh).
					Return(refreshEntity).Once()

				return mConverter
			},

			mockRefreshRepo: func() *mocks.RefreshRepository {
				mRefreshRepo := &mocks.RefreshRepository{}

				refresh := initRefresh(VALID_REFRESH_TOKEN, validUserId.String())
				refreshEntity := initRefreshEntityFromModel(refresh)

				mRefreshRepo.EXPECT().
					CreateRefreshToken(context.TODO(), refreshEntity).
					Return(refreshEntity, nil).Once()

				return mRefreshRepo
			},

			expectedOutput: initRefresh(VALID_REFRESH_TOKEN, validUserId.String()),
		},
		{
			testName: "Failed to create refresh token, user not registered in system",

			email: INVALID_USER_EMAIL,

			mockUserService: func() *mocks.UserService {
				mUserService := &mocks.UserService{}

				mUserService.EXPECT().
					GetUserRecordByEmail(context.TODO(), INVALID_USER_EMAIL).
					Return(nil, errByUserService).Once()
				return mUserService
			},

			err: errByUserService,
		},
		{
			testName: "Failed to create refresh token, error by refresh repository",

			email: VALID_USER_EMAIL,

			refreshToken: INVALID_REFRESH_TOKEN,

			mockUserService: func() *mocks.UserService {
				mUserService := &mocks.UserService{}

				mUserService.EXPECT().
					GetUserRecordByEmail(context.TODO(), VALID_USER_EMAIL).
					Return(validUser, nil).Once()

				return mUserService
			},
			mockConverter: func() *mocks.Converter {
				mConverter := &mocks.Converter{}

				refresh := initRefresh(INVALID_REFRESH_TOKEN, validUserId.String())
				entityRefresh := initRefreshEntityFromModel(refresh)

				mConverter.EXPECT().
					ToEntity(refresh).
					Return(entityRefresh).Once()

				return mConverter
			},

			mockRefreshRepo: func() *mocks.RefreshRepository {
				mRefreshRepo := &mocks.RefreshRepository{}

				refresh := initRefresh(INVALID_REFRESH_TOKEN, validUserId.String())
				entityRefresh := initRefreshEntityFromModel(refresh)

				mRefreshRepo.EXPECT().
					CreateRefreshToken(context.TODO(), entityRefresh).
					Return(nil, errWhenCreatingRefreshToken).Once()

				return mRefreshRepo
			},

			err: errWhenCreatingRefreshToken,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mUserService := &mocks.UserService{}
			if test.mockUserService != nil {
				mUserService = test.mockUserService()
			}

			mConverter := &mocks.Converter{}
			if test.mockConverter != nil {
				mConverter = test.mockConverter()
			}

			mRefreshRepo := &mocks.RefreshRepository{}
			if test.mockRefreshRepo != nil {
				mRefreshRepo = test.mockRefreshRepo()
			}

			rService := NewService(mRefreshRepo, mUserService, mConverter, nil)

			gotRefresh, err := rService.CreateRefreshToken(context.TODO(), test.email, test.refreshToken)

			if test.err != nil {
				require.EqualError(t, test.err, err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedOutput, gotRefresh)
			mock.AssertExpectationsForObjects(t, mUserService, mConverter, mRefreshRepo)
		})
	}
}

func TestService_UpdateRefreshToken(t *testing.T) {
	tests := []struct {
		testName       string
		refreshToken   string
		userId         string
		mockRepo       func() *mocks.RefreshRepository
		mockConverter  func() *mocks.Converter
		expectedOutput *models.Refresh
		err            error
	}{
		{
			testName: "Successfully updating refresh token",

			refreshToken: VALID_REFRESH_TOKEN,

			userId: validUserId.String(),

			mockRepo: func() *mocks.RefreshRepository {
				mRepo := &mocks.RefreshRepository{}

				entityRefresh := &entities.Refresh{
					RefreshToken: VALID_REFRESH_TOKEN,
					UserId:       validUserId,
				}

				mRepo.EXPECT().
					UpdateRefreshToken(context.TODO(), VALID_REFRESH_TOKEN, validUserId.String()).
					Return(entityRefresh, nil).Once()

				return mRepo
			},

			mockConverter: func() *mocks.Converter {
				mConverter := &mocks.Converter{}

				entityRefresh := initRefreshEntity(VALID_REFRESH_TOKEN, validUserId)

				modelRefresh := initRefreshModelFromEntity(entityRefresh)

				mConverter.EXPECT().
					ToModel(entityRefresh).
					Return(modelRefresh).Once()

				return mConverter
			},

			expectedOutput: initRefresh(VALID_REFRESH_TOKEN, validUserId.String()),
		},

		{
			testName: "Failed to update refresh token, error by refresh repository",

			refreshToken: VALID_REFRESH_TOKEN,

			userId: invalidUserId.String(),

			mockRepo: func() *mocks.RefreshRepository {
				mRepo := &mocks.RefreshRepository{}

				mRepo.EXPECT().
					UpdateRefreshToken(context.TODO(), VALID_REFRESH_TOKEN, invalidUserId.String()).
					Return(nil, errByRefreshRepo).Once()

				return mRepo
			},

			err: errByRefreshRepo,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mRepo := &mocks.RefreshRepository{}
			if test.mockRepo != nil {
				mRepo = test.mockRepo()
			}

			mConverter := &mocks.Converter{}
			if test.mockConverter != nil {
				mConverter = test.mockConverter()
			}

			rService := NewService(mRepo, nil, mConverter, nil)

			gotOutput, err := rService.UpdateRefreshToken(context.TODO(), test.refreshToken, test.userId)

			if test.err != nil {
				require.EqualError(t, test.err, err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedOutput, gotOutput)
			mock.AssertExpectationsForObjects(t, mRepo, mConverter)
		})
	}
}

func TestService_GetTokenOwner(t *testing.T) {
	tests := []struct {
		testName          string
		refreshToken      string
		mockRepo          func() *mocks.RefreshRepository
		mockUserConverter func() *mocks.UserConverter
		expectedOutput    *models.User
		err               error
	}{
		{
			testName: "Successfully getting token owner",

			refreshToken: VALID_REFRESH_TOKEN,

			mockRepo: func() *mocks.RefreshRepository {
				mRepo := &mocks.RefreshRepository{}

				entityUser := initUserEntity(validUserId, VALID_USER_EMAIL)

				mRepo.EXPECT().
					GetTokenOwner(context.TODO(), VALID_REFRESH_TOKEN).
					Return(entityUser, nil).Once()

				return mRepo
			},

			mockUserConverter: func() *mocks.UserConverter {
				mUserConverter := &mocks.UserConverter{}

				entityUser := initUserEntity(validUserId, VALID_USER_EMAIL)
				userModel := initUser(validUserId.String(), VALID_USER_EMAIL)

				mUserConverter.EXPECT().
					ConvertFromDBEntityToModel(entityUser).
					Return(userModel).Once()

				return mUserConverter
			},

			expectedOutput: initUser(validUserId.String(), VALID_USER_EMAIL),
		},
		{
			testName: "Failed to get token owner error by refresh repository",

			refreshToken: INVALID_REFRESH_TOKEN,

			mockRepo: func() *mocks.RefreshRepository {
				mRepo := &mocks.RefreshRepository{}

				mRepo.EXPECT().
					GetTokenOwner(context.TODO(), INVALID_REFRESH_TOKEN).
					Return(nil, errByRefreshRepoInvalidRefreshToken).Once()

				return mRepo
			},

			err: errByRefreshRepoInvalidRefreshToken,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mRepo := &mocks.RefreshRepository{}

			if test.mockRepo != nil {
				mRepo = test.mockRepo()
			}

			mUserConverter := &mocks.UserConverter{}

			if test.mockUserConverter != nil {
				mUserConverter = test.mockUserConverter()
			}

			rService := NewService(mRepo, nil, nil, mUserConverter)

			gotOutput, err := rService.GetTokenOwner(context.TODO(), test.refreshToken)
			if test.err != nil {
				require.EqualError(t, test.err, err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedOutput, gotOutput)
			mock.AssertExpectationsForObjects(t, mRepo, mUserConverter)
		})
	}
}
