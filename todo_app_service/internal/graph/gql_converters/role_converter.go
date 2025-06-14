package gql_converters

import (
	"github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/gql_constants"
	gql "github.com/I763039/Todo-List/internProject/todo_app_service/internal/graph/model"
)

type roleConverter struct{}

func NewRoleConverter() *roleConverter {
	return &roleConverter{}
}

func (*roleConverter) ToStringRole(role *gql.UserRole) string {
	ptrValue := *role

	if ptrValue == gql.UserRoleAdmin {
		return gql_constants.ADMIN_LOWERCASE
	} else if ptrValue == gql.UserRoleReader {
		return gql_constants.READER_LOWERCASE
	} else if ptrValue == gql.UserRoleWriter {
		return gql_constants.WRITER_LOWERCASE
	}

	return ""
}
