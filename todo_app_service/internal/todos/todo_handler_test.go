package todos

/*
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"internProject/internal/handler_models"
	"internProject/internal/todos/mocks"
	"internProject/middlewares"
	"internProject/pkg/constants"
	"internProject/pkg/models"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandler_HandleGetTodo(t *testing.T) {
	mockTime := time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC)
	const (
		listId = "list1"
		todoId = "todo1"
	)

	tests := []struct {
		testName             string
		isReturningError     bool
		wantedHttpCode       int
		errReturnedByService error
		expectedTodo         *models.Todo
	}{
		{
			"should result in an error with httpStatusInternalServerError",
			true,
			http.StatusInternalServerError,
			fmt.Errorf("list with id %s does not exist", "list1"),
			nil,
		},
		{
			"should encode httpStatusCodeOK and correct todo JSON representation",
			false,
			http.StatusOK,
			nil,
			&models.Todo{
				Id:          todoId,
				Name:        "todo",
				Description: "description",
				ListId:      listId,
				CreatedAt:   mockTime,
				LastUpdated: mockTime,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockService := new(mocks.IService)
			handler := NewHandler(mockService)
			mockService.EXPECT().GetTodoRecord(mock.AnythingOfType("string"), mock.AnythingOfType("string")).
				Return(test.expectedTodo, test.errReturnedByService).Once()

			rr := httptest.NewRecorder()
			ctx := context.Background()
			ctx = context.WithValue(ctx, middlewares.ListId, listId)
			ctx = context.WithValue(ctx, middlewares.TodoId, todoId)

			req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil)

			handler.HandleGetTodo(rr, req)
			if test.isReturningError {
				errorMatchHelper(t, rr, test.errReturnedByService)
			} else {
				var encodedTodo models.Todo
				err := json.Unmarshal(rr.Body.Bytes(), &encodedTodo)
				require.NoError(t, err)

				require.Equal(t, *test.expectedTodo, encodedTodo)
			}
			require.Equal(t, test.wantedHttpCode, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_HandleGetTodos(t *testing.T) {

	mockTime := time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC)
	const (
		listId = "list1"
		todoId = "todo1"
	)

	tests := []struct {
		testName             string
		isReturningError     bool
		wantedHttpCode       int
		errReturnedByService error
		expectedTodoArray    []*models.Todo
	}{
		{
			"should encode httpStatusInternalServerError and error",
			true,
			http.StatusInternalServerError,
			fmt.Errorf("non-existing list-id"),
			nil,
		},
		{
			"should encode httpStatusOK and correct JSON Todo array representation",
			false,
			http.StatusOK,
			nil,
			[]*models.Todo{
				&models.Todo{
					Id:          todoId,
					Name:        "todo",
					Description: "description",
					ListId:      listId,
					CreatedAt:   mockTime,
					LastUpdated: mockTime,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockService := new(mocks.IService)
			handler := NewHandler(mockService)

			mockService.EXPECT().GetTodoRecords(mock.AnythingOfType("string")).
				Return(test.expectedTodoArray, test.errReturnedByService).Once()
			rr := httptest.NewRecorder()
			ctx := context.Background()

			ctx = context.WithValue(ctx, middlewares.ListId, listId)
			req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
			handler.HandleGetTodos(rr, req)

			if test.isReturningError {
				errorMatchHelper(t, rr, test.errReturnedByService)
			} else {
				var encodedTodoArray []*models.Todo
				err := json.Unmarshal(rr.Body.Bytes(), &encodedTodoArray)
				require.NoError(t, err)

				require.Equal(t, test.expectedTodoArray, encodedTodoArray)
			}
			require.Equal(t, test.wantedHttpCode, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_HandleTodoCreation(t *testing.T) {
	mockTime := time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC)
	const (
		listId = "list1"
		todoId = "todo1"
	)
	requestBodyHandlerTodo := handler_models.Todo{
		Name:        "name",
		Description: "description",
		Status:      "finished",
	}

	tests := []struct {
		testName                string
		isReturningError        bool
		err                     error
		requestBody             interface{}
		returnedTodoFromService *models.Todo
		expectedHttpCode        int
	}{
		{
			"service layer returns error and this should encode httpStatusInternalServerError",
			true,
			fmt.Errorf("non-existing list id %s", listId),
			requestBodyHandlerTodo,
			nil,
			http.StatusInternalServerError,
		},
		{
			"service layer successfully creates todo and encodes it correctly and encodes httpStatusCreated",
			false,
			nil,
			requestBodyHandlerTodo,
			&models.Todo{
				Id:          todoId,
				Name:        "name",
				Description: "description",
				ListId:      listId,
				Status:      "finished",
				CreatedAt:   mockTime,
				LastUpdated: mockTime,
			},
			http.StatusCreated,
		},
		{
			"handler encodes httpStatusBadRequest because of invalid todo provided in the request body",
			true,
			fmt.Errorf(constants.INVALID_REQUEST_BODY),
			"invalid JSON",
			nil,
			http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockService := new(mocks.IService)

			handler := NewHandler(mockService)
			rr := httptest.NewRecorder()
			ctx := context.Background()
			ctx = context.WithValue(ctx, middlewares.ListId, listId)

			var buff bytes.Buffer
			switch v := test.requestBody.(type) {
			case string:
				buff.WriteString(v)
			default:
				require.NoError(t, json.NewEncoder(&buff).Encode(v))
				mockService.EXPECT().CreateTodoRecord(listId, mock.AnythingOfType("string"),
					mock.AnythingOfType("string"), mock.AnythingOfType("constants.TodoStatus")).
					Return(test.returnedTodoFromService, test.err).Once()

			}
			require.NoError(t, json.NewEncoder(&buff).Encode(test.requestBody))

			req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/", &buff)

			handler.HandleTodoCreation(rr, req)
			if test.isReturningError {
				errorMatchHelper(t, rr, test.err)
			} else {
				var createdTodo models.Todo
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &createdTodo))
				require.Equal(t, test.returnedTodoFromService, &createdTodo)
			}
			require.Equal(t, test.expectedHttpCode, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_HandleDeleteTodo(t *testing.T) {

	const (
		listId = "list1"
		todoId = "todo1"
	)

	tests := []struct {
		testName         string
		isReturningError bool
		err              error
		expectedHttpCode int
	}{
		{
			"service layer returns error because of invalid listId" +
				"and handler should encode httpStatusInternalServerError",
			true,
			fmt.Errorf("list with id %s does not exist", listId),
			http.StatusInternalServerError,
		},
		{
			"service layer returns error because of invalid todoId" +
				"and handler should encode httpStatusInternalServerError",
			true,
			fmt.Errorf("todo with id %s does not exist", todoId),
			http.StatusInternalServerError,
		},
		{
			"service layer successfully deletes todo and handler" +
				"should encode httpStatusNoContent",
			false,
			nil,
			http.StatusNoContent,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockService := new(mocks.IService)
			mockService.EXPECT().DeleteTodoRecord(mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(test.err).Once()
			handler := NewHandler(mockService)

			rr := httptest.NewRecorder()
			ctx := context.Background()
			ctx = context.WithValue(ctx, middlewares.ListId, listId)
			ctx = context.WithValue(ctx, middlewares.TodoId, todoId)

			req := httptest.NewRequestWithContext(ctx, http.MethodDelete, "/", nil)
			handler.HandleDeleteTodo(rr, req)

			if test.isReturningError {
				errorMatchHelper(t, rr, test.err)
			}
			require.Equal(t, test.expectedHttpCode, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_HandleUpdateTodoRecord(t *testing.T) {
	const (
		listId = "list1"
		todoId = "todo1"
	)

	todoRequestBody := struct {
		Status constants.TodoStatus `json:"status"`
	}{}

	tests := []struct {
		testName         string
		isReturningError bool
		err              error
		requestBody      interface{}
		expectedHttpCode int
	}{
		{
			"service layer returns error invalid listId and " +
				"handler encodes httpStatusInternalServerError",
			true,
			fmt.Errorf("list with id %s does not exist", listId),
			todoRequestBody,
			http.StatusInternalServerError,
		},
		{
			"service layer returns error invalid todoId and " +
				"handler encodes httpStatusInternalServerError",
			true,
			fmt.Errorf("todo with id %s does not exist", todoId),
			todoRequestBody,
			http.StatusInternalServerError,
		},
		{
			"service layer successfully updates todo and handler encodes " +
				"httpStatusOK",
			false,
			nil,
			todoRequestBody,
			http.StatusOK,
		},
		{
			"incorrect request body, encodes httpStatusBadRequest",
			true,
			fmt.Errorf(constants.INVALID_REQUEST_BODY),
			"invalid JSON",
			http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockService := new(mocks.IService)
			handler := NewHandler(mockService)

			rr := httptest.NewRecorder()
			ctx := context.Background()
			ctx = context.WithValue(ctx, middlewares.ListId, listId)
			ctx = context.WithValue(ctx, middlewares.TodoId, todoId)

			var buff bytes.Buffer
			switch v := test.requestBody.(type) {
			case string:
				buff.WriteString(v)
			default:
				require.NoError(t, json.NewEncoder(&buff).Encode(v))
				mockService.EXPECT().UpdateTodoRecord(mock.AnythingOfType("string"),
					mock.AnythingOfType("string"), mock.AnythingOfType("constants.TodoStatus")).Return(test.err).Once()
			}

			req := httptest.NewRequestWithContext(ctx, http.MethodPatch, "/", &buff)
			handler.HandleUpdateTodoRecord(rr, req)

			if test.isReturningError {
				errorMatchHelper(t, rr, test.err)
			}
			require.Equal(t, test.expectedHttpCode, rr.Code)
			mockService.AssertExpectations(t)
		})
	}
}
*/
