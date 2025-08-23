package creators

import (
	"Todo-List/internProject/todo_app_service/internal/resource_identifier"
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"context"
)

func init() {
	resource_identifier.GetAdapterInstance().Register(&listsRfCreator{})
}

type listsRfCreator struct{}

func (*listsRfCreator) Create(rf resource_identifier.ResourceIdentifier) resource_identifier.ResourceIdentifier {
	adaptResourceIdentifierIfNeeded(rf, constants.ListIdentifier, constants.ListsSQLTableName)

	log.C(context.Background()).Infof("izobshto tuka li e %s", rf.GetResourceIdentifier())
	return rf
}
