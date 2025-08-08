package url_decorators_creators

import (
	"Todo-List/internProject/graphQL_service/internal/gql_constants"
	url_decorators2 "Todo-List/internProject/graphQL_service/internal/url_decorators"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
)

func init() {
	url_decorators2.GetUrlDecoratorFactoryInstance().Register(&beforeCursorCreator{})
}

type beforeCursorCreator struct{}

func (c *beforeCursorCreator) Create(ctx context.Context, inner url_decorators2.QueryParamsRetrievers, uFilters url_decorators2.UrlFilters) url_decorators2.QueryParamsRetrievers {
	log.C(ctx).Info("creating before cursor url decorator in cursor creator")

	cursor, containsCursor := uFilters.GetFilters()[gql_constants.BEFORE]
	if containsCursor && cursor != nil {
		log.C(ctx).Info("successfully creating before cursor url decorator in cursor creator")
		inner = url_decorators2.NewCriteriaDecorator(inner, gql_constants.BEFORE, *cursor)
	}

	return inner
}
