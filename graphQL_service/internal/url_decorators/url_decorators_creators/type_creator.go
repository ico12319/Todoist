package url_decorators_creators

import (
	"Todo-List/internProject/graphQL_service/internal/gql_constants"
	"Todo-List/internProject/graphQL_service/internal/url_decorators"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
)

func init() {
	url_decorators.GetUrlDecoratorFactoryInstance().Register(&typeCreator{})
}

type typeCreator struct{}

func (*typeCreator) Create(ctx context.Context, inner url_decorators.QueryParamsRetrievers, uFilters url_decorators.UrlFilters) url_decorators.QueryParamsRetrievers {
	log.C(ctx).Info("creating status url decorator in status creator")

	tType, containsType := uFilters.GetFilters()[gql_constants.TYPE]
	if containsType && tType != nil {
		log.C(ctx).Info("successfully creating type url decorator in status creator")
		inner = url_decorators.NewCriteriaDecorator(inner, gql_constants.TYPE, *tType)
	}

	return inner
}
