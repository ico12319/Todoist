package filters

import "Todo-List/internProject/todo_app_service/pkg/constants"

type BaseFilters struct {
	First  string `validate:"omitempty,numeric,min=1"`
	Last   string `validate:"omitempty,numeric,min=1"`
	After  string `validate:"omitempty,min=1"`
	Before string `validate:"omitempty,min=1"`
}

func (b *BaseFilters) GetFilters() map[string]string {
	return map[string]string{
		constants.FIRST:  b.First,
		constants.LAST:   b.Last,
		constants.AFTER:  b.After,
		constants.BEFORE: b.Before,
	}
}

type TodoFilters struct {
	BaseFilters
	Status   string
	Priority string
	ID       string
	Name     string
	Overdue  string
}

func (t *TodoFilters) GetFilters() map[string]string {
	return map[string]string{
		constants.FIRST:    t.First,
		constants.LAST:     t.Last,
		constants.AFTER:    t.After,
		constants.BEFORE:   t.Before,
		constants.STATUS:   t.Status,
		constants.PRIORITY: t.Priority,
		constants.OVERDUE:  t.Overdue,
		constants.ID:       t.ID,
		constants.NAME:     t.Name,
	}
}

type ListFilters struct {
	BaseFilters
	Name string
	ID   string
}

func (l *ListFilters) GetFilters() map[string]string {
	return map[string]string{
		constants.FIRST:  l.First,
		constants.LAST:   l.Last,
		constants.AFTER:  l.After,
		constants.BEFORE: l.Before,
		constants.NAME:   l.Name,
		constants.ID:     l.ID,
	}
}

type UserFilters struct {
	BaseFilters
	Email string
	ID    string
}

func (u *UserFilters) GetFilters() map[string]string {
	return map[string]string{
		constants.FIRST:  u.First,
		constants.LAST:   u.Last,
		constants.AFTER:  u.After,
		constants.BEFORE: u.Before,
		constants.ID:     u.ID,
		constants.EMAIL:  u.Email,
	}
}
