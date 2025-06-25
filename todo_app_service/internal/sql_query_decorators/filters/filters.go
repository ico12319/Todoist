package filters

import "Todo-List/internProject/todo_app_service/pkg/constants"

type BaseFilters struct {
	Limit  string `validate:"omitempty,numeric,min=1"`
	Cursor string `validate:"omitempty,min=1"`
}

func (b *BaseFilters) GetFilters() map[string]string {
	return map[string]string{
		constants.LIMIT:  b.Limit,
		constants.CURSOR: b.Cursor,
	}
}

type TodoFilters struct {
	BaseFilters
	Status   string
	Priority string
}

func (t *TodoFilters) GetFilters() map[string]string {
	return map[string]string{
		constants.LIMIT:    t.Limit,
		constants.CURSOR:   t.Cursor,
		constants.STATUS:   t.Status,
		constants.PRIORITY: t.Priority,
	}
}

type ListFilters struct {
	BaseFilters
	ListId string
}

func (l *ListFilters) GetFilters() map[string]string {
	return map[string]string{
		constants.LIMIT:  l.Limit,
		constants.CURSOR: l.Cursor,
	}
}

type UserFilters struct {
	BaseFilters
	UserId string
}

func (u *UserFilters) GetFilters() map[string]string {
	return map[string]string{
		constants.LIMIT:   u.Limit,
		constants.CURSOR:  u.Cursor,
		constants.USER_ID: u.UserId,
	}
}
