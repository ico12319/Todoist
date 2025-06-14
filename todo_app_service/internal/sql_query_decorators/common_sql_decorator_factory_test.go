package sql_query_decorators

import (
	"context"
	"fmt"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/sql_query_decorators/filters"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCommonDecoratorFactory_CreateCommonDecorator(t *testing.T) {
	passedBaseDecorator := newBaseQuery(baseQueryString)

	tests := []struct {
		testName          string
		passedDecorator   SqlQueryRetriever
		filters           *filters.BaseFilters
		expectedDecorator SqlQueryRetriever
		expectedError     error
	}{
		{
			testName:        "Successfully returning only cursor decorator",
			passedDecorator: passedBaseDecorator,
			filters: &filters.BaseFilters{
				Cursor: cursorValue.String(),
			},
			expectedDecorator: newCursorDecorator(passedBaseDecorator, cursorValue.String()),
		},
		{
			testName:        "Successfully returning only limit decorator",
			passedDecorator: passedBaseDecorator,
			filters: &filters.BaseFilters{
				Limit: limitValueString,
			},
			expectedDecorator: newLimitDecorator(passedBaseDecorator, limitValue),
		},
		{
			testName:        "Returning error due to invalid limit param passed",
			passedDecorator: passedBaseDecorator,
			filters: &filters.BaseFilters{
				Limit: invalidLimitValue,
			},
			expectedError: fmt.Errorf("invalid limit: %s", invalidLimitValue),
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			cFactory := NewCommonFactory()
			gotDecorator, err := cFactory.CreateCommonDecorator(context.TODO(), test.passedDecorator, test.filters)

			if test.expectedError != nil {
				require.EqualError(t, test.expectedError, err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedDecorator, gotDecorator)
		})
	}
}
