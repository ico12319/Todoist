package url_decorators_creators

import (
	"Todo-List/internProject/graphQL_service/internal/gql_constants"
	"Todo-List/internProject/graphQL_service/internal/url_decorators"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
)

func init() {
	url_decorators.GetUrlDecoratorFactoryInstance().Register(&statusCreator{})
}

type statusCreator struct{}

func (*statusCreator) Create(ctx context.Context, inner url_decorators.QueryParamsRetrievers, uFilters url_decorators.UrlFilters) url_decorators.QueryParamsRetrievers {
	log.C(ctx).Info("creating status url decorator in status creator")

	status, containsStatus := uFilters.GetFilters()[gql_constants.STATUS]
	if containsStatus && status != nil {
		log.C(ctx).Info("successfully creating status url decorator in status creator")
		inner = url_decorators.NewCriteriaDecorator(inner, gql_constants.STATUS, *status)
	}

	return inner
}
