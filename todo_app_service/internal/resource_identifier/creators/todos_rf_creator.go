package creators

import (
	"Todo-List/internProject/todo_app_service/internal/resource_identifier"
	"Todo-List/internProject/todo_app_service/pkg/constants"
)

func init() {
	resource_identifier.GetAdapterInstance().Register(&todosRfCreator{})
}

type todosRfCreator struct{}

func (*todosRfCreator) Create(rf resource_identifier.ResourceIdentifier) resource_identifier.ResourceIdentifier {
	adaptResourceIdentifierIfNeeded(rf, constants.TodosIdentifier, constants.TodosSQLTableName)

	return rf
}
