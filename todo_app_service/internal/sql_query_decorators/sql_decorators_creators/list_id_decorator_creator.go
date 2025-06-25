package sql_decorators_creators

import (
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"context"
)

func init() {
	sql_query_decorators.GetDecoratorFactoryInstance().RegisterCreator(&listIdDecoratorCreator{})
}

type listIdDecoratorCreator struct{}

func (*listIdDecoratorCreator) Create(ctx context.Context, inner sql_query_decorators.SqlQueryRetriever, f sql_query_decorators.Filters) (sql_query_decorators.SqlQueryRetriever, error) {
	log.C(ctx).Info("creating list_id decorator in list decorator creator")

	listId, containsListId := f.GetFilters()[constants.LIST_ID]
	if containsListId && len(listId) != 0 {
		inner = sql_query_decorators.NewCriteriaDecorator(inner, constants.LIST_ID, listId)
	}

	return inner, nil
}

func (*listIdDecoratorCreator) Priority() int {
	return constants.DEFAULT_PRIORITY
}
