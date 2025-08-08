package url_decorators_creators

import (
	"Todo-List/internProject/graphQL_service/internal/gql_constants"
	url_decorators2 "Todo-List/internProject/graphQL_service/internal/url_decorators"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
)

func init() {
	url_decorators2.GetUrlDecoratorFactoryInstance().Register(&afterCursorCreator{})
}

type afterCursorCreator struct{}

func (c *afterCursorCreator) Create(ctx context.Context, inner url_decorators2.QueryParamsRetrievers, uFilters url_decorators2.UrlFilters) url_decorators2.QueryParamsRetrievers {
	log.C(ctx).Info("creating after cursor url decorator in cursor creator")

	cursor, containsCursor := uFilters.GetFilters()[gql_constants.AFTER]
	if containsCursor && cursor != nil {
		log.C(ctx).Info("successfully creating after cursor url decorator in cursor creator")
		inner = url_decorators2.NewCriteriaDecorator(inner, gql_constants.AFTER, *cursor)
	}

	return inner
}
