package users

import (
	"Todo-List/internProject/todo_app_service/pkg/models"
	"fmt"
	"strings"
)

func parseUserQuery(sqlFields []string) string {
	return fmt.Sprintf("UPDATE users SET %s WHERE id = :id", strings.Join(sqlFields, ", "))
}

func determineSqlFieldsAndParamsUser(user *models.User, sqlExecParams map[string]interface{}, sqlFields *[]string) {
	if len(user.Email) != 0 {
		sqlExecParams["email"] = user.Email
		*sqlFields = append(*sqlFields, "email = :email")
	}

	if len(user.Role) != 0 {
		sqlExecParams["role"] = user.Role
		*sqlFields = append(*sqlFields, "role = :role")
	}
}
