package url_decorators

import (
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
	"sync"
)

type UrlFilters interface {
	GetFilters() map[string]*string
}

type urlDecoratorsCreator interface {
	Create(context.Context, QueryParamsRetrievers, UrlFilters) QueryParamsRetrievers
}

type urlDecoratorFactory struct {
	creators []urlDecoratorsCreator
}

var (
	mu       sync.Mutex
	once     sync.Once
	instance *urlDecoratorFactory
)

func GetUrlDecoratorFactoryInstance() *urlDecoratorFactory {
	once.Do(func() {
		instance = &urlDecoratorFactory{make([]urlDecoratorsCreator, 0)}
	})

	return instance
}

func (u *urlDecoratorFactory) Register(creator urlDecoratorsCreator) {
	mu.Lock()
	defer mu.Unlock()

	u.creators = append(u.creators, creator)
}

func (u *urlDecoratorFactory) CreateUrlDecorator(ctx context.Context, initialUrl string, filters UrlFilters) QueryParamsRetrievers {
	if filters == nil {
		return nil
	}

	mu.Lock()
	defer mu.Unlock()

	log.C(ctx).Info("creating url decorator in url decorator factory")

	baseDecorator := newBaseUrl(initialUrl)

	for _, creator := range u.creators {
		baseDecorator = creator.Create(ctx, baseDecorator, filters)
	}

	return baseDecorator
}
