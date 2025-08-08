package user

import (
	gql "Todo-List/internProject/graphQL_service/graph/model"
	gql_constants2 "Todo-List/internProject/graphQL_service/internal/gql_constants"
	"Todo-List/internProject/graphQL_service/internal/resolvers/user/mocks"
	"Todo-List/internProject/graphQL_service/internal/url_decorators"
	"Todo-List/internProject/graphQL_service/internal/url_decorators/url_filters"
	"Todo-List/internProject/internal/gql_constants"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
)

func TestResolver_Users(t *testing.T) {
	tests := []struct {
		testName                string
		filters                 *url_filters.BaseFilters
		urlDecoratorFactoryMock func() *mocks.UrlDecoratorFactory
		httpServiceMock         func() *mocks.HttpService
		userConverterMock       func() *mocks.UserConverter
		err                     error
		expectedUsersPage       *gql.UserPage
	}{
		{
			testName: "Successfully getting users",

			filters: &url_filters.BaseFilters{
				Limit: &limit,
			},

			urlDecoratorFactoryMock: func() *mocks.UrlDecoratorFactory {
				mck := &mocks.UrlDecoratorFactory{}
				baseDecorator := url_decorators.NewBaseUrl("")

				returnedDecorator := url_decorators.NewCriteriaDecorator(baseDecorator, gql_constants.LIMIT, limit)

				mck.EXPECT().
					CreateUrlDecorator(mock.Anything, gql_constants2.USER_PATH, &url_filters.BaseFilters{
						Limit: &limit,
					}).
					Return(returnedDecorator).
					Once()

				return mck
			},

			httpServiceMock: func() *mocks.HttpService {
				mck := &mocks.HttpService{}
				usersBytes, err := json.Marshal(mockUsers)
				require.NoError(t, err)

				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(usersBytes)),
				}

				mck.EXPECT().
					GetHttpResponseWithAuthHeader(mock.Anything, http.MethodGet, url+"?limit=200", nil).
					Return(resp, nil)

				return mck
			},

			userConverterMock: func() *mocks.UserConverter {
				mck := &mocks.UserConverter{}

				mck.EXPECT().
					ManyToGQL(mockUsers).
					Return(mockGqlUsers).
					Once()

				return mck
			},

			expectedUsersPage: &gql.UserPage{
				TotalCount: 2,
				Data:       mockGqlUsers,
				PageInfo: &gql.PageInfo{
					StartCursor: mockGqlUsers[0].ID,
					EndCursor:   mockGqlUsers[len(mockGqlUsers)-1].ID,
				},
			},
		},

		{
			testName: "Failed to get users, error when trying to get http response",

			filters: &url_filters.BaseFilters{},

			urlDecoratorFactoryMock: func() *mocks.UrlDecoratorFactory {
				mck := &mocks.UrlDecoratorFactory{}
				baseDecorator := url_decorators.NewBaseUrl("")

				mck.EXPECT().
					CreateUrlDecorator(mock.Anything, gql_constants2.USER_PATH, &url_filters.BaseFilters{}).
					Return(baseDecorator).
					Once()

				return mck
			},

			httpServiceMock: func() *mocks.HttpService {
				mck := &mocks.HttpService{}

				mck.EXPECT().
					GetHttpResponseWithAuthHeader(mock.Anything, http.MethodGet, url, nil).
					Return(nil, errWhenTryingToGetHttpResponse).Once()

				return mck
			},

			err: errWhenTryingToGetHttpResponse,
		},

		{
			testName: "Failed to get users, http status code 500 in http response",

			filters: &url_filters.BaseFilters{},

			urlDecoratorFactoryMock: func() *mocks.UrlDecoratorFactory {
				mck := &mocks.UrlDecoratorFactory{}
				baseDecorator := url_decorators.NewBaseUrl("")

				mck.EXPECT().
					CreateUrlDecorator(mock.Anything, gql_constants2.USER_PATH, &url_filters.BaseFilters{}).
					Return(baseDecorator).
					Once()

				return mck
			},

			httpServiceMock: func() *mocks.HttpService {
				mck := &mocks.HttpService{}

				resp := &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}

				mck.EXPECT().
					GetHttpResponseWithAuthHeader(mock.Anything, http.MethodGet, url, nil).
					Return(resp, nil)

				return mck
			},

			err: errInternalServerErrorCodeReceivedInHttpResponse,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockHttpService := &mocks.HttpService{}
			if test.httpServiceMock != nil {
				mockHttpService = test.httpServiceMock()
			}

			mockUserConverter := &mocks.UserConverter{}
			if test.userConverterMock != nil {
				mockUserConverter = test.userConverterMock()
			}

			factoryMock := &mocks.UrlDecoratorFactory{}
			if test.urlDecoratorFactoryMock != nil {
				factoryMock = test.urlDecoratorFactoryMock()
			}

			userResolver := NewResolver(mockUserConverter, nil, nil, url, factoryMock, mockHttpService)

			receivedUserPage, err := userResolver.Users(context.TODO(), test.filters)
			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedUsersPage, receivedUserPage)
			mock.AssertExpectationsForObjects(t, mockHttpService, mockUserConverter, factoryMock)
		})
	}
}

