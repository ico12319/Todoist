package users

import (
	"fmt"
	"github.com/I763039/Todo-List/internProject/todo_app_service/pkg/models"
	"strings"
)

var baseUserGetQuery = "SELECT id,email,role FROM (SELECT * FROM users ORDER BY id)"
var baseUserGetLists = "WITH sorted_list_cte AS (SELECT lists.id,lists.name,lists.description,lists.created_at,lists.last_updated,lists.owner, user_id FROM lists JOIN user_lists ON lists.id = user_lists.list_id order by lists.id) SELECT id, name, description, created_at, last_updated, owner FROM sorted_list_cte"
var baseUserGetTodos = "WITH sorted_todo_cte AS (SELECT todos.id, todos.name, todos.description, todos.list_id, todos.status,todos.created_at, todos.last_updated, todos.assigned_to, todos.due_date, todos.priority, users.id AS user_id FROM todos JOIN users ON todos.assigned_to = users.id ORDER BY todos.id) SELECT id, name, description, list_id, status, created_at, last_updated, assigned_to, due_date, priority FROM sorted_todo_cte"

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
