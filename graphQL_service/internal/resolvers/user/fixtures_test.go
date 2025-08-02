package user

import (
	gql "Todo-List/internProject/graphQL_service/graph/model"
	"Todo-List/internProject/graphQL_service/internal/gql_constants"
	"Todo-List/internProject/graphQL_service/internal/resolvers/user/mocks"
	"Todo-List/internProject/graphQL_service/internal/url_decorators"
	"Todo-List/internProject/graphQL_service/internal/url_decorators/url_filters"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"Todo-List/internProject/todo_app_service/pkg/models"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"io"
	"net/http"
	"testing"
)

var (
	invalidLimit                 = "invalid limit"
	limit                        = "200"
	url                          = "test.com"
	urlGeneratedByLimitDecorator = url + "?limit=200"
)

var (
	userId1 = "id1"
	userId2 = "id2"

	userEmail1 = "test1@email.com"
	userEmail2 = "test2@gmail.com"

	userRole1 = "admin"
	userRole2 = "reader"

	todoId1 = "todo id1"
	todoId2 = "todo id2"

	todoName1 = "todo name1"
	todoName2 = "todo name2"

	todoDescription1 = "todo description1"
	todoDescription2 = "todo description2"

	todoPriority1 = "medium"
	todoPriority2 = "low"

	listId1 = "list id1"
	listId2 = "list id2"

	listName1 = "list name1"
	listName2 = "list name2"

	listDescription1 = "list description1"
	listDescription2 = "list description2"
)

var (
	errWhenTryingToGetHttpResponse                   = errors.New("error when trying to get http response")
	errInternalServerErrorCodeReceivedInHttpResponse = &gqlerror.Error{
		Message:    "Internal error, please try again later.",
		Extensions: map[string]interface{}{"code": "INTERNAL_SERVER_ERROR"},
	}

	errWhenTryingToDecodeJSON = errors.New("EOF")
)

var (
	mockUsers                           = initMockUsers(userId1, userId1)
	mockGqlUsers                        = initMockGqlUsers()
	mockUser                            = initMockUser()
	mockGqlUser                         = initMockGqlUser()
	mockSuccessfulDeleteUserPayload     = initMockSuccessfulGqlDeleteUserPayload()
	mockSuccessfulDeletePayloads        = initSuccessfulDeleteUserPayloads()
	mockUnSuccessfulDeleteUserPayload   = initMockUnSuccessfulGqlDeleteUserPayload()
	mockTodosAssignedToUser             = initMockTodosAssignedToUser()
	mockGqlTodosAssignedToUser          = initMockGqlTodosAssignedToUser()
	mockTodoPage                        = initTodoPage()
	mockListsWhereUserParticipatesIn    = initListsWhereUserParticipatesIn()
	mockGqlListsWhereUserParticipatesIn = initGqlListsWhereUserParticipatesIn()
	mockListPage                        = initListPage()
)

func initMockUsers(userId1 string, userId2 string) []*models.User {
	return []*models.User{
		{
			Id:    userId1,
			Email: userEmail1,
			Role:  constants.UserRole(userRole1),
		},
		{
			Id:    userId2,
			Email: userEmail2,
			Role:  constants.UserRole(userRole2),
		},
	}
}

func initMockTodosAssignedToUser() []*models.Todo {
	return []*models.Todo{
		{
			Id:          todoId1,
			Name:        todoName1,
			Description: todoDescription1,
			ListId:      listId1,
			Priority:    constants.Priority(todoPriority1),
			AssignedTo:  &userId1,
		},
		{
			Id:          todoId2,
			Name:        todoName2,
			Description: todoDescription2,
			ListId:      listId2,
			Priority:    constants.Priority(todoPriority2),
			AssignedTo:  &userId1,
		},
	}
}

func initMockGqlTodosAssignedToUser() []*gql.Todo {
	return []*gql.Todo{
		{
			ID:          todoId1,
			Name:        todoName1,
			Description: todoDescription1,
			Priority:    gql.Priority(todoPriority1),
		},
		{
			ID:          todoId2,
			Name:        todoName2,
			Description: todoDescription2,
			Priority:    gql.Priority(todoPriority2),
		},
	}
}

func initMockGqlUsers() []*gql.User {
	gqlUserRole1 := gql.UserRole(userRole1)
	gqlUserRole2 := gql.UserRole(userRole2)

	return []*gql.User{
		{
			ID:    userId1,
			Email: userEmail1,
			Role:  &gqlUserRole1,
		},
		{
			ID:    userId2,
			Email: userEmail2,
			Role:  &gqlUserRole2,
		},
	}
}

func initSuccessfulDeleteUserPayloads() []*gql.DeleteUserPayload {
	role1 := gql.UserRole(userRole1)
	role2 := gql.UserRole(userRole2)

	return []*gql.DeleteUserPayload{
		{
			ID:    userId1,
			Email: &userEmail1,
			Role:  &role1,
		},
		{
			ID:    userId2,
			Email: &userEmail2,
			Role:  &role2,
		},
	}
}

func initMockUser() *models.User {
	return &models.User{
		Id:    userId1,
		Email: userEmail1,
		Role:  constants.UserRole(userRole1),
	}
}

func initMockGqlUser() *gql.User {
	role := gql.UserRole(userRole1)

	return &gql.User{
		ID:    userId1,
		Email: userEmail1,
		Role:  &role,
	}
}

func initMockSuccessfulGqlDeleteUserPayload() *gql.DeleteUserPayload {
	role := gql.UserRole(userRole1)

	return &gql.DeleteUserPayload{
		Success: true,
		ID:      userId1,
		Email:   &userEmail1,
		Role:    &role,
	}
}

