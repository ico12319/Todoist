package sql_decorators_creators

import (
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"context"
)

func init() {
	sql_query_decorators.GetDecoratorFactoryInstance().RegisterCreator(&statusDecoratorCreator{})
}

type statusDecoratorCreator struct{}

func (s *statusDecoratorCreator) Create(ctx context.Context, inner sql_query_decorators.SqlQueryRetriever, f sql_query_decorators.Filters) (sql_query_decorators.SqlQueryRetriever, error) {
	log.C(ctx).Info("creating status decorator in status decorator creator")

	status, containsStatus := f.GetFilters()[constants.STATUS]
	if containsStatus && len(status) != 0 {
		inner = sql_query_decorators.NewCriteriaDecorator(inner, constants.STATUS, status)
	}

	return inner, nil
}

func (*statusDecoratorCreator) Priority() int {
	return constants.DEFAULT_PRIORITY
}
