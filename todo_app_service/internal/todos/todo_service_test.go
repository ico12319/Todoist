package todos

/*
import (
	"fmt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"internProject/internal/todos/mocks"
	"internProject/pkg/constants"
	"internProject/pkg/models"
	"testing"
	"time"
)

func TestService_CreateTodoRecord(t *testing.T) {
	tests := []struct {
		testName        string
		err             error
		listId          string
		todoName        string
		todoDescription string
		status          constants.TodoStatus
		expectedTodo    *models.Todo
	}{
		{
			"create todo in list with listId list1",
			nil,
			"list1",
			"todo1",
			"todo description",
			"mockStatus",
			&models.Todo{
				Id:          "mockUuid",
				Name:        "todo1",
				Description: "todo description",
				ListId:      "list1",
				Status:      "mockStatus",
				CreatedAt: time.Date(2025, time.January, 1,
					1, 1, 1, 1, time.UTC),
				LastUpdated: time.Date(2025, time.January, 1,
					1, 1, 1, 1, time.UTC),
			},
		},
		{
			"todo creation should result in an error",
			fmt.Errorf("invalid list_id or todo_id provided"),
			"list1",
			"todo1",
			"todo description",
			"mockStatus",
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockUuidGenerator := new(mocks.IUuidGenerator)
			mockTimeGenerator := new(mocks.ITimeGenerator)
			mockRepo := new(mocks.TodoRepo)

			mockUuid := "mockUuid"
			mockTime := time.Date(2025, time.January, 1,
				1, 1, 1, 1, time.UTC)

			mockUuidGenerator.EXPECT().Generate().Return(mockUuid).Once()
			mockTimeGenerator.EXPECT().Now().Return(mockTime).Times(2)

			mockRepo.EXPECT().CreateTodo(mock.AnythingOfType("string"), mock.AnythingOfType("*models.Todo")).
				Return(test.expectedTodo, test.err).Once()

			service := NewService(mockRepo, mockUuidGenerator, mockTimeGenerator)
			todo, err := service.CreateTodoRecord(test.listId, test.testName, test.todoDescription, test.status)

			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.expectedTodo, todo)

			mockRepo.AssertExpectations(t)
			mockUuidGenerator.AssertExpectations(t)
			mockTimeGenerator.AssertExpectations(t)
		})
	}
}

func TestService_DeleteTodoRecord(t *testing.T) {
	tests := []struct {
		testName string
		err      error
		listId   string
		todoId   string
	}{
		{
			"todo deletion that should result in an invalid listId error",
			fmt.Errorf("invalid list_id or todo_id provided"),
			"list1",
			"todo1",
		},
		{
			"todo deletion that should result in an invalid todo error",
			fmt.Errorf("invalid list_id or todo_id provided"),
			"list1",
			"todo1",
		},
		{
			"successful todo deletion where no error is expected",
			nil,
			"list1",
			"todo1",
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockRepo := new(mocks.TodoRepo)

			mockRepo.EXPECT().DeleteTodo(test.listId, test.todoId).Return(test.err).Once()

			service := NewService(mockRepo, nil, nil)
			err := service.DeleteTodoRecord(test.listId, test.todoId)

			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
				return
			}
			require.NoError(t, err)

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetTodoRecord(t *testing.T) {
	tests := []struct {
		testName     string
		err          error
		listId       string
		todoId       string
		expectedTodo *models.Todo
	}{
		{"correct todo returned",
			nil,
			"list1",
			"todo1",
			&models.Todo{
				Id:          "todo1",
				Name:        "todo",
				Description: "description",
				ListId:      "list1",
				Status:      "finished",
				CreatedAt: time.Date(2025, time.January, 1,
					1, 1, 1, 1, time.UTC),
				LastUpdated: time.Date(2025, time.January, 1,
					1, 1, 1, 1, time.UTC),
			},
		},
		{
			"should result in an invalid listId error",
			fmt.Errorf("invalid list_id or todo_id provided"),
			"list1",
			"todo1",
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockRepo := new(mocks.TodoRepo)

			service := NewService(mockRepo, nil, nil)
			mockRepo.EXPECT().GetTodo(test.listId, test.todoId).Return(test.expectedTodo, test.err).Once()

			returnedTodo, err := service.GetTodoRecord(test.listId, test.todoId)
			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.expectedTodo, returnedTodo)
			mockRepo.AssertExpectations(t)
		})
	}

}

func TestService_GetTodoRecords(t *testing.T) {
	tests := []struct {
		testName      string
		err           error
		listId        string
		expectedTodos []*models.Todo
	}{
		{
			"should return correct array of todos",
			nil,
			"list1",
			[]*models.Todo{
				&models.Todo{
					Id:          "todo1",
					Name:        "todo",
					Description: "description",
					ListId:      "list1",
					Status:      "finished",
					CreatedAt: time.Date(2025, time.January, 1,
						1, 1, 1, 1, time.UTC),
					LastUpdated: time.Date(2025, time.January, 1,
						1, 1, 1, 1, time.UTC),
				},
			},
		},
		{
			"should result in an error of type non-existing list-id",
			fmt.Errorf("invalid list_id or todo_id provided"),
			"list1",
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockRepo := new(mocks.TodoRepo)

			mockRepo.EXPECT().GetTodosByListId(test.listId).Return(test.expectedTodos, test.err).Once()
			service := NewService(mockRepo, nil, nil)

			returnedTodos, err := service.GetTodoRecords(test.listId)

			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.expectedTodos, returnedTodos)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_UpdateTodoRecord(t *testing.T) {
	mockTime := time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC)
	wantedUpdatedTime := time.Date(2024, time.January, 1, 12, 1, 1, 1, time.UTC)

	dummyTodo := &models.Todo{
		Id:          "todo1",
		Name:        "todo",
		Description: "description",
		ListId:      "list1",
		Status:      "status",
		CreatedAt:   mockTime,
		LastUpdated: mockTime,
	}

	tests := []struct {
		testName     string
		err          error
		listId       string
		todoId       string
		status       constants.TodoStatus
		returnedTodo *models.Todo
	}{
		{
			"should result in no error and the object " +
				"status and last updated fields should be updated correctly " +
				"when passed to the repo function",
			nil,
			"list1",
			"todo1",
			"finished",
			dummyTodo,
		},
		{
			"should result in an list id does not exist error",
			fmt.Errorf("invalid list_id or todo_id provided"),
			"list1",
			"todo1",
			"in progress",
			nil,
		},
		{
			"should result in an todo id does not exist error",
			fmt.Errorf("invalid list_id or todo_id provided"),
			"list1",
			"todo1",
			"finished",
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockRepo := new(mocks.TodoRepo)
			mockTimeGen := new(mocks.ITimeGenerator)

			mockRepo.EXPECT().GetTodo(test.listId, test.todoId).Return(test.returnedTodo, test.err).Once()

			if test.err == nil {
				mockTimeGen.EXPECT().Now().Return(wantedUpdatedTime).Once()

				mockRepo.EXPECT().UpdateTodo(test.listId, test.todoId, mock.MatchedBy(func(arg *models.Todo) bool {
					return arg.LastUpdated.Equal(wantedUpdatedTime) && arg.Status == test.status
				})).Return(test.err).Once()
			}

			service := NewService(mockRepo, nil, mockTimeGen)
			err := service.UpdateTodoRecord(test.listId, test.todoId, test.status)
			if test.err != nil {
				require.EqualError(t, err, test.err.Error())
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockTimeGen.AssertExpectations(t)
		})
	}
}
*/