func TestResolver_User(t *testing.T) {
	tests := []struct {
		testName          string
		id                string
		httpServiceMock   func() *mocks.HttpService
		userConverterMock func() *mocks.UserConverter
		err               error
		expectedUser      *gql.User
	}{
		{
			testName: "Successfully getting user",

			id: userId1,

			httpServiceMock: func() *mocks.HttpService {
				return getHttpResponseWithCorrectUser(t)
			},

			userConverterMock: func() *mocks.UserConverter {
				return getCorrectlyConvertedUserFromModelToGql()
			},

			expectedUser: mockGqlUser,
		},

		{
			testName: "Failed to get user, error when trying to get http response",

			id: userId1,

			httpServiceMock: func() *mocks.HttpService {
				mck := &mocks.HttpService{}
				formattedSuffix := fmt.Sprintf("/%s", userId1)
				mockUrl := url + gql_constants2.USER_PATH + formattedSuffix

				mck.EXPECT().
					GetHttpResponseWithAuthHeader(mock.Anything, http.MethodGet, mockUrl, nil).
					Return(nil, errWhenTryingToGetHttpResponse).
					Once()

				return mck
			},

			err: errWhenTryingToGetHttpResponse,
		},

		{
			testName: "Failed to get user, http status code not found received",

			id: userId2,

			httpServiceMock: func() *mocks.HttpService {
				return getHttpResponseWithNilBodyAndStatusNotFound()
			},
		},

		{
			testName: "Failed to get user, http status code internal server error received",

			id: userId1,

			httpServiceMock: func() *mocks.HttpService {
				formattedSuffix := fmt.Sprintf("/%s", userId1)
				mockUrl := url + gql_constants2.USER_PATH + formattedSuffix

				return getHttpResponseWithStatusInternalServerErrorWhenGettingUser(mockUrl)
			},

			err: errInternalServerErrorCodeReceivedInHttpResponse,
		},

		{
			testName: "Failed t get user, error when trying to decode response body",

			id: userId2,

			httpServiceMock: func() *mocks.HttpService {
				mck := &mocks.HttpService{}

				formattedSuffix := fmt.Sprintf("/%s", userId2)
				mockUrl := url + gql_constants2.USER_PATH + formattedSuffix

				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}

				mck.EXPECT().
					GetHttpResponseWithAuthHeader(context.TODO(), http.MethodGet, mockUrl, nil).
					Return(resp, nil).
					Once()

				return mck
			},

			err: errWhenTryingToDecodeJSON,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			httpServiceMock := &mocks.HttpService{}

			if test.httpServiceMock != nil {
				httpServiceMock = test.httpServiceMock()
			}

			userConverterMock := &mocks.UserConverter{}
			if test.userConverterMock != nil {
				userConverterMock = test.userConverterMock()
			}

			userResolver := NewResolver(userConverterMock, nil, nil, url, nil, httpServiceMock)

			receivedUser, err := userResolver.User(context.TODO(), test.id)
			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedUser, receivedUser)
			mock.AssertExpectationsForObjects(t, httpServiceMock, userConverterMock)
		})
	}
}

