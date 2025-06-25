package gql_converters

import (
	gql "Todo-List/internProject/graphQL_service/graph/model"
	"Todo-List/internProject/graphQL_service/internal/gql_constants"
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
