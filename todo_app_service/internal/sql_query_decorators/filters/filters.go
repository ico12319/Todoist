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
	Overdue  string
}

func (t *TodoFilters) GetFilters() map[string]string {
	return map[string]string{
		constants.LIMIT:    t.Limit,
		constants.CURSOR:   t.Cursor,
		constants.STATUS:   t.Status,
		constants.PRIORITY: t.Priority,
		constants.OVERDUE:  t.Overdue,
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

type RolePair struct {
	roleType string
	id       string
}

type UserFilters struct {
	BaseFilters
	OwnerId       string
	ParticipantId string
}

func (u *UserFilters) GetFilters() map[string]string {
	return map[string]string{
		constants.LIMIT:            u.Limit,
		constants.CURSOR:           u.Cursor,
		constants.OWNER_ROLE:       u.OwnerId,
		constants.PARTICIPANT_ROLE: u.ParticipantId,
	}
}
