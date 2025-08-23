package url_decorators_creators

import (
	"Todo-List/internProject/graphQL_service/internal/gql_constants"
	"Todo-List/internProject/graphQL_service/internal/url_decorators"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
)

func init() {
	url_decorators.GetUrlDecoratorFactoryInstance().Register(&limitCreator{})
}

type limitCreator struct{}

func (*limitCreator) Create(ctx context.Context, inner url_decorators.QueryParamsRetrievers, uFilters url_decorators.UrlFilters) url_decorators.QueryParamsRetrievers {
	log.C(ctx).Info("creating limit url decorator in limit creator")

	limit, containsLimit := uFilters.GetFilters()[gql_constants.LIMIT]
	if containsLimit && limit != nil {
		log.C(ctx).Info("successfully creating limit url decorator in limit creator")
		inner = url_decorators.NewCriteriaDecorator(inner, gql_constants.LIMIT, *limit)
	}

	return inner
}
