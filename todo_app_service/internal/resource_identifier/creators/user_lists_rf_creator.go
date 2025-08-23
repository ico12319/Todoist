package creators

import (
	"Todo-List/internProject/todo_app_service/internal/resource_identifier"
	"Todo-List/internProject/todo_app_service/pkg/constants"
)

func init() {
	resource_identifier.GetAdapterInstance().Register(&userListsRfCreator{})
}

type userListsRfCreator struct{}

func (*userListsRfCreator) Create(rf resource_identifier.ResourceIdentifier) resource_identifier.ResourceIdentifier {
	adaptResourceIdentifierIfNeeded(rf, constants.UserListsTableName, constants.UserListsTableName)

	return rf
}