func initMockUnSuccessfulGqlDeleteUserPayload() *gql.DeleteUserPayload {
	return &gql.DeleteUserPayload{
		Success: false,
	}
}

func initTodoPage() *gql.TodoPage {
	return &gql.TodoPage{
		Data:       mockGqlTodosAssignedToUser,
		TotalCount: 2,
		PageInfo: &gql.PageInfo{
			StartCursor: mockGqlTodosAssignedToUser[0].ID,
			EndCursor:   mockGqlTodosAssignedToUser[len(mockGqlTodosAssignedToUser)-1].ID,
		},
	}
}

func initListPage() *gql.ListPage {
	return &gql.ListPage{
		Data:       mockGqlListsWhereUserParticipatesIn,
		TotalCount: 2,
		PageInfo: &gql.PageInfo{
			StartCursor: mockGqlListsWhereUserParticipatesIn[0].ID,
			EndCursor:   mockGqlListsWhereUserParticipatesIn[len(mockGqlListsWhereUserParticipatesIn)-1].ID,
		},
	}
}

func initListsWhereUserParticipatesIn() []*models.List {
	return []*models.List{
		{
			Id:          listId1,
			Name:        listName1,
			Description: listDescription1,
			Owner:       userId1,
		},
		{
			Id:          listId2,
			Name:        listName2,
			Description: listDescription2,
			Owner:       userId1,
		},
	}
}

func initGqlListsWhereUserParticipatesIn() []*gql.List {
	return []*gql.List{
		{
			ID:          listId1,
			Name:        listName1,
			Description: listDescription1,
		},
		{
			ID:          listId2,
			Name:        listName2,
			Description: listDescription2,
		},
	}
}

func getHttpResponseWithCorrectUser(t *testing.T) *mocks.HttpService {
	t.Helper()

	mck := &mocks.HttpService{}
	formattedSuffix := fmt.Sprintf("/%s", userId1)
	mockUrl := url + gql_constants.USER_PATH + formattedSuffix

	mockUserBytes, err := json.Marshal(mockUser)
	require.NoError(t, err)

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(mockUserBytes)),
	}

	mck.EXPECT().
		GetHttpResponseWithAuthHeader(mock.Anything, http.MethodGet, mockUrl, nil).
		Return(resp, nil).
		Once()

	return mck
}

func getHttpResponseWithCorrectUsers(t *testing.T) *mocks.HttpService {
	t.Helper()

	mck := &mocks.HttpService{}
	mockUrl := getUserPath()

	mockUsersBytes, err := json.Marshal(mockUsers)
	require.NoError(t, err)

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(mockUsersBytes)),
	}

	mck.EXPECT().
		GetHttpResponseWithAuthHeader(mock.Anything, http.MethodGet, mockUrl, nil).
		Return(resp, nil).
		Once()

	return mck
}

func getUserConverterManyToGql(t *testing.T) *mocks.UserConverter {
	t.Helper()

	mck := &mocks.UserConverter{}
	mck.EXPECT().
		ManyToGQL(mockUsers).
		Return(mockGqlUsers).
		Once()

	return mck
}

func getHttpResponseWithNilBodyAndStatusNotFound() *mocks.HttpService {
	mck := &mocks.HttpService{}
	formattedSuffix := fmt.Sprintf("/%s", userId2)
	mockUrl := url + gql_constants.USER_PATH + formattedSuffix

	resp := &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(bytes.NewReader(nil)),
	}

	mck.EXPECT().
		GetHttpResponseWithAuthHeader(mock.Anything, http.MethodGet, mockUrl, nil).
		Return(resp, nil).
		Once()

	return mck
}

func getHttpResponseWithStatusInternalServerErrorWhenGettingUser(url string) *mocks.HttpService {
	mck := &mocks.HttpService{}

	resp := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(bytes.NewReader(nil)),
	}

	mck.EXPECT().
		GetHttpResponseWithAuthHeader(mock.Anything, http.MethodGet, url, nil).
		Return(resp, nil).
		Once()

	return mck
}

func getUrlFactoryMockWithEmptyTodoFilters() *mocks.UrlDecoratorFactory {
	mck := &mocks.UrlDecoratorFactory{}

	mockUrl := getPathForGettingTodosAssignedToUser()
	baseDecorator := url_decorators.NewBaseUrl(mockUrl)

	mck.EXPECT().
		CreateUrlDecorator(mock.Anything, mockUrl, &url_filters.TodoFilters{}).
		Return(baseDecorator).
		Once()

	return mck
}

func getUrlFactoryMockWithEmptyUserFilters() *mocks.UrlDecoratorFactory {
	mck := &mocks.UrlDecoratorFactory{}

	mockUrl := getPathForGettingListsWhereUserParticipates()
	baseDecorator := url_decorators.NewBaseUrl(mockUrl)

	mck.EXPECT().
		CreateUrlDecorator(mock.Anything, mockUrl, &url_filters.UserFilters{}).
		Return(baseDecorator).
		Once()

	return mck
}

func getCorrectlyConvertedUserFromModelToGql() *mocks.UserConverter {
	mck := &mocks.UserConverter{}

	mck.EXPECT().
		ToGQL(mockUser).
		Return(mockGqlUser).
		Once()

	return mck
}

func getPathForGettingTodosAssignedToUser() string {
	formattedSuffix := fmt.Sprintf("/%s", userId1)
	mockUrl := gql_constants.USER_PATH + formattedSuffix + gql_constants.TODO_PATH

	return mockUrl

}

func getPathForGettingListsWhereUserParticipates() string {
	formattedSuffix := fmt.Sprintf("/%s", userId1)

	mockUrl := gql_constants.USER_PATH + formattedSuffix + gql_constants.LISTS_PATH

	return mockUrl

}

func getUserPath() string {
	return url + gql_constants.USER_PATH
}
