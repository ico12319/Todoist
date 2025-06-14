package list

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	gql "github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/model"
	mocks2 "github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/resolvers/list/mocks"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/handler_models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"io"
	"net/http"
	"testing"
)

func TestResolver_List(t *testing.T) {
	modelList := initModelList()
	convertedGqlModel := initGqlModel()

	tests := []struct {
		testName          string
		listId            string
		mockListConverter func() *mocks2.ListConverter
		mockHttpClient    func() *mocks2.HttpClient
		expectedList      *gql.List
		expectedError     error
	}{
		{
			testName: "successful list fetch by id",
			listId:   listId.String(),
			mockListConverter: func() *mocks2.ListConverter {
				mock := &mocks2.ListConverter{}
				mock.EXPECT().ToGQL(modelList).Return(convertedGqlModel).Once()
				return mock
			},
			mockHttpClient: func() *mocks2.HttpClient {
				mck := &mocks2.HttpClient{}

				bodyBytes, _ := json.Marshal(modelList)

				mockResponse := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(bodyBytes)),
				}
				mck.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == http.MethodGet && req.URL.Path == fmt.Sprintf("/lists/%s", listId) && req.Body == nil
				})).Return(mockResponse, nil).Once()

				return mck
			},
			expectedList:  convertedGqlModel,
			expectedError: nil,
		},
		{
			testName: "unable to fetch list by id, http status code not found return when calling rest api",
			listId:   invalidListId.String(),
			mockHttpClient: func() *mocks2.HttpClient {
				mck := &mocks2.HttpClient{}

				mockResponse := &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}

				mck.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == http.MethodGet && req.URL.Path == fmt.Sprintf("/lists/%s", invalidListId.String()) && req.Body == nil
				})).Return(mockResponse, nil)
				return mck
			},
		},
		{
			testName: "unable to fetch list by id, error when trying to get http response",
			listId:   listId.String(),
			mockHttpClient: func() *mocks2.HttpClient {
				mck := &mocks2.HttpClient{}

				mck.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == http.MethodGet && req.URL.Path == fmt.Sprintf("/lists/%s", listId.String()) && req.Body == nil
				})).Return(nil, assert.AnError)
				return mck
			},
			expectedError: expectedErrorWhenMakingResponseFailsList,
		},
		{
			testName: "unable to fetch list by id, http status code internal server error received when calling rest api",
			listId:   listId.String(),
			mockHttpClient: func() *mocks2.HttpClient {
				mck := &mocks2.HttpClient{}

				mockResponse := &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}

				mck.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == http.MethodGet && req.URL.Path == fmt.Sprintf("/lists/%s", listId.String()) && req.Body == nil
				})).Return(mockResponse, nil)

				return mck
			},
			expectedError: &gqlerror.Error{
				Message:    "Internal error, please try again later.",
				Extensions: map[string]interface{}{"code": "INTERNAL_SERVER_ERROR"},
			},
		},
		{
			testName: "unable to fetch list by id, http status code bad request received when calling rest api",
			listId:   listId.String(),
			mockHttpClient: func() *mocks2.HttpClient {
				mck := &mocks2.HttpClient{}

				mockResponse := &http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}

				mck.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == http.MethodGet && req.URL.Path == fmt.Sprintf("/lists/%s", listId.String()) && req.Body == nil
				})).Return(mockResponse, nil)

				return mck
			},
			expectedError: &gqlerror.Error{
				Message:    "Invalid Request",
				Extensions: map[string]interface{}{"code": "BAD_REQUEST"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockConverter := &mocks2.ListConverter{}
			if test.mockListConverter != nil {
				mockConverter = test.mockListConverter()
			}
			mockClient := &mocks2.HttpClient{}
			if test.mockHttpClient != nil {
				mockClient = test.mockHttpClient()
			}

			listResolver := NewResolver(mockClient, mockConverter, nil, nil, "", nil, nil)
			receivedGqlList, err := listResolver.List(context.Background(), test.listId)
			if test.expectedError != nil {
				require.EqualError(t, test.expectedError, err.Error())
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.expectedList, receivedGqlList)
			mock.AssertExpectationsForObjects(t, mockConverter, mockClient)

		})
	}
}

