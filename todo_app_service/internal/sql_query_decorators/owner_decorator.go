package sql_query_decorators

import (
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
	"fmt"
)

type ownerDecorator struct {
	inner   SqlQueryRetriever
	ownerId string
}

func NewOwnerDecorator(inner SqlQueryRetriever, ownerId string) *ownerDecorator {
	return &ownerDecorator{
		inner:   inner,
		ownerId: ownerId,
	}
}

func (o *ownerDecorator) DetermineCorrectSqlQuery(ctx context.Context) string {
	log.C(ctx).Info("determining correct sql query in owner decorator")

	currentQuery := o.inner.DetermineCorrectSqlQuery(ctx)
	addition := determineUserListsAddition(currentQuery)

	formattedSuffix := fmt.Sprintf(" %s owner = '%s'", addition, o.ownerId)

	currentQuery += formattedSuffix

	return currentQuery
}