func TestResolver_DeleteUser(t *testing.T) {
	tests := []struct {
		testName                  string
		id                        string
		httpServiceMock           func() *mocks.HttpService
		userConverterMock         func() *mocks.UserConverter
		err                       error
		expectedDeleteUserPayload *gql.DeleteUserPayload
	}{
		{
			testName: "Successfully deleting user",

			id: userId1,

			httpServiceMock: func() *mocks.HttpService {
				mck := &mocks.HttpService{}
				mck = getHttpResponseWithCorrectUser(t)

				formattedSuffix := fmt.Sprintf("/%s", userId1)
				mockUrl := url + gql_constants2.USER_PATH + formattedSuffix

				resp := &http.Response{
					StatusCode: http.StatusNoContent,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}

				mck.EXPECT().
					GetHttpResponseWithAuthHeader(mock.Anything, http.MethodDelete, mockUrl, nil).
					Return(resp, nil).
					Once()

				return mck
			},

			userConverterMock: func() *mocks.UserConverter {
				mck := &mocks.UserConverter{}
				mck = getCorrectlyConvertedUserFromModelToGql()

				mck.EXPECT().
					FromGQLToDeleteUserPayload(mockGqlUser, true).
					Return(mockSuccessfulDeleteUserPayload).
					Once()

				return mck
			},

			expectedDeleteUserPayload: mockSuccessfulDeleteUserPayload,
		},

		{
			testName: "Failed to delete user, user that is trying to be deleted is not found",

			id: userId2,

			httpServiceMock: func() *mocks.HttpService {
				mck := &mocks.HttpService{}
				mck = getHttpResponseWithNilBodyAndStatusNotFound()

				return mck
			},

			expectedDeleteUserPayload: mockUnSuccessfulDeleteUserPayload,
		},

		{
			testName: "Failed to delete user, error when trying to get user",

			id: userId1,

			httpServiceMock: func() *mocks.HttpService {
				formattedSuffix := fmt.Sprintf("/%s", userId1)
				mockUrl := url + gql_constants2.USER_PATH + formattedSuffix

				return getHttpResponseWithStatusInternalServerErrorWhenGettingUser(mockUrl)
			},

			err: errInternalServerErrorCodeReceivedInHttpResponse,

			expectedDeleteUserPayload: mockUnSuccessfulDeleteUserPayload,
		},

		{
			testName: "Failed to delete user, http status internal server error received when trying to make DELETE request",

			id: userId1,

			httpServiceMock: func() *mocks.HttpService {
				mck := &mocks.HttpService{}
				mck = getHttpResponseWithCorrectUser(t)

				formattedSuffix := fmt.Sprintf("/%s", userId1)
				mockUrl := url + gql_constants2.USER_PATH + formattedSuffix

				resp := &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}

				mck.EXPECT().
					GetHttpResponseWithAuthHeader(mock.Anything, http.MethodDelete, mockUrl, nil).
					Return(resp, nil).
					Once()

				return mck
			},

			userConverterMock: func() *mocks.UserConverter {
				mck := &mocks.UserConverter{}
				mck = getCorrectlyConvertedUserFromModelToGql()

				return mck
			},

			err: errInternalServerErrorCodeReceivedInHttpResponse,

			expectedDeleteUserPayload: mockUnSuccessfulDeleteUserPayload,
		},

		{
			testName: "Failed to delete user, error when trying to get http response when making DELETE request",

			id: userId1,

			httpServiceMock: func() *mocks.HttpService {
				mck := &mocks.HttpService{}
				mck = getHttpResponseWithCorrectUser(t)

				formattedSuffix := fmt.Sprintf("/%s", userId1)
				mockUrl := url + gql_constants2.USER_PATH + formattedSuffix

				mck.EXPECT().
					GetHttpResponseWithAuthHeader(mock.Anything, http.MethodDelete, mockUrl, nil).
					Return(nil, errWhenTryingToGetHttpResponse).
					Once()

				return mck
			},

			userConverterMock: func() *mocks.UserConverter {
				mck := &mocks.UserConverter{}
				mck = getCorrectlyConvertedUserFromModelToGql()

				return mck
			},

			err: errWhenTryingToGetHttpResponse,

			expectedDeleteUserPayload: mockUnSuccessfulDeleteUserPayload,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			httpServiceMock := &mocks.HttpService{}
			if test.httpServiceMock != nil {
				httpServiceMock = test.httpServiceMock()
			}

			userConverterMock := &mocks.UserConverter{}
			if test.userConverterMock != nil {
				userConverterMock = test.userConverterMock()
			}

			userResolver := NewResolver(userConverterMock, nil, nil, url, nil, httpServiceMock)

			receivedDeleteUserPayload, err := userResolver.DeleteUser(context.TODO(), test.id)
			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedDeleteUserPayload, receivedDeleteUserPayload)
			mock.AssertExpectationsForObjects(t, httpServiceMock, userConverterMock)
		})
	}
}

