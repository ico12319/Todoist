package sql_decorators_creators

import (
	"Todo-List/internProject/todo_app_service/internal/sql_query_decorators"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"context"
)

func init() {
	sql_query_decorators.GetDecoratorFactoryInstance().RegisterCreator(&participantDecoratorCreator{})
}

type participantDecoratorCreator struct{}

func (p *participantDecoratorCreator) Create(ctx context.Context, inner sql_query_decorators.SqlQueryRetriever, f sql_query_decorators.Filters) (sql_query_decorators.SqlQueryRetriever, error) {
	log.C(ctx).Info("creating an owner decorator in owner decorator creator")

	participantId, containsParticipantId := f.GetFilters()[constants.PARTICIPANT_ROLE]
	if containsParticipantId && len(participantId) != 0 {
		inner = sql_query_decorators.NewParticipantDecorator(inner, participantId)
	}

	return inner, nil
}

func (p *participantDecoratorCreator) Priority() int {
	return constants.DEFAULT_PRIORITY
}
