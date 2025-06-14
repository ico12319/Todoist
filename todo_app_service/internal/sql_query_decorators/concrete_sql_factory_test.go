package sql_query_decorators

import (
	"context"
	"fmt"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/sql_query_decorators/filters"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/sql_query_decorators/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)ж

func TestConcreteDecoratorFactory_CreateTodoQueryDecorator(t *testing.T) {
	baseDecorator := newBaseQuery(baseQueryString)
	baseFilters := &filters.BaseFilters{
		Limit:  limitValueString,
		Cursor: cursorValue.String(),
	}

	statusDecorator, expectedOuter := initStatusAndLimitDecorators(baseDecorator)

	tests := []struct {
		testName          string
		mockCommonFactory func() *mocks.CommonFactory
		filters           *filters.TodoFilters
		expectedDecorator SqlQueryRetriever
		expectedError     error
	}{
		{
			testName:      "Unable to create todo decorator, nil filters",
			expectedError: fmt.Errorf("invalid todo filters provided"),
		},
		{
			testName: "Successfully returning correct status decorator enhanced with common factory fields",
			mockCommonFactory: func() *mocks.CommonFactory {
				mock := &mocks.CommonFactory{}
				mock.EXPECT().CreateCommonDecorator(context.TODO(), statusDecorator, baseFilters).Return(expectedOuter, nil).Once()

				return mock
			},
			filters: &filters.TodoFilters{
				Status:      statusValue,
				BaseFilters: *baseFilters,
			},
			expectedDecorator: expectedOuter,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockFactory := &mocks.CommonFactory{}
			if test.mockCommonFactory != nil {
				mockFactory = test.mockCommonFactory()
			}

			cFactory := NewConcreteQueryDecoratorFactory(mockFactory)
			receivedDecorator, err := cFactory.CreateTodoQueryDecorator(context.TODO(), baseQueryString, test.filters)
			if test.expectedError != nil {
				require.EqualError(t, test.expectedError, err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedDecorator, receivedDecorator)
			mock.AssertExpectationsForObjects(t, mockFactory)
		})
	}
}
