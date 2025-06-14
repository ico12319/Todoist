package sql_query_decorators

import (
	"context"
	"fmt"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/sql_query_decorators/filters"
	log "github.com/I763039/Todo-List/internProject/todo_app_service/pkg/configuration"
	"strconv"
)

type commonDecoratorFactory struct{}

func NewCommonFactory() *commonDecoratorFactory {
	return &commonDecoratorFactory{}
}

func (c *commonDecoratorFactory) CreateCommonDecorator(ctx context.Context, inner SqlQueryRetriever, baseFilters *filters.BaseFilters) (SqlQueryRetriever, error) {
	log.C(ctx).Info("creating common decorator in common decorator factory")
	retriever := inner

	if len(baseFilters.Cursor) != 0 {
		retriever = newCursorDecorator(retriever, baseFilters.Cursor)
	}

	if len(baseFilters.Limit) != 0 {
		lim, err := strconv.Atoi(baseFilters.Limit)
		if err != nil {
			return nil, fmt.Errorf("invalid limit: %s", baseFilters.Limit)
		}
		retriever = newLimitDecorator(retriever, lim)
	}

	return retriever, nil
}
