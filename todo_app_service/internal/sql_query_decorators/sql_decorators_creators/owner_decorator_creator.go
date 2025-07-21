package sql_decorators_creators

import (
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"context"
)

func init() {
	sql_query_decorators.GetDecoratorFactoryInstance().RegisterCreator(&ownerDecoratorCreator{})
}

type ownerDecoratorCreator struct{}

func (o *ownerDecoratorCreator) Create(ctx context.Context, inner sql_query_decorators.SqlQueryRetriever, f sql_query_decorators.Filters) (sql_query_decorators.SqlQueryRetriever, error) {
	log.C(ctx).Info("creating an owner decorator in owner decorator creator")

	ownerId, containsOwnerId := f.GetFilters()[constants.OWNER_ROLE]
	if containsOwnerId && len(ownerId) != 0 {
		inner = sql_query_decorators.NewOwnerDecorator(inner, ownerId)
	}

	return inner, nil
}

func (o *ownerDecoratorCreator) Priority() int {
	return constants.DEFAULT_PRIORITY
}
