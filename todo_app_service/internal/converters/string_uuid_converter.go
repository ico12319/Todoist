package converters

import (
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/utils"
	"github.com/gofrs/uuid"
)

type stringUUIDConverter struct{}

func NewStringUUIDConverter() *stringUUIDConverter {
	return &stringUUIDConverter{}
}

func (*stringUUIDConverter) ConvertFromStringToUUID(str string) uuid.UUID {
	return utils.ConvertFromStringToUUID(str)
}

func (*stringUUIDConverter) ConvertFromUUIDToString(id uuid.UUID) string {
	return id.String()
}
