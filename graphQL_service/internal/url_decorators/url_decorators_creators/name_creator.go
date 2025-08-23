package url_decorators_creators

import (
	"Todo-List/internProject/graphQL_service/internal/gql_constants"
	url_decorators2 "Todo-List/internProject/graphQL_service/internal/url_decorators"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
)

func init() {
	url_decorators2.GetUrlDecoratorFactoryInstance().Register(&nameCreator{})
}

type nameCreator struct{}

func (c *nameCreator) Create(ctx context.Context, inner url_decorators2.QueryParamsRetrievers, uFilters url_decorators2.UrlFilters) url_decorators2.QueryParamsRetrievers {
	log.C(ctx).Info("creating name url decorator in name creator")

	name, containsName := uFilters.GetFilters()[gql_constants.NAME]
	if containsName && name != nil {
		log.C(ctx).Info("successfully creating name url decorator in name creator")
		inner = url_decorators2.NewCriteriaDecorator(inner, gql_constants.NAME, *name)
	}

	return inner
}