// not okay?
func TestResolver_DeleteList(t *testing.T) {
	modelList := initModelList()
	gqlList := initGqlModel()
	successfulDeleteListPayload := initSuccessfulGqlDeleteListPayload()

	tests := []struct {
		testName                  string
		listId                    string
		mockConverter             func() *mocks2.ListConverter
		mockHttpClient            func() *mocks2.HttpClient
		expectedDeleteListPayload *gql.DeleteListPayload
		expectedError             error
	}{
		{
			testName: "successfully deleting list",
			listId:   listId.String(),
			mockConverter: func() *mocks2.ListConverter {
				mck := &mocks2.ListConverter{}
				mck.EXPECT().ToGQL(modelList).Return(gqlList).Once()
				mck.EXPECT().FromGQLModelToDeleteListPayload(gqlList, true).Return(successfulDeleteListPayload).Once()
				return mck
			},
			mockHttpClient: func() *mocks2.HttpClient {
				mck := &mocks2.HttpClient{}

				bodyBytes, _ := json.Marshal(modelList)
				mockResponseGetList := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(bodyBytes)),
				}

				mck.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == http.MethodGet && req.URL.Path == fmt.Sprintf("/lists/%s", listId.String()) && req.Body == nil
				})).Return(mockResponseGetList, nil).Once()

				mockResponseDeleteList := &http.Response{
					StatusCode: http.StatusNoContent,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}

				mck.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == http.MethodDelete && req.URL.Path == fmt.Sprintf("/lists/%s", listId.String()) && req.Body == nil
				})).Return(mockResponseDeleteList, nil).Once()

				return mck
			},
			expectedDeleteListPayload: successfulDeleteListPayload,
		},
		{
			testName: "Unable to delete list, http status not found received",
			listId:   invalidListId.String(),
			mockHttpClient: func() *mocks2.HttpClient {
				mck := &mocks2.HttpClient{}

				mockResponseGetList := &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}

				mck.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == http.MethodGet && req.URL.Path == fmt.Sprintf("/lists/%s", invalidListId.String()) && req.Body == nil
				})).Return(mockResponseGetList, nil).Once()
				return mck
			},
			expectedDeleteListPayload: &gql.DeleteListPayload{
				Success: false,
			},
		},
		{
			testName: "Unable to delete list, error when trying to receive http response",
			listId:   listId.String(),
			mockConverter: func() *mocks2.ListConverter {
				mck := &mocks2.ListConverter{}
				mck.EXPECT().ToGQL(modelList).Return(gqlList).Once()
				return mck
			},
			mockHttpClient: func() *mocks2.HttpClient {
				mck := &mocks2.HttpClient{}

				bodyBytes, _ := json.Marshal(modelList)
				mockResponseGetList := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(bodyBytes)),
				}

				mck.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == http.MethodGet && req.URL.Path == fmt.Sprintf("/lists/%s", listId.String()) && req.Body == nil
				})).Return(mockResponseGetList, nil).Once()

				mck.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == http.MethodDelete && req.URL.Path == fmt.Sprintf("/lists/%s", listId.String()) && req.Body == nil
				})).Return(nil, assert.AnError).Once()
				return mck
			},
			expectedDeleteListPayload: &gql.DeleteListPayload{
				Success: false,
			},
			expectedError: fmt.Errorf("failed to fetch list response"),
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockConverter := &mocks2.ListConverter{}
			if test.mockConverter != nil {
				mockConverter = test.mockConverter()
			}

			mockClient := &mocks2.HttpClient{}
			if test.mockHttpClient != nil {
				mockClient = test.mockHttpClient()
			}

			listResolver := NewResolver(mockClient, mockConverter, nil, nil, "", nil, nil)

			receivedPayload, err := listResolver.DeleteList(context.Background(), test.listId)
			if test.expectedError != nil {
				require.EqualError(t, test.expectedError, err.Error())
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.expectedDeleteListPayload, receivedPayload)
			mock.AssertExpectationsForObjects(t, mockConverter, mockClient)
		})
	}
}

