package sql_query_decorators

import (
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
	"sort"
	"sync"
)

type Filters interface {
	GetFilters() map[string]string
}

type SqlDecoratorCreator interface {
	Create(context.Context, SqlQueryRetriever, Filters) (SqlQueryRetriever, error)
	Priority() int
}
type decoratorFactory struct {
	creators []SqlDecoratorCreator
}

var (
	once     sync.Once
	instance *decoratorFactory
)

func GetDecoratorFactoryInstance() *decoratorFactory {
	once.Do(func() {
		instance = &decoratorFactory{make([]SqlDecoratorCreator, 0)}
	})

	return instance
}

func (c *decoratorFactory) RegisterCreator(creator SqlDecoratorCreator) {
	c.creators = append(c.creators, creator)
}

func (c *decoratorFactory) CreateSqlDecorator(ctx context.Context, f Filters, initialQuery string) (SqlQueryRetriever, error) {
	log.C(ctx).Info("creating common decorator in common decorator factory")

	retriever := newBaseQuery(initialQuery)

	c.creators = SortCreators(c.creators)

	var err error
	for _, creator := range c.creators {
		retriever, err = creator.Create(ctx, retriever, f)
		if err != nil {
			log.C(ctx).Errorf("failed to create decorator, error %s", err.Error())
			return nil, err
		}
	}

	return retriever, nil
}

func SortCreators(creators []SqlDecoratorCreator) []SqlDecoratorCreator {
	sort.Slice(creators, func(index1, index2 int) bool {
		return creators[index1].Priority() < creators[index2].Priority()
	})

	return creators
}
