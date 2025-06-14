package sql_query_decorators

import (
	"context"
	"fmt"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/configuration"
)

type limitDecorator struct {
	retriever SqlQueryRetriever
	limit     int
}

func newLimitDecorator(retriever SqlQueryRetriever, limit int) SqlQueryRetriever {
	return &limitDecorator{retriever: retriever, limit: limit}
}

func (l *limitDecorator) DetermineCorrectSqlQuery(ctx context.Context) string {
	log.C(ctx).Infof("getting todos in limit retriever")

	currentQuery := l.retriever.DetermineCorrectSqlQuery(ctx)

	formattedSuffix := fmt.Sprintf(" LIMIT %d", l.limit)

	currentQuery += formattedSuffix
	return currentQuery

}