func TestResolver_AssignedTo(t *testing.T) {
	tests := []struct {
		testName             string
		obj                  *gql.User
		filters              *url_filters.TodoFilters
		decoratorFactoryMock func() *mocks.UrlDecoratorFactory
		httpServiceMock      func() *mocks.HttpService
		todoConverterMock    func() *mocks.TodoConverter
		err                  error
		expectedTodoPage     *gql.TodoPage
	}{
		{
			testName: "Successfully getting todos assigned to  user",

			obj: &gql.User{
				ID: userId1,
			},

			filters: &url_filters.TodoFilters{
				BaseFilters: url_filters.BaseFilters{
					Limit: &limit,
				},
			},

			decoratorFactoryMock: func() *mocks.UrlDecoratorFactory {
				mck := &mocks.UrlDecoratorFactory{}
				baseDecorator := url_decorators.NewBaseUrl(getPathForGettingTodosAssignedToUser())

				mockUrl := getPathForGettingTodosAssignedToUser()

				returnedDecorator := url_decorators.NewCriteriaDecorator(baseDecorator, gql_constants.LIMIT, limit)

				mck.EXPECT().
					CreateUrlDecorator(mock.Anything, mockUrl, &url_filters.TodoFilters{
						BaseFilters: url_filters.BaseFilters{
							Limit: &limit,
						},
					}).
					Return(returnedDecorator).
					Once()

				return mck
			},

			httpServiceMock: func() *mocks.HttpService {
				mck := &mocks.HttpService{}

				mockUrl := getPathForGettingTodosAssignedToUser()
				mockUrl += "?limit=200"

				todosAssignedToUserBytes, err := json.Marshal(mockTodosAssignedToUser)
				require.NoError(t, err)

				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(todosAssignedToUserBytes)),
				}

				mck.EXPECT().
					GetHttpResponseWithAuthHeader(mock.Anything, http.MethodGet, mockUrl, nil).
					Return(resp, nil).
					Once()

				return mck
			},

			todoConverterMock: func() *mocks.TodoConverter {
				mck := &mocks.TodoConverter{}

				mck.EXPECT().
					ManyToGQL(mockTodosAssignedToUser).
					Return(mockGqlTodosAssignedToUser).
					Once()

				return mck
			},

			expectedTodoPage: mockTodoPage,
		},

		{
			testName: "Failed to get todos assigned to user, http status internal server error received",

			obj: mockGqlUser,

			filters: &url_filters.TodoFilters{},

			decoratorFactoryMock: func() *mocks.UrlDecoratorFactory {
				return getUrlFactoryMockWithEmptyTodoFilters()
			},

			httpServiceMock: func() *mocks.HttpService {
				mck := &mocks.HttpService{}

				mockUrl := getPathForGettingTodosAssignedToUser()

				resp := &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}

				mck.EXPECT().
					GetHttpResponseWithAuthHeader(mock.Anything, http.MethodGet, mockUrl, nil).
					Return(resp, nil).
					Once()

				return mck
			},

			err: errInternalServerErrorCodeReceivedInHttpResponse,
		},

		{
			testName: "Failed to get todos assigned to user, error when trying to get http response",

			obj: mockGqlUser,

			filters: &url_filters.TodoFilters{},

			decoratorFactoryMock: func() *mocks.UrlDecoratorFactory {
				return getUrlFactoryMockWithEmptyTodoFilters()
			},

			httpServiceMock: func() *mocks.HttpService {
				mck := &mocks.HttpService{}

				mockUrl := getPathForGettingTodosAssignedToUser()

				mck.EXPECT().
					GetHttpResponseWithAuthHeader(mock.Anything, http.MethodGet, mockUrl, nil).
					Return(nil, errWhenTryingToGetHttpResponse).
					Once()

				return mck
			},

			err: errWhenTryingToGetHttpResponse,
		},

		{
			testName: "Failed to get todos assigned to user, http status not found",

			obj: mockGqlUser,

			filters: &url_filters.TodoFilters{},

			decoratorFactoryMock: func() *mocks.UrlDecoratorFactory {
				return getUrlFactoryMockWithEmptyTodoFilters()
			},

			httpServiceMock: func() *mocks.HttpService {
				mck := &mocks.HttpService{}

				mockUrl := getPathForGettingTodosAssignedToUser()

				resp := &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}

				mck.EXPECT().
					GetHttpResponseWithAuthHeader(mock.Anything, http.MethodGet, mockUrl, nil).
					Return(resp, nil).
					Once()

				return mck
			},
		},

		{
			testName: "Failed to get todos assigned to user, failed to decode response body",

			obj: mockGqlUser,

			filters: &url_filters.TodoFilters{},

			decoratorFactoryMock: func() *mocks.UrlDecoratorFactory {
				return getUrlFactoryMockWithEmptyTodoFilters()
			},

			httpServiceMock: func() *mocks.HttpService {
				mck := &mocks.HttpService{}

				mockUrl := getPathForGettingTodosAssignedToUser()

				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}

				mck.EXPECT().
					GetHttpResponseWithAuthHeader(mock.Anything, http.MethodGet, mockUrl, nil).
					Return(resp, nil).
					Once()

				return mck
			},

			err: errWhenTryingToDecodeJSON,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			decoratorFactoryMock := &mocks.UrlDecoratorFactory{}
			if test.decoratorFactoryMock != nil {
				decoratorFactoryMock = test.decoratorFactoryMock()
			}

			httpServiceMock := &mocks.HttpService{}
			if test.httpServiceMock != nil {
				httpServiceMock = test.httpServiceMock()
			}

			todoConverterMock := &mocks.TodoConverter{}
			if test.todoConverterMock != nil {
				todoConverterMock = test.todoConverterMock()
			}

			userResolver := NewResolver(nil, nil, todoConverterMock, "", decoratorFactoryMock, httpServiceMock)

			receivedTodoPage, err := userResolver.AssignedTo(context.TODO(), test.obj, test.filters)
			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedTodoPage, receivedTodoPage)
			mock.AssertExpectationsForObjects(t, decoratorFactoryMock, httpServiceMock, todoConverterMock)
		})
	}
}

