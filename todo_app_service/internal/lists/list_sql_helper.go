package lists

import (
	models2 "Todo-List/internProject/todo_app_service/pkg/models"
	"fmt"
	"strings"
)

var baseListGetQuery = "SELECT id,name,created_at,last_updated,owner,description FROM (SELECT * FROM lists ORDER BY id)"
var baseCollaboratorsGetQuery = "WITH sorted_user_cte AS (SELECT users.id,users.email,users.role,list_id FROM users JOIN user_lists ON users.id = user_lists.user_id order by users.id) SELECT id, email, role FROM sorted_user_cte"

func parseSqlUpdateListQuery(sqlFields []string) string {
	sqlQuery := fmt.Sprintf("UPDATE lists SET %s WHERE id = :id", strings.Join(sqlFields, ", "))
	return sqlQuery
}

func determineSqlFieldsAndParamsList(list *models2.List, sqlExecParams map[string]interface{}, sqlFields *[]string) {
	if len(list.Name) != 0 {
		sqlExecParams["name"] = list.Name
		*sqlFields = append(*sqlFields, "name = :name")
	}

	if len(list.Description) != 0 {
		sqlExecParams["description"] = list.Description
		*sqlFields = append(*sqlFields, "description = :description")
	}

	if !list.CreatedAt.IsZero() {
		sqlExecParams["created_at"] = list.CreatedAt
		*sqlFields = append(*sqlFields, "created_at = :created_at")
	}

	if !list.LastUpdated.IsZero() {
		sqlExecParams["last_updated"] = list.LastUpdated
		*sqlFields = append(*sqlFields, "last_updated = :last_updated")
	}

	if len(list.Owner) != 0 {
		sqlExecParams["owner"] = list.Owner
		*sqlFields = append(*sqlFields, "owner = :owner")
	}
}
