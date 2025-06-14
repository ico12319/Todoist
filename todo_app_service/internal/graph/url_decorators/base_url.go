package url_decorators

import (
	"context"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/configuration"
	"net/url"
)

// abstract component!
type QueryParamsRetrievers interface {
	DetermineCorrectQueryParams(ctx context.Context, serverAddress string) (string, error)
}

// concrete component!
type baseUrl struct {
	initialUrl string
}

func newBaseUrl(initialUrl string) QueryParamsRetrievers {
	return &baseUrl{initialUrl: initialUrl}
}

func (b *baseUrl) DetermineCorrectQueryParams(ctx context.Context, serverAddress string) (string, error) {
	log.C(ctx).Debugf("crafting correct query params in all todos retriever")

	u, err := url.Parse(serverAddress)
	if err != nil {
		log.C(ctx).Errorf("error when trying to parse url address in all todos retriever")
		return "", err
	}

	u.Path += b.initialUrl

	return u.String(), nil
}
