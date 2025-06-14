package sql_query_decorators

import "strings"

func determineAddition(baseQuery string) string {
	var addition string
	if strings.Contains(baseQuery, "WHERE") {
		addition = "AND"
	} else {
		addition = "WHERE"
	}

	return addition
}
