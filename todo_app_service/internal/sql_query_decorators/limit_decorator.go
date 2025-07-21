package sql_query_decorators

import (
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
	"fmt"
)

type limitDecorator struct {
	retriever SqlQueryRetriever
	limit     int
}

func NewLimitDecorator(retriever SqlQueryRetriever, limit int) *limitDecorator {
	return &limitDecorator{retriever: retriever, limit: limit}
}

func (l *limitDecorator) DetermineCorrectSqlQuery(ctx context.Context) string {
	log.C(ctx).Infof("getting todos in limit retriever")

	currentQuery := l.retriever.DetermineCorrectSqlQuery(ctx)

	formattedSuffix := fmt.Sprintf(" LIMIT %d", l.limit)

	currentQuery += formattedSuffix
	return currentQuery
}
