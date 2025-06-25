package sql_decorators_creators

import (
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"context"
	"fmt"
	"strconv"
)

func init() {
	sql_query_decorators.GetDecoratorFactoryInstance().RegisterCreator(&limitDecoratorCreator{})
}

type limitDecoratorCreator struct{}

func (*limitDecoratorCreator) Create(ctx context.Context, inner sql_query_decorators.SqlQueryRetriever, f sql_query_decorators.Filters) (sql_query_decorators.SqlQueryRetriever, error) {
	log.C(ctx).Info("creating limit decorator in limit decorator creator")

	limit, containsLimit := f.GetFilters()[constants.LIMIT]
	if containsLimit && len(limit) != 0 {
		parsedLimit, err := strconv.Atoi(limit)
		if err != nil {
			log.C(ctx).Errorf("failed to create limit decorator, error %s when trying to parse limit", err.Error())
			return nil, fmt.Errorf("invalid limit provided %s", limit)
		}

		inner = sql_query_decorators.NewLimitDecorator(inner, parsedLimit)
	}

	return inner, nil
}

func (*limitDecoratorCreator) Priority() int {
	return constants.LIMIT_PRIORITY
}
