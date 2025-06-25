package url_decorators

import (
	"Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
	"net/url"
)

// concrete decorator
type criteriaDecorator struct {
	inner          QueryParamsRetrievers
	condition      string
	conditionValue string
}

func NewCriteriaDecorator(inner QueryParamsRetrievers, condition string, conditionValue string) QueryParamsRetrievers {
	return &criteriaDecorator{inner: inner, condition: condition, conditionValue: conditionValue}
}

func (t *criteriaDecorator) DetermineCorrectQueryParams(ctx context.Context, serverAddress string) (string, error) {
	log.C(ctx).Debugf("crafting correct query params in todos by criteria retriever")

	currentUrl, err := t.inner.DetermineCorrectQueryParams(ctx, serverAddress)
	if err != nil {
		log.C(ctx).Error("error when trying to determine query params in todos by criteria retriever")
		return "", err
	}

	u, err := url.Parse(currentUrl)
	if err != nil {
		log.C(ctx).Error("error when trying to parse query in todos by criteria retriever")
		return "", err
	}

	query := u.Query()
	query.Set(t.condition, t.conditionValue)
	u.RawQuery = query.Encode()

	return u.String(), nil
}
