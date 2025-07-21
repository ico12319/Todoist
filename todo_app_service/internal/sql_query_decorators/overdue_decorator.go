package sql_query_decorators

import (
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
	"fmt"
)

type overdueDecorator struct {
	inner        SqlQueryRetriever
	cmpOperator  string
	nullModifier string
	logicalGlue  string
}

func NewOverdueDecorator(inner SqlQueryRetriever, cmpOperator string, nullModifier string, logicalGlue string) *overdueDecorator {
	return &overdueDecorator{
		inner:        inner,
		cmpOperator:  cmpOperator,
		nullModifier: nullModifier,
		logicalGlue:  logicalGlue,
	}
}

func (e *overdueDecorator) DetermineCorrectSqlQuery(ctx context.Context) string {
	log.C(ctx).Info("determining correct sql query in expired decorator")

	currentQuery := e.inner.DetermineCorrectSqlQuery(ctx)
	addition := determineAddition(currentQuery)

	formattedSuffix := fmt.Sprintf(" %s due_date IS %s NULL %s current_date %s due_date", addition, e.nullModifier, e.logicalGlue, e.cmpOperator)

	currentQuery += formattedSuffix
	return currentQuery
}
