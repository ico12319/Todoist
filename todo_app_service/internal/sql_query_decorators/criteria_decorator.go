package sql_query_decorators

import (
	"context"
	"fmt"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/configuration"
)

// concrete decorator!
type allTodosByCriteriaDecorator struct {
	retriever      SqlQueryRetriever
	condition      string
	conditionValue string
}

func newCriteriaDecorator(retriever SqlQueryRetriever, condition string, conditionValue string) SqlQueryRetriever {
	return &allTodosByCriteriaDecorator{retriever: retriever, condition: condition, conditionValue: conditionValue}
}

func (a *allTodosByCriteriaDecorator) DetermineCorrectSqlQuery(ctx context.Context) string {
	log.C(ctx).Debugf("getting todos by criteria %s in all todos by priority retriever", a.conditionValue)

	currentQuery := a.retriever.DetermineCorrectSqlQuery(ctx)

	addition := determineAddition(currentQuery)

	formattedSuffix := fmt.Sprintf(" %s %s = '%s'", addition, a.condition, a.conditionValue)
	currentQuery += formattedSuffix

	return currentQuery
}