func TestResolver_UpdateList(t *testing.T) {

	restModel := &handler_models.UpdateList{
		Name:        &updatedListName,
		Description: &updatedDescription,
	}

	updateInput := gql.UpdateListInput{
		Name:        &updatedListName,
		Description: &updatedDescription,
	}

	modelList := initModelList()
	gqlList := initGqlModel()

	tests := []struct {
		testName        string
		listId          string
		updateListInput gql.UpdateListInput
		expectedGqlList *gql.List
		mockConverter   func() *mocks2.ListConverter
		mockHttpClient  func() *mocks2.HttpClient
		expectedError   error
	}{
		{
			testName:        "Successfully updating list",
			listId:          listId.String(),
			updateListInput: updateInput,
			expectedGqlList: gqlList,
			mockConverter: func() *mocks2.ListConverter {
				mck := &mocks2.ListConverter{}
				mck.EXPECT().UpdateListInputGQLToHandlerModel(updateInput).Return(restModel).Once()

				mck.EXPECT().ToGQL(modelList).Return(gqlList).Once()
				return mck
			},
			mockHttpClient: func() *mocks2.HttpClient {
				mck := &mocks2.HttpClient{}

				modelList.Description = updatedDescription
				modelList.Name = updatedListName

				bodyBytes, _ := json.Marshal(modelList)
				mockResponse := &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(bodyBytes)),
				}

				mck.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == http.MethodPatch && req.URL.Path == fmt.Sprintf("/lists/%s", listId.String())
				})).Return(mockResponse, nil).Once()
				return mck
			},
		},
		{
			testName:        "Unable to update list, error when making http response",
			listId:          listId.String(),
			updateListInput: updateInput,
			mockConverter: func() *mocks2.ListConverter {
				mck := &mocks2.ListConverter{}

				mck.EXPECT().UpdateListInputGQLToHandlerModel(updateInput).Return(restModel).Once()
				return mck
			},
			mockHttpClient: func() *mocks2.HttpClient {
				mck := &mocks2.HttpClient{}

				mck.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == http.MethodPatch && req.URL.Path == fmt.Sprintf("/lists/%s", listId.String())
				})).Return(nil, assert.AnError).Once()
				return mck
			},
			expectedError: expectedErrorWhenMakingResponseFailsList,
		},
		{
			testName:        "Unable to update list, http status internal server received",
			listId:          listId.String(),
			updateListInput: updateInput,
			mockConverter: func() *mocks2.ListConverter {
				mck := &mocks2.ListConverter{}

				mck.EXPECT().UpdateListInputGQLToHandlerModel(updateInput).Return(restModel).Once()
				return mck
			},
			mockHttpClient: func() *mocks2.HttpClient {
				mck := &mocks2.HttpClient{}

				mockResponse := &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}

				mck.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == http.MethodPatch && req.URL.Path == fmt.Sprintf("/lists/%s", listId.String())
				})).Return(mockResponse, nil).Once()
				return mck
			},
			expectedError: gqlInternalServerError,
		},
		{
			testName:        "Unable to update list, http status bad request server received",
			updateListInput: updateInput,
			mockConverter: func() *mocks2.ListConverter {
				mck := &mocks2.ListConverter{}

				mck.EXPECT().UpdateListInputGQLToHandlerModel(updateInput).Return(restModel).Once()
				return mck
			},
			mockHttpClient: func() *mocks2.HttpClient {
				mck := &mocks2.HttpClient{}

				mockResponse := &http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}

				mck.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == http.MethodPatch && req.URL.Path == fmt.Sprintf("/lists/")
				})).Return(mockResponse, nil).Once()
				return mck
			},
			expectedError: gqlBadRequestError,
		},
		{
			testName:        "Unable to update list, http status not found received",
			listId:          invalidListId.String(),
			updateListInput: updateInput,
			mockConverter: func() *mocks2.ListConverter {
				mck := &mocks2.ListConverter{}

				mck.EXPECT().UpdateListInputGQLToHandlerModel(updateInput).Return(restModel).Once()
				return mck
			},
			mockHttpClient: func() *mocks2.HttpClient {
				mck := &mocks2.HttpClient{}

				mockResponse := &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}

				mck.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == http.MethodPatch && req.URL.Path == fmt.Sprintf("/lists/%s", invalidListId.String())
				})).Return(mockResponse, nil).Once()
				return mck
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockConverter := &mocks2.ListConverter{}
			if test.mockConverter != nil {
				mockConverter = test.mockConverter()
			}

			mockClient := &mocks2.HttpClient{}
			if test.mockHttpClient != nil {
				mockClient = test.mockHttpClient()
			}

			listResolver := NewResolver(mockClient, mockConverter, nil, nil, "", nil, nil)

			receivedGqlList, err := listResolver.UpdateList(context.Background(), test.listId, test.updateListInput)
			if test.expectedError != nil {
				require.EqualError(t, test.expectedError, err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedGqlList, receivedGqlList)
			mock.AssertExpectationsForObjects(t, mockConverter, mockConverter)
		})
	}
}

