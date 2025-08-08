package sql_query_decorators

import (
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
	"fmt"
)

type participantDecorator struct {
	inner         SqlQueryRetriever
	participantId string
}

func NewParticipantDecorator(inner SqlQueryRetriever, participantId string) *participantDecorator {
	return &participantDecorator{
		inner:         inner,
		participantId: participantId,
	}
}

func (p *participantDecorator) DetermineCorrectSqlQuery(ctx context.Context) string {
	log.C(ctx).Info("determining correct sql query in participant decorator")

	currentQuery := p.inner.DetermineCorrectSqlQuery(ctx)

	formattedSuffix := fmt.Sprintf(" AND user_id = '%s'", p.participantId)

	currentQuery += formattedSuffix

	return currentQuery
}
