package url_decorators_creators

import (
	"Todo-List/internProject/graphQL_service/internal/gql_constants"
	"Todo-List/internProject/graphQL_service/internal/url_decorators"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
)

func init() {
	url_decorators.GetUrlDecoratorFactoryInstance().Register(&cursorCreator{})
}

type cursorCreator struct{}

func (c *cursorCreator) Create(ctx context.Context, inner url_decorators.QueryParamsRetrievers, uFilters url_decorators.UrlFilters) url_decorators.QueryParamsRetrievers {
	log.C(ctx).Info("creating cursor url decorator in cursor creator")

	cursor, containsCursor := uFilters.GetFilters()[gql_constants.CURSOR]

	if containsCursor && cursor != nil {
		log.C(ctx).Info("successfully creating cursor url decorator in cursor creator")
		inner = url_decorators.NewCriteriaDecorator(inner, gql_constants.CURSOR, *cursor)
	}

	return inner
}