func TestResolver_ParticipateIn(t *testing.T) {
	tests := []struct {
		testName             string
		obj                  *gql.User
		filters              *url_filters.UserFilters
		decoratorFactoryMock func() *mocks.UrlDecoratorFactory
		httpServiceMock      func() *mocks.HttpService
		listConverterMock    func() *mocks.ListConverter
		err                  error
		expectedListPage     *gql.ListPage
	}{
		{
			testName: "Successfully getting lists where user participates in",

			obj: mockGqlUser,

			filters: &url_filters.UserFilters{
				BaseFilters: url_filters.BaseFilters{
					Limit: &limit,
				},
			},

			decoratorFactoryMock: func() *mocks.UrlDecoratorFactory {
				mck := &mocks.UrlDecoratorFactory{}
				mockUrl := getPathForGettingListsWhereUserParticipates()

				baseDecorator := url_decorators.NewBaseUrl(mockUrl)

				returnedDecorator := url_decorators.NewCriteriaDecorator(baseDecorator, gql_constants.LIMIT, limit)

				mck.EXPECT().
					CreateUrlDecorator(mock.Anything, mockUrl, &url_filters.UserFilters{
						BaseFilters: url_filters.BaseFilters{
							Limit: &limit,
						},
					}).
					Return(returnedDecorator).
					Once()

				return mck
			},

			httpServiceMock: func() *mocks.HttpService {
				mck := &mocks.HttpService{}

				mockUrl := getPathForGettingListsWhereUserParticipates()
				mockUrl += "?limit=200"

				listsWhereUserParticipatesBytes, err := json.Marshal(mockListsWhereUserParticipatesIn)
				require.NoError(t, err)

				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(listsWhereUserParticipatesBytes)),
				}

				mck.EXPECT().
					GetHttpResponseWithAuthHeader(mock.Anything, http.MethodGet, mockUrl, nil).
					Return(resp, nil).
					Once()

				return mck
			},

			listConverterMock: func() *mocks.ListConverter {
				mck := &mocks.ListConverter{}

				mck.EXPECT().
					ManyToGQL(mockListsWhereUserParticipatesIn).
					Return(mockGqlListsWhereUserParticipatesIn).
					Once()

				return mck
			},

			expectedListPage: mockListPage,
		},

		{
			testName: "Failed to get lists where user participates, http status internal server error received",

			obj: mockGqlUser,

			filters: &url_filters.UserFilters{},

			decoratorFactoryMock: func() *mocks.UrlDecoratorFactory {
				return getUrlFactoryMockWithEmptyUserFilters()
			},

			httpServiceMock: func() *mocks.HttpService {
				mck := &mocks.HttpService{}

				mockUrl := getPathForGettingListsWhereUserParticipates()

				resp := &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}

				mck.EXPECT().
					GetHttpResponseWithAuthHeader(mock.Anything, http.MethodGet, mockUrl, nil).
					Return(resp, nil).
					Once()

				return mck
			},

			err: errInternalServerErrorCodeReceivedInHttpResponse,
		},

		{
			testName: "Failed to get lists where user participates in, error when trying to get http response",

			obj: mockGqlUser,

			filters: &url_filters.UserFilters{},

			decoratorFactoryMock: func() *mocks.UrlDecoratorFactory {
				return getUrlFactoryMockWithEmptyUserFilters()
			},

			httpServiceMock: func() *mocks.HttpService {
				mck := &mocks.HttpService{}

				mockUrl := getPathForGettingListsWhereUserParticipates()

				mck.EXPECT().
					GetHttpResponseWithAuthHeader(mock.Anything, http.MethodGet, mockUrl, nil).
					Return(nil, errWhenTryingToGetHttpResponse).
					Once()

				return mck
			},

			err: errWhenTryingToGetHttpResponse,
		},

		{
			testName: "Failed to get lists where user participates in, http status not found",

			obj: mockGqlUser,

			filters: &url_filters.UserFilters{},

			decoratorFactoryMock: func() *mocks.UrlDecoratorFactory {
				return getUrlFactoryMockWithEmptyUserFilters()
			},

			httpServiceMock: func() *mocks.HttpService {
				mck := &mocks.HttpService{}

				mockUrl := getPathForGettingListsWhereUserParticipates()

				resp := &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}

				mck.EXPECT().
					GetHttpResponseWithAuthHeader(mock.Anything, http.MethodGet, mockUrl, nil).
					Return(resp, nil).
					Once()

				return mck
			},
		},

		{
			testName: "Failed to get lists where user participates in, failed to decode response body",

			obj: mockGqlUser,

			filters: &url_filters.UserFilters{},

			decoratorFactoryMock: func() *mocks.UrlDecoratorFactory {
				return getUrlFactoryMockWithEmptyUserFilters()
			},

			httpServiceMock: func() *mocks.HttpService {
				mck := &mocks.HttpService{}

				mockUrl := getPathForGettingListsWhereUserParticipates()

				resp := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}

				mck.EXPECT().
					GetHttpResponseWithAuthHeader(mock.Anything, http.MethodGet, mockUrl, nil).
					Return(resp, nil).
					Once()

				return mck
			},

			err: errWhenTryingToDecodeJSON,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			decoratorFactoryMock := &mocks.UrlDecoratorFactory{}
			if test.decoratorFactoryMock != nil {
				decoratorFactoryMock = test.decoratorFactoryMock()
			}

			httpServiceMock := &mocks.HttpService{}
			if test.httpServiceMock != nil {
				httpServiceMock = test.httpServiceMock()
			}

			listConverterMock := &mocks.ListConverter{}
			if test.listConverterMock != nil {
				listConverterMock = test.listConverterMock()
			}

			userResolver := NewResolver(nil, listConverterMock, nil, "", decoratorFactoryMock, httpServiceMock)

			receivedListPage, err := userResolver.ParticipateIn(context.TODO(), test.obj, test.filters)
			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedListPage, receivedListPage)
			mock.AssertExpectationsForObjects(t, decoratorFactoryMock, httpServiceMock, listConverterMock)
		})
	}
}

