package creators

import (
	"Todo-List/internProject/todo_app_service/internal/resource_identifier"
)

func adaptResourceIdentifierIfNeeded(rf resource_identifier.ResourceIdentifier, desiredResource string, adaptedResource string) {
	currentRf := rf.GetResourceIdentifier()
	if currentRf == desiredResource {
		adaptedRf := adaptedResource

		rf.SetResourceIdentifier(adaptedRf)
	}
}
