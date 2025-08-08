package helpers

import (
	gql "Todo-List/internProject/graphQL_service/graph/model"
	"Todo-List/internProject/graphQL_service/internal/gql_constants"
	"Todo-List/internProject/graphQL_service/internal/url_decorators/url_filters"
	"strconv"
)

func preparePointer(ptr *int32) *string {
	var preparedPtr *string
	if ptr != nil {
		value := strconv.Itoa(int(*ptr))
		preparedPtr = &value
	}

	return preparedPtr
}

func ExtractLastAndFirstPointers(first *int32, last *int32) (*string, *string) {
	preparedFirst := preparePointer(first)
	preparedLast := preparePointer(last)

	if preparedFirst == nil && preparedLast == nil {
		defaultLimitValue := gql_constants.DEFAULT_LIMIT_VALUE
		preparedFirst = &defaultLimitValue
	}

	return preparedFirst, preparedLast
}

func InitBaseFilters(first *int32, after *string, last *int32, before *string) *url_filters.BaseFilters {
	preparedFirst, preparedLast := ExtractLastAndFirstPointers(first, last)

	return &url_filters.BaseFilters{
		First:  preparedFirst,
		Last:   preparedLast,
		After:  after,
		Before: before,
	}
}

func InitTodoFilters(first *int32, after *string, last *int32, before *string, tFilters *gql.TodosFilterInput) *url_filters.TodoFilters {
	preparedFirst, preparedLast := ExtractLastAndFirstPointers(first, last)

	return &url_filters.TodoFilters{
		BaseFilters: url_filters.BaseFilters{
			First:  preparedFirst,
			Last:   preparedLast,
			After:  after,
			Before: before,
		},
		TodoFilters: tFilters,
	}
}
