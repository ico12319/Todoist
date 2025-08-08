package sql_query_decorators

import (
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
	"fmt"
)

type orderByDecorator struct {
	inner           SqlQueryRetriever
	sortingCriteria string
}

func NewOrderByDecorator(inner SqlQueryRetriever, sortingCriteria string) *orderByDecorator {
	return &orderByDecorator{
		inner:           inner,
		sortingCriteria: sortingCriteria,
	}
}

func (o *orderByDecorator) DetermineCorrectSqlQuery(ctx context.Context) string {
	log.C(ctx).Info("determining correct sql query in expired decorator")

	currentQuery := o.inner.DetermineCorrectSqlQuery(ctx)

	formattedSuffix := fmt.Sprintf(" ORDER BY id %s ", o.sortingCriteria)

	currentQuery += formattedSuffix
	return currentQuery
}