func TestResolver_CreateList(t *testing.T) {
	listInput := gql.CreateListInput{
		Name:        listName,
		Description: listDescription,
	}
	listHandler := &handler_models.CreateList{
		Name:        listName,
		Description: listDescription,
	}

	modelList := initModelList()
	gqlList := initGqlModel()

	tests := []struct {
		testName        string
		input           gql.CreateListInput
		expectedGqlList *gql.List
		mockConverter   func() *mocks2.ListConverter
		mockHttpClient  func() *mocks2.HttpClient
		expectedError   error
	}{
		{
			testName:        "Successfully creating list",
			input:           listInput,
			expectedGqlList: gqlList,
			mockConverter: func() *mocks2.ListConverter {
				mck := &mocks2.ListConverter{}

				mck.EXPECT().CreateListInputGQLToHandlerModel(listInput).Return(listHandler).Once()
				mck.EXPECT().ToGQL(modelList).Return(gqlList).Once()
				return mck
			},
			mockHttpClient: func() *mocks2.HttpClient {
				mck := &mocks2.HttpClient{}

				bodyBytes, _ := json.Marshal(modelList)
				mockResponse := &http.Response{
					StatusCode: http.StatusCreated,
					Body:       io.NopCloser(bytes.NewReader(bodyBytes)),
				}

				mck.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == http.MethodPost && req.URL.Path == "/lists"
				})).Return(mockResponse, nil).Once()
				return mck
			},
		},
		{
			testName: "Unable to create list, error when trying to receive http response",
			input:    listInput,
			mockConverter: func() *mocks2.ListConverter {
				mck := &mocks2.ListConverter{}

				mck.EXPECT().CreateListInputGQLToHandlerModel(listInput).Return(listHandler).Once()
				return mck
			},
			mockHttpClient: func() *mocks2.HttpClient {
				mck := &mocks2.HttpClient{}

				mck.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == http.MethodPost && req.URL.Path == "/lists"
				})).Return(nil, assert.AnError).Once()
				return mck
			},
			expectedError: expectedErrorWhenMakingResponseFailsList,
		},
		{
			testName: "Unable to create list, http status internal server error received",
			input:    listInput,
			mockConverter: func() *mocks2.ListConverter {
				mck := &mocks2.ListConverter{}

				mck.EXPECT().CreateListInputGQLToHandlerModel(listInput).Return(listHandler).Once()
				return mck
			},
			mockHttpClient: func() *mocks2.HttpClient {
				mck := &mocks2.HttpClient{}

				mockResponse := &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}

				mck.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == http.MethodPost && req.URL.Path == "/lists"
				})).Return(mockResponse, nil).Once()
				return mck
			},
			expectedError: gqlInternalServerError,
		},
		{
			testName: "Unable to create list, http status bad request received",
			input:    listInput,
			mockConverter: func() *mocks2.ListConverter {
				mck := &mocks2.ListConverter{}

				mck.EXPECT().CreateListInputGQLToHandlerModel(listInput).Return(listHandler).Once()
				return mck
			},
			mockHttpClient: func() *mocks2.HttpClient {
				mck := &mocks2.HttpClient{}

				modelList.Name = "" // set it to empty string so we can simulate bad request error

				mockResponse := &http.Response{
					StatusCode: http.StatusBadRequest,
					Body:       io.NopCloser(bytes.NewReader(nil)),
				}

				mck.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
					return req.Method == http.MethodPost && req.URL.Path == "/lists"
				})).Return(mockResponse, nil).Once()
				return mck
			},
			expectedError: gqlBadRequestError,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockConverter := &mocks2.ListConverter{}
			if test.mockConverter != nil {
				mockConverter = test.mockConverter()
			}

			mockClient := &mocks2.HttpClient{}
			if test.mockHttpClient != nil {
				mockClient = test.mockHttpClient()
			}

			listResolver := NewResolver(mockClient, mockConverter, nil, nil, "", nil, nil)

			receivedGqlList, err := listResolver.CreateList(context.Background(), test.input)
			if test.expectedError != nil {
				require.EqualError(t, test.expectedError, err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedGqlList, receivedGqlList)
			mock.AssertExpectationsForObjects(t, mockConverter, mockClient)

		})
	}
}

func TestResolver_AddListCollaborator(t *testing.T) {

	tests := []struct {
		testName          string
		input             gql.CollaboratorInput
		mockUserConverter fun
		mockConverter     func() *mocks2.ListConverter
		mockHttpClient    func() *mocks2.HttpClient
		expectedPayload   *gql.CreateCollaboratorPayload
		expectedError     error
	}{
		{
			testName: "Successfully adding collaborator",
			input: gql.CollaboratorInput{
				ListID: listId.String(),
				UserID: userId.String(),
			},
			mockConverter: func() *mocks2.ListConverter {
				mck := &mocks2.ListConverter{}
				returnedHandlerMode := &handler_models.AddCollaborator{
					Id: userId.String(),
				}
				mck.EXPECT().Fr
			},
		},
	}
}
