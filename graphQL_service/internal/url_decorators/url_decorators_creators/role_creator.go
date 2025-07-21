package url_decorators_creators

import (
	"Todo-List/internProject/graphQL_service/internal/url_decorators"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"context"
)

func init() {
	url_decorators.GetUrlDecoratorFactoryInstance().Register(&roleCreator{})
}

type roleCreator struct{}

func (o *roleCreator) Create(ctx context.Context, inner url_decorators.QueryParamsRetrievers, uFilters url_decorators.UrlFilters) url_decorators.QueryParamsRetrievers {
	log.C(ctx).Info("creating owner decorator in owner creator")

	role, containsRole := uFilters.GetFilters()[constants.ROLE]
	if containsRole && role != nil {
		inner = url_decorators.NewCriteriaDecorator(inner, constants.ROLE, *role)
	}

	return inner
}
