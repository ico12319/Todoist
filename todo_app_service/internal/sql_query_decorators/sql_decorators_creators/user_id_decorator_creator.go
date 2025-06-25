package sql_decorators_creators

import (
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"context"
)

func init() {
	sql_query_decorators.GetDecoratorFactoryInstance().RegisterCreator(&userIdDecoratorCreator{})
}

type userIdDecoratorCreator struct{}

func (*userIdDecoratorCreator) Create(ctx context.Context, inner sql_query_decorators.SqlQueryRetriever, f sql_query_decorators.Filters) (sql_query_decorators.SqlQueryRetriever, error) {
	log.C(ctx).Info("creating user_id in user_id creator")

	userId, containsUserId := f.GetFilters()[constants.USER_ID]
	if containsUserId && len(userId) != 0 {
		inner = sql_query_decorators.NewCriteriaDecorator(inner, constants.USER_ID, userId)
	}

	return inner, nil
}

func (*userIdDecoratorCreator) Priority() int {
	return constants.DEFAULT_PRIORITY
}
