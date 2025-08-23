package resource_identifier

import "sync"

type adaptedResourceIdentifiersCreators interface {
	Create(rf ResourceIdentifier) ResourceIdentifier
}

type adapter struct {
	creators []adaptedResourceIdentifiersCreators // the actual creator will be one but all of them will be registered!
}

var (
	once     sync.Once
	instance *adapter
)

func GetAdapterInstance() *adapter {
	once.Do(func() {
		instance = &adapter{make([]adaptedResourceIdentifiersCreators, 0)}
	})

	return instance
}

func (a *adapter) Register(c adaptedResourceIdentifiersCreators) {
	a.creators = append(a.creators, c)
}

func (a *adapter) AdaptResourceIdentifier(rf ResourceIdentifier) string {
	for _, creator := range a.creators {
		rf = creator.Create(rf)
	}

	return rf.GetResourceIdentifier()
}
