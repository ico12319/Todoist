package url_decorators_creators

import (
	"Todo-List/internProject/graphQL_service/internal/gql_constants"
	url_decorators2 "Todo-List/internProject/graphQL_service/internal/url_decorators"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
)

func init() {
	url_decorators2.GetUrlDecoratorFactoryInstance().Register(&lastLimitCreator{})
}

type lastLimitCreator struct{}

func (*lastLimitCreator) Create(ctx context.Context, inner url_decorators2.QueryParamsRetrievers, uFilters url_decorators2.UrlFilters) url_decorators2.QueryParamsRetrievers {
	log.C(ctx).Info("creating limit url decorator in limit creator")

	limit, containsLimit := uFilters.GetFilters()[gql_constants.LAST]
	if containsLimit && limit != nil {
		log.C(ctx).Info("successfully creating limit url decorator in limit creator")
		inner = url_decorators2.NewCriteriaDecorator(inner, gql_constants.LAST, *limit)
	}

	return inner
}
