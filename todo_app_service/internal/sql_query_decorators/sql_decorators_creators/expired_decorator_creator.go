package sql_decorators_creators

import (
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"context"
)

func init() {
	sql_query_decorators.GetDecoratorFactoryInstance().RegisterCreator(&expiredDecoratorCreator{})
}

type expiredDecoratorCreator struct{}

func (e *expiredDecoratorCreator) Create(ctx context.Context, inner sql_query_decorators.SqlQueryRetriever, f sql_query_decorators.Filters) (sql_query_decorators.SqlQueryRetriever, error) {
	log.C(ctx).Info("creating overdue decorator in overdue decorator creator")

	overdue, containsOverdue := f.GetFilters()[constants.OVERDUE]
	if containsOverdue && overdue == constants.TRUE_VALUE {
		inner = sql_query_decorators.NewOverdueDecorator(inner, ">", "NOT", "AND")
	}

	return inner, nil
}

func (e *expiredDecoratorCreator) Priority() int {
	return constants.DEFAULT_PRIORITY
}
