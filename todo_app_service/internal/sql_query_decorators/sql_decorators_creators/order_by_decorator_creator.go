package sql_decorators_creators

import (
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"context"
)

func init() {
	sql_query_decorators.GetDecoratorFactoryInstance().RegisterCreator(&orderByDecoratorCreator{})
}

type orderByDecoratorCreator struct{}

func (o *orderByDecoratorCreator) Create(ctx context.Context, inner sql_query_decorators.SqlQueryRetriever, f sql_query_decorators.Filters) (sql_query_decorators.SqlQueryRetriever, error) {
	log.C(ctx).Info("creating overdue decorator in overdue decorator creator")

	last, containsLast := f.GetFilters()[constants.LAST]
	if containsLast && len(last) != 0 {
		inner = sql_query_decorators.NewOrderByDecorator(inner, constants.DESC_ORDER)
	} else {
		inner = sql_query_decorators.NewOrderByDecorator(inner, constants.ASC_ORDER)
	}

	return inner, nil
}

func (o *orderByDecoratorCreator) Priority() int {
	return constants.ORDER_BY_PRIORITY
}
