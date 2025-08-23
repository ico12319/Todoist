package filters

import (
	"Todo-List/internProject/todo_app_service/pkg/constants"
	"fmt"
	"strings"
)

type PaginationFilters struct {
	First  string `validate:"omitempty,numeric,min=1"`
	Last   string `validate:"omitempty,numeric,min=1"`
	After  string `validate:"omitempty,min=1"`
	Before string `validate:"omitempty,min=1"`
}

func (p *PaginationFilters) GetFilters() map[string]string {
	return map[string]string{
		constants.FIRST:  p.First,
		constants.LAST:   p.Last,
		constants.AFTER:  p.After,
		constants.BEFORE: p.Before,
	}
}

type TodoFilters struct {
	PaginationFilters
	Status   string
	Priority string
	Name     string
	Overdue  string
	ListID   string
	UserID   string
}

func (t *TodoFilters) GetFilters() map[string]string {
	return map[string]string{
		constants.FIRST:  t.First,
		constants.LAST:   t.Last,
		constants.AFTER:  t.After,
		constants.BEFORE: t.Before,
	}
}

// BuildSQLFiltering returns the where filtering clause plus the needed params so you can inject them
func (t *TodoFilters) BuildSQLFiltering() (string, []interface{}) {
	fields := make([]string, 0)
	params := make([]interface{}, 0)
	paramCounter := 0

	if len(t.Status) != 0 {
		paramCounter++
		elem := fmt.Sprintf(`status = $%d`, paramCounter)
		fields = append(fields, elem)

		params = append(params, t.Status)
	}
	if len(t.Priority) != 0 {
		paramCounter++
		elem := fmt.Sprintf(`priority = $%d`, paramCounter)
		fields = append(fields, elem)

		params = append(params, t.Priority)
	}
	if len(t.Name) != 0 {
		paramCounter++
		elem := fmt.Sprintf(`name = $%d`, paramCounter)
		fields = append(fields, elem)

		params = append(params, t.Name)
	}
	if t.Overdue == constants.TRUE_VALUE {
		elem := `(due_date IS NOT NULL AND due_date < current_date)`
		fields = append(fields, elem)
	}
	if t.Overdue == constants.FALSE_VALUE {
		elem := `(due_date IS NULL OR due_date > current_date)`
		fields = append(fields, elem)
	}
	if len(t.ListID) != 0 {
		paramCounter++
		elem := fmt.Sprintf(`list_id = $%d`, paramCounter)
		fields = append(fields, elem)

		params = append(params, t.ListID)
	}
	if len(t.UserID) != 0 {
		paramCounter++
		elem := fmt.Sprintf(`user_id = $%d`, paramCounter)
		fields = append(fields, elem)

		params = append(params, t.UserID)
	}

	var args string
	if len(fields) == 0 {
		args = `TRUE`
	} else {
		args = strings.Join(fields, " AND ")
	}

	filteringClause := fmt.Sprintf(` WHERE %s`, args)
	return filteringClause, params
}

type ListFilters struct {
	PaginationFilters
	Name    string
	OwnerID string
}

func (l *ListFilters) GetFilters() map[string]string {
	return map[string]string{
		constants.FIRST:  l.First,
		constants.LAST:   l.Last,
		constants.AFTER:  l.After,
		constants.BEFORE: l.Before,
	}
}

func (l *ListFilters) BuildSQLFiltering() (string, []interface{}) {
	fields := make([]string, 0)
	params := make([]interface{}, 0)
	paramCounter := 0

	if len(l.Name) != 0 {
		paramCounter++
		elem := fmt.Sprintf(`name = $%d`, paramCounter)
		fields = append(fields, elem)

		params = append(params, l.Name)
	}

	if len(l.OwnerID) != 0 {
		paramCounter++
		elem := fmt.Sprintf(`owner = $%d`, paramCounter)
		fields = append(fields, elem)

		params = append(params, l.OwnerID)
	}

	var args string
	if len(fields) == 0 {
		args = `TRUE`
	} else {
		args = strings.Join(fields, " AND ")
	}

	filteringClause := fmt.Sprintf(` WHERE %s`, args)
	return filteringClause, params
}

type UserFilters struct {
	PaginationFilters
	ListID string
}

func (u *UserFilters) GetFilters() map[string]string {
	return map[string]string{
		constants.FIRST:  u.First,
		constants.LAST:   u.Last,
		constants.AFTER:  u.After,
		constants.BEFORE: u.Before,
	}
}

func (u *UserFilters) BuildSQLFiltering() (string, []interface{}) {
	fields := make([]string, 0)
	params := make([]interface{}, 0)
	paramCounter := 0

	if len(u.ListID) != 0 {
		paramCounter++
		elem := fmt.Sprintf(`list_id = $%d`, paramCounter)
		fields = append(fields, elem)

		params = append(params, u.ListID)
	}

	var args string
	if len(fields) == 0 {
		args = `TRUE`
	} else {
		args = strings.Join(fields, " AND ")
	}

	filteringClause := fmt.Sprintf(` WHERE %s`, args)
	return filteringClause, params
}

type SqlFilters interface {
	GetFilters() map[string]string
	BuildSQLFiltering() (string, []interface{})
}