func TestResolver_DeleteUsers(t *testing.T) {
	tests := []struct {
		testName                  string
		httpServiceMock           func() *mocks.HttpService
		userConverterMock         func() *mocks.UserConverter
		err                       error
		expectedDeleteUserPayload []*gql.DeleteUserPayload
	}{
		{
			testName: "Successfully deleting users",

			httpServiceMock: func() *mocks.HttpService {
				mck := &mocks.HttpService{}
				mck = getHttpResponseWithCorrectUsers(t)
				mockUrl := getUserPath()

				resp := &http.Response{
					StatusCode: http.StatusNoContent,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}

				mck.EXPECT().
					GetHttpResponseWithAuthHeader(mock.Anything, http.MethodDelete, mockUrl, nil).
					Return(resp, nil).
					Once()

				return mck
			},

			userConverterMock: func() *mocks.UserConverter {
				mck := &mocks.UserConverter{}
				mck = getUserConverterManyToGql(t)

				mck.EXPECT().
					ManyFromGQLToDeleteUserPayload(mockGqlUsers, true).
					Return(mockSuccessfulDeletePayloads).Once()

				return mck
			},

			expectedDeleteUserPayload: mockSuccessfulDeletePayloads,
		},

		{
			testName: "Failed to delete users, http status internal server error status code received",

			httpServiceMock: func() *mocks.HttpService {
				mockUrl := url + gql_constants2.USER_PATH

				return getHttpResponseWithStatusInternalServerErrorWhenGettingUser(mockUrl)
			},

			err: errInternalServerErrorCodeReceivedInHttpResponse,

			expectedDeleteUserPayload: []*gql.DeleteUserPayload{
				{
					Success: false,
				},
			},
		},

		{
			testName: "Failed to delete users, http status internal server error received when calling with DELETE method",

			httpServiceMock: func() *mocks.HttpService {
				mck := &mocks.HttpService{}
				mck = getHttpResponseWithCorrectUsers(t)

				mockUrl := url + gql_constants2.USER_PATH
				mck.EXPECT().
					GetHttpResponseWithAuthHeader(mock.Anything, http.MethodDelete, mockUrl, nil).
					Return(nil, errWhenTryingToGetHttpResponse).
					Once()

				return mck
			},

			userConverterMock: func() *mocks.UserConverter {
				return getUserConverterManyToGql(t)
			},

			err: errWhenTryingToGetHttpResponse,

			expectedDeleteUserPayload: []*gql.DeleteUserPayload{
				{
					Success: false,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockHttpService := &mocks.HttpService{}
			if test.httpServiceMock != nil {
				mockHttpService = test.httpServiceMock()
			}

			mockUserConverter := &mocks.UserConverter{}
			if test.userConverterMock != nil {
				mockUserConverter = test.userConverterMock()
			}

			userResolver := NewResolver(mockUserConverter, nil, nil, url, nil, mockHttpService)
			receivedDeleteUsersPayload, err := userResolver.DeleteUsers(context.TODO())
			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedDeleteUserPayload, receivedDeleteUsersPayload)
			mock.AssertExpectationsForObjects(t, mockUserConverter, mockHttpService)
		})
	}
}
