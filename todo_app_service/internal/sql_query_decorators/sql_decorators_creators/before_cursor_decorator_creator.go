package sql_decorators_creators

import (
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"context"
)

func init() {
	sql_query_decorators.GetDecoratorFactoryInstance().RegisterCreator(&beforeCursorDecoratorCreator{})
}

type beforeCursorDecoratorCreator struct{}

func (*beforeCursorDecoratorCreator) Create(ctx context.Context, inner sql_query_decorators.SqlQueryRetriever, filters sql_query_decorators.Filters) (sql_query_decorators.SqlQueryRetriever, error) {
	log.C(ctx).Info("creating before cursor decorator in before cursor decorator creator")

	cursor, contains := filters.GetFilters()[constants.BEFORE]
	if contains && len(cursor) != 0 {
		inner = sql_query_decorators.NewCursorDecorator(inner, cursor, "<")
	}

	return inner, nil
}

func (*beforeCursorDecoratorCreator) Priority() int {
	return constants.CURSOR_PRIORITY
}
