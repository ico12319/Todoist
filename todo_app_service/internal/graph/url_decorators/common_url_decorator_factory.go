package url_decorators

import (
	"context"
	"fmt"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/url_decorators/url_filters"
	log "github.com/I763039/Todo-List/internProject/todo_app_service/pkg/configuration"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/constants"
	"strconv"
)

type commonDecoratorFactory struct{}

func NewCommonDecoratorFactory() *commonDecoratorFactory {
	return &commonDecoratorFactory{}
}

func (c *commonDecoratorFactory) CreateCommonUrlDecorator(ctx context.Context, initialUrl string, baseFilters *url_filters.BaseFilters) (QueryParamsRetrievers, error) {
	log.C(ctx).Info("creating common url decorator in common decorator factory")

	if baseFilters == nil {
		return nil, fmt.Errorf("invalid base filters provided")
	}

	currentUrlRetriever := newBaseUrl(initialUrl)

	if baseFilters.Limit != nil {
		parsedLimit := strconv.Itoa(int(*baseFilters.Limit))
		currentUrlRetriever = newCriteriaRetriever(currentUrlRetriever, constants.LIMIT, parsedLimit)
	}

	if baseFilters.Cursor != nil {
		currentUrlRetriever = newCriteriaRetriever(currentUrlRetriever, constants.CURSOR, *baseFilters.Cursor)
	}

	return currentUrlRetriever, nil
}
