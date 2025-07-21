package sql_query_decorators

import (
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
)

// the idea of this abstract component is to dynamically build sql query in order to extract todos!

//go:generate mockery --name=SqlQueryRetriever --output=./mocks --outpkg=mocks --filename=sql_retriever.go --with-expecter=true
type SqlQueryRetriever interface {
	DetermineCorrectSqlQuery(context.Context) string
}

// concrete component!
type baseQuery struct {
	initialQuery string
}

func NewBaseQuery(initialQuery string) SqlQueryRetriever {
	return &baseQuery{initialQuery: initialQuery}
}

func (b *baseQuery) DetermineCorrectSqlQuery(ctx context.Context) string {
	log.C(ctx).Info("getting todos in all todos retriever")

	return b.initialQuery
}
