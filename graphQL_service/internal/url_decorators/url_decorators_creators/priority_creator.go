package url_decorators_creators

import (
	"Todo-List/internProject/graphQL_service/internal/gql_constants"
	"Todo-List/internProject/graphQL_service/internal/url_decorators"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
)

func init() {
	url_decorators.GetUrlDecoratorFactoryInstance().Register(&priorityCreator{})
}

type priorityCreator struct{}

func (*priorityCreator) Create(ctx context.Context, inner url_decorators.QueryParamsRetrievers, uFilters url_decorators.UrlFilters) url_decorators.QueryParamsRetrievers {
	log.C(ctx).Info("creating priority decorator in priority creator")

	priority, containsPriority := uFilters.GetFilters()[gql_constants.PRIORITY]
	if containsPriority && priority != nil {
		log.C(ctx).Info("successfully creating priority decorator in priority creator")
		inner = url_decorators.NewCriteriaDecorator(inner, gql_constants.PRIORITY, *priority)
	}

	return inner
}
