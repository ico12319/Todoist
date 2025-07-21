package sql_decorators_creators

import (
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"context"
)

func init() {
	sql_query_decorators.GetDecoratorFactoryInstance().RegisterCreator(&activeDecoratorCreator{})
}

type activeDecoratorCreator struct{}

func (a *activeDecoratorCreator) Create(ctx context.Context, inner sql_query_decorators.SqlQueryRetriever, f sql_query_decorators.Filters) (sql_query_decorators.SqlQueryRetriever, error) {
	log.C(ctx).Info("creating overdue decorator in overdue decorator creator")

	overdue, containsOverdue := f.GetFilters()[constants.OVERDUE]
	if containsOverdue && overdue == constants.FALSE_VALUE {
		inner = sql_query_decorators.NewOverdueDecorator(inner, "<", "", "OR")
	}

	return inner, nil
}

func (e *activeDecoratorCreator) Priority() int {
	return constants.DEFAULT_PRIORITY
}
