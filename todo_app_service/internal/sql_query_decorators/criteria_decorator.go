package sql_query_decorators

import (
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
	"fmt"
)

// concrete decorator!
type criteriaDecorator struct {
	retriever      SqlQueryRetriever
	condition      string
	conditionValue string
}

func NewCriteriaDecorator(retriever SqlQueryRetriever, condition string, conditionValue string) *criteriaDecorator {
	return &criteriaDecorator{retriever: retriever, condition: condition, conditionValue: conditionValue}
}

func (a *criteriaDecorator) DetermineCorrectSqlQuery(ctx context.Context) string {
	log.C(ctx).Debugf("getting todos by criteria %s in all todos by priority retriever", a.conditionValue)

	currentQuery := a.retriever.DetermineCorrectSqlQuery(ctx)

	addition := determineAddition(currentQuery)

	formattedSuffix := fmt.Sprintf(" %s %s = %s", addition, a.condition, a.conditionValue)
	currentQuery += formattedSuffix

	return currentQuery
}
