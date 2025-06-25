package todos

import (
	"Todo-List/internProject/todo_app_service/pkg/models"
	"fmt"
	"strings"
)

var baseTodoGetQuery = "SELECT id,name,description,list_id,status,created_at,last_updated,assigned_to,due_date,priority FROM (SELECT * FROM todos ORDER BY id)"
var baseTodoGetByListIdQuery = "SELECT id,name,description,list_id,status,created_at,last_updated,assigned_to,due_date,priority FROM (SELECT * FROM todos ORDER BY id) WHERE list_Id = $1"
var (
	todoColumns = []string{
		"id", "name", "description", "list_id",
		"status", "created_at", "last_updated",
		"assigned_to", "due_date", "priority",
	}
	baseSelectTodos = fmt.Sprintf(
		"SELECT %s FROM todos ",
		strings.Join(todoColumns, ", "),
	)
)

func parseTodoQuery(sqlFields []string) string {
	return fmt.Sprintf("UPDATE todos SET %s WHERE id = :id", strings.Join(sqlFields, ", "))
}

func determineSqlFieldsAndParamsTodo(todo *models.Todo, sqlExecParams map[string]interface{}, sqlFields *[]string) {
	if len(todo.Name) != 0 {
		sqlExecParams["name"] = todo.Name
		*sqlFields = append(*sqlFields, "name = :name")
	}

	if len(todo.Description) != 0 {
		sqlExecParams["description"] = todo.Description
		*sqlFields = append(*sqlFields, "description = :description")
	}

	if len(todo.Status) != 0 {
		sqlExecParams["status"] = todo.Status
		*sqlFields = append(*sqlFields, "status = :status")
	}

	if len(todo.Priority) != 0 {
		sqlExecParams["priority"] = todo.Priority
		*sqlFields = append(*sqlFields, "priority = :priority")
	}

	if todo.AssignedTo != nil {
		sqlExecParams["assigned_to"] = *todo.AssignedTo
		*sqlFields = append(*sqlFields, "assigned_to = :assigned_to")
	}

	if todo.DueDate != nil {
		sqlExecParams["due_date"] = *todo.DueDate
		*sqlFields = append(*sqlFields, "due_date = :due_date")
	}
}

func determineAddition(baseQuery string) string {
	var addition string
	if strings.Contains(baseQuery, "WHERE") {
		addition = "AND"
	} else {
		addition = "WHERE"
	}

	return addition
}

func getCorrectTodosByListIdAdditionToQuery(baseQuery string) string {
	addition := determineAddition(baseQuery)
	return fmt.Sprintf("%s list_id = $1", addition)
}
