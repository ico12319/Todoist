package sql_query_decorators

import (
	"context"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/sql_query_decorators/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCursorDecorator_DetermineCorrectSqlQuery(t *testing.T) {
	tests := []struct {
		testName                string
		mockInnerDecorator      func() *mocks.SqlQueryRetriever
		expectedDecoratedString string
	}{
		{
			testName: "Returns correctly decorated string where the inner decorator does not contain WHERE",
			mockInnerDecorator: func() *mocks.SqlQueryRetriever {
				mock := &mocks.SqlQueryRetriever{}
				mock.EXPECT().DetermineCorrectSqlQuery(context.TODO()).Return(baseQueryString).Once()

				return mock
			},
			expectedDecoratedString: expectedCursorAdditionWhenInnerNotContainingWhere,
		},
		{
			testName: "Returns correctly decorated string where the inner decorator contains WHERE",
			mockInnerDecorator: func() *mocks.SqlQueryRetriever {
				mock := &mocks.SqlQueryRetriever{}
				mock.EXPECT().DetermineCorrectSqlQuery(context.TODO()).Return(baseQueryStringWithWhere).Once()

				return mock
			},
			expectedDecoratedString: expectedCursorAdditionWhenInnerContainsWhere,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockInnerDecorator := &mocks.SqlQueryRetriever{}
			if test.mockInnerDecorator != nil {
				mockInnerDecorator = test.mockInnerDecorator()
			}

			cDecorator := newCursorDecorator(mockInnerDecorator, cursorValue.String())

			receivedCursorDecoratedQuery := cDecorator.DetermineCorrectSqlQuery(context.TODO())
			require.Equal(t, test.expectedDecoratedString, receivedCursorDecoratedQuery)

			mock.AssertExpectationsForObjects(t, mockInnerDecorator)

		})
	}
}
