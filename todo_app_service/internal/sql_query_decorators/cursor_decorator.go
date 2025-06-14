package sql_query_decorators

import (
	"context"
	"fmt"
	log "github.com/I763039/Todo-List/internProject/todo_app_service/pkg/configuration"
)

type cursorDecorator struct {
	inner  SqlQueryRetriever
	cursor string
}

func newCursorDecorator(inner SqlQueryRetriever, cursor string) SqlQueryRetriever {
	return &cursorDecorator{inner: inner, cursor: cursor}
}

func (c *cursorDecorator) DetermineCorrectSqlQuery(ctx context.Context) string {
	log.C(ctx).Info("determining correct sql query in cursor retriever")

	currentQuery := c.inner.DetermineCorrectSqlQuery(ctx)
	addition := determineAddition(currentQuery)

	formattedSuffix := fmt.Sprintf(" %s id > '%s'", addition, c.cursor)
	currentQuery += formattedSuffix

	return currentQuery
}
