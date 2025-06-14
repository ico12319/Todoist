package sql_query_decorators

import (
	"context"
	"fmt"
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/sql_query_decorators/filters"
	log "github.com/I763039/Todo-List/internProject/todo_app_service/pkg/configuration"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/constants"
)

//go:generate mockery --name=commonFactory --output=./mocks --outpkg=mocks --filename=common_sql_decorator_factory.go --with-expecter=true
type commonFactory interface {
	CreateCommonDecorator(ctx context.Context, inner SqlQueryRetriever, baseFilters *filters.BaseFilters) (SqlQueryRetriever, error)
}

type concreteDecoratorFactory struct {
	cFactory commonFactory
}

func NewConcreteQueryDecoratorFactory(cFactory commonFactory) *concreteDecoratorFactory {
	return &concreteDecoratorFactory{cFactory: cFactory}
}

func (c *concreteDecoratorFactory) CreateTodoQueryDecorator(ctx context.Context, initialQuery string, todoFilters *filters.TodoFilters) (SqlQueryRetriever, error) {
	log.C(ctx).Info("creating todo sql query decorator in concrete decorator factory")
	if todoFilters == nil {
		return nil, fmt.Errorf("invalid todo filters provided")
	}

	currenRetriever := newBaseQuery(initialQuery)

	if len(todoFilters.Priority) != 0 {
		currenRetriever = newCriteriaDecorator(currenRetriever, constants.PRIORITY, todoFilters.Priority)
	}

	if len(todoFilters.Status) != 0 {
		currenRetriever = newCriteriaDecorator(currenRetriever, constants.STATUS, todoFilters.Status)
	}

	currenRetriever, err := c.cFactory.CreateCommonDecorator(ctx, currenRetriever, &todoFilters.BaseFilters)
	if err != nil {
		log.C(ctx).Errorf("failed to create todo query decorator, error %s when calling common factory", err.Error())
		return nil, err
	}

	return currenRetriever, nil
}

func (c *concreteDecoratorFactory) CreateUserQueryDecorator(ctx context.Context, initialQuery string, userFilters *filters.UserFilters) (SqlQueryRetriever, error) {
	log.C(ctx).Info("creating user query decorator in concrete decorator factory")
	if userFilters == nil {
		return nil, fmt.Errorf("invalid user filters provided")
	}

	currentRetriever := newBaseQuery(initialQuery)
	if len(userFilters.UserId) != 0 {
		currentRetriever = newCriteriaDecorator(currentRetriever, constants.USER_ID, userFilters.UserId)
	}

	currentRetriever, err := c.cFactory.CreateCommonDecorator(ctx, currentRetriever, &userFilters.BaseFilters)
	if err != nil {
		log.C(ctx).Errorf("failed to create todo query decorator, error when calling common factory")
		return nil, err
	}
	return currentRetriever, nil
}

func (c *concreteDecoratorFactory) CreateListQueryDecorator(ctx context.Context, initialQuery string, listFilters *filters.ListFilters) (SqlQueryRetriever, error) {
	log.C(ctx).Info("creating list query decorator in concrete decorator factory")

	if listFilters == nil {
		return nil, fmt.Errorf("invalid list filters provided")
	}

	currentRetriever := newBaseQuery(initialQuery)
	if len(listFilters.ListId) != 0 {
		currentRetriever = newCriteriaDecorator(currentRetriever, constants.LIST_ID, listFilters.ListId)
	}

	currentRetriever, err := c.cFactory.CreateCommonDecorator(ctx, currentRetriever, &listFilters.BaseFilters)
	if err != nil {
		log.C(ctx).Errorf("failed to create list query decorator, error when calling common factory")
		return nil, err
	}

	return currentRetriever, nil
}
