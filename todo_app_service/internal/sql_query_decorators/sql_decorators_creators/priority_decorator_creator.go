package sql_decorators_creators

import (
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"context"
)

func init() {
	sql_query_decorators.GetDecoratorFactoryInstance().RegisterCreator(&priorityDecoratorCreator{})
}

type priorityDecoratorCreator struct{}

func (*priorityDecoratorCreator) Create(ctx context.Context, inner sql_query_decorators.SqlQueryRetriever, f sql_query_decorators.Filters) (sql_query_decorators.SqlQueryRetriever, error) {
	log.C(ctx).Info("creating priority decorator in priority decorator creator")

	priority, containsPriority := f.GetFilters()[constants.PRIORITY]
	if containsPriority && len(priority) != 0 {
		inner = sql_query_decorators.NewCriteriaDecorator(inner, constants.PRIORITY, priority)
	}

	return inner, nil
}

func (*priorityDecoratorCreator) Priority() int {
	return constants.DEFAULT_PRIORITY
}
