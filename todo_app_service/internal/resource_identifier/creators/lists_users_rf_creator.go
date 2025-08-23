package creators

import (
	"Todo-List/internProject/todo_app_service/internal/resource_identifier"
	"Todo-List/internProject/todo_app_service/pkg/constants"
)

func init() {
	resource_identifier.GetAdapterInstance().Register(&listsUsersRfCreator{})
}

type listsUsersRfCreator struct{}

func (*listsUsersRfCreator) Create(rf resource_identifier.ResourceIdentifier) resource_identifier.ResourceIdentifier {
	adaptResourceIdentifierIfNeeded(rf, constants.ListsUsersIdentifier, constants.ListsCollaboratorsTableName)

	return rf
}
