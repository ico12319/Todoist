package sql_query_decorators

import (
	"fmt"
	"github.com/google/uuid"
)

var (
	cursorValue                                       = uuid.New()
	expectedCursorAdditionWhenInnerNotContainingWhere = fmt.Sprintf("base query WHERE id > '%s'", cursorValue)
	expectedCursorAdditionWhenInnerContainsWhere      = fmt.Sprintf("base query WHERE AND id > '%s'", cursorValue)
)

const (
	invalidLimitValue               = "invalid limit value"
	limitValueString                = "3"
	limitValue                      = 3
	baseQueryString                 = "base query"
	baseQueryStringWithWhere        = "base query WHERE"
	limitQueryString                = baseQueryString + " LIMIT 3"
	priorityCondition               = "priority"
	statusCondition                 = "status"
	priorityValue                   = "priority value"
	statusValue                     = "status value"
	decoratedPriorityQueryWithWhere = "base query WHERE priority = 'priority value'"
	decoratedStatusQueryWithWhere   = "base query WHERE status = 'status value'"

	decoratedPriorityQueryWithAnd = "base query WHERE AND priority = 'priority value'"
	decoratedStatusQueryWithAnd   = "base query WHERE AND status = 'status value'"
)

func initStatusAndLimitDecorators(baseDecorator SqlQueryRetriever) (SqlQueryRetriever, SqlQueryRetriever) {
	statusDecorator := newCriteriaDecorator(baseDecorator, "status", statusValue)
	expectedOuter := newLimitDecorator(statusDecorator, limitValue)

	return statusDecorator, expectedOuter
}
