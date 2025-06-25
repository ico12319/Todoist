package sql_query_decorators

import (
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators/mocks"
	"context"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAllTodosByCriteriaDecorator_DetermineCorrectSqlQuery(t *testing.T) {
	tests := []struct {
		testName                string
		mockInnerDecorator      func() *mocks.SqlQueryRetriever
		condition               string
		conditionValue          string
		expectedDecoratedString string
	}{
		{
			testName: "Successfully creating decorated query with priority",
			mockInnerDecorator: func() *mocks.SqlQueryRetriever {
				mock := &mocks.SqlQueryRetriever{}
				mock.EXPECT().DetermineCorrectSqlQuery(context.TODO()).Return(baseQueryString).Once()

				return mock
			},
			condition:               priorityCondition,
			conditionValue:          priorityValue,
			expectedDecoratedString: decoratedPriorityQueryWithWhere,
		},

		{
			testName: "Successfully creating decorated query with priority containing WHERE",
			mockInnerDecorator: func() *mocks.SqlQueryRetriever {
				mock := &mocks.SqlQueryRetriever{}
				mock.EXPECT().DetermineCorrectSqlQuery(context.TODO()).Return(baseQueryStringWithWhere).Once()

				return mock
			},
			condition:               priorityCondition,
			conditionValue:          priorityValue,
			expectedDecoratedString: decoratedPriorityQueryWithAnd,
		},

		{
			testName: "Successfully creating decorated query status",
			mockInnerDecorator: func() *mocks.SqlQueryRetriever {
				mock := &mocks.SqlQueryRetriever{}
				mock.EXPECT().DetermineCorrectSqlQuery(context.TODO()).Return(baseQueryString).Once()

				return mock
			},
			condition:               statusCondition,
			conditionValue:          statusValue,
			expectedDecoratedString: decoratedStatusQueryWithWhere,
		},

		{
			testName: "Successfully creating decorated query status containing Where",
			mockInnerDecorator: func() *mocks.SqlQueryRetriever {
				mock := &mocks.SqlQueryRetriever{}
				mock.EXPECT().DetermineCorrectSqlQuery(context.TODO()).Return(baseQueryStringWithWhere).Once()

				return mock
			},
			condition:               statusCondition,
			conditionValue:          statusValue,
			expectedDecoratedString: decoratedStatusQueryWithAnd,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockDecorator := &mocks.SqlQueryRetriever{}
			if test.mockInnerDecorator != nil {
				mockDecorator = test.mockInnerDecorator()
			}

			cDecorator := newCriteriaDecorator(mockDecorator, test.condition, test.conditionValue)
			receivedDecoratedValue := cDecorator.DetermineCorrectSqlQuery(context.TODO())

			require.Equal(t, test.expectedDecoratedString, receivedDecoratedValue)
			mock.AssertExpectationsForObjects(t, mockDecorator)
		})
	}
}
