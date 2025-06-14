package sql_query_decorators

import (
	"context"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/sql_query_decorators/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLimitDecorator_DetermineCorrectSqlQuery(t *testing.T) {
	tests := []struct {
		testName                string
		mockInnerDecorator      func() *mocks.SqlQueryRetriever
		expectedDecoratedString string
	}{
		{
			testName: "Returns correctly decorated query with limit",
			mockInnerDecorator: func() *mocks.SqlQueryRetriever {
				mock := &mocks.SqlQueryRetriever{}
				mock.EXPECT().DetermineCorrectSqlQuery(context.TODO()).Return(baseQueryString).Once()

				return mock
			},
			expectedDecoratedString: limitQueryString,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockDecorator := &mocks.SqlQueryRetriever{}
			if test.mockInnerDecorator != nil {
				mockDecorator = test.mockInnerDecorator()
			}

			lDecorator := newLimitDecorator(mockDecorator, limitValue)
			limitDecoratedQuery := lDecorator.DetermineCorrectSqlQuery(context.TODO())

			require.Equal(t, test.expectedDecoratedString, limitDecoratedQuery)
			mock.AssertExpectationsForObjects(t, mockDecorator)
		})
	}
}
