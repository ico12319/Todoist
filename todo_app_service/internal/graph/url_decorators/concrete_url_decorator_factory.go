package url_decorators

import (
	"context"
	"fmt"
	gql "github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/model"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/url_decorators/url_filters"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/configuration"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/constants"
)

type statusConverter interface {
	ToStringStatus(status *gql.TodoStatus) string
}

type priorityConverter interface {
	ToStringPriority(priority *gql.Priority) string
}

type commonFactory interface {
	CreateCommonUrlDecorator(ctx context.Context, initialUrl string, baseFilters *url_filters.BaseFilters) (QueryParamsRetrievers, error)
}

// factory to dynamically build correct retriever, without nesting if else constructions
type concreteUrlDecoratorFactory struct {
	sConverter statusConverter
	pConverter priorityConverter
	cFactory   commonFactory
}

func NewQueryParamsRetrieverFactory(sConverter statusConverter, pConverter priorityConverter, cFactory commonFactory) *concreteUrlDecoratorFactory {
	return &concreteUrlDecoratorFactory{sConverter: sConverter, pConverter: pConverter, cFactory: cFactory}
}

func (c *concreteUrlDecoratorFactory) CreateConcreteUrlDecorator(ctx context.Context, initialUrl string, todoFilters *url_filters.TodoFilters) (QueryParamsRetrievers, error) {
	log.C(ctx).Debug("creating todo correct todo retriever in factory")

	if todoFilters == nil {
		return nil, fmt.Errorf("invalid todo filters passed")
	}

	currentUrlRetriever, err := c.cFactory.CreateCommonUrlDecorator(ctx, initialUrl, &todoFilters.BaseFilters)
	if err != nil {
		log.C(ctx).Error("failed to create todo url decorator, error when calling common factory")
		return nil, err
	}

	if todoFilters.TodoFilters != nil && todoFilters.TodoFilters.Status != nil {
		convertedStatus := c.sConverter.ToStringStatus(todoFilters.TodoFilters.Status)
		currentUrlRetriever = newCriteriaRetriever(currentUrlRetriever, constants.STATUS, convertedStatus)
	}

	if todoFilters.TodoFilters != nil && todoFilters.TodoFilters.Priority != nil {
		convertedPriority := c.pConverter.ToStringPriority(todoFilters.TodoFilters.Priority)
		currentUrlRetriever = newCriteriaRetriever(currentUrlRetriever, constants.PRIORITY, convertedPriority)
	}

	return currentUrlRetriever, nil
}
