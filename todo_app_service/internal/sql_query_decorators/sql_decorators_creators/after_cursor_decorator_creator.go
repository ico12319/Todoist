package sql_decorators_creators

import (
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"context"
)

func init() {
	sql_query_decorators.GetDecoratorFactoryInstance().RegisterCreator(&afterCursorDecoratorCreator{})
}

type afterCursorDecoratorCreator struct{}

func (*afterCursorDecoratorCreator) Create(ctx context.Context, inner sql_query_decorators.SqlQueryRetriever, filters sql_query_decorators.Filters) (sql_query_decorators.SqlQueryRetriever, error) {
	log.C(ctx).Info("creating cursor decorator in cursor decorator creator")

	cursor, contains := filters.GetFilters()[constants.AFTER]
	if contains && len(cursor) != 0 {
		inner = sql_query_decorators.NewCursorDecorator(inner, cursor, ">")
	}

	return inner, nil
}

func (*afterCursorDecoratorCreator) Priority() int {
	return constants.CURSOR_PRIORITY
}
