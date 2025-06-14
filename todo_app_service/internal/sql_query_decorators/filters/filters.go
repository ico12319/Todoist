package filters

type BaseFilters struct {
	Limit  string `validate:"omitempty,numeric,min=1"`
	Cursor string `validate:"omitempty,min=1"`
}

type TodoFilters struct {
	BaseFilters
	Status   string
	Priority string
}

type ListFilters struct {
	BaseFilters
	ListId string
}

type UserFilters struct {
	BaseFilters
	UserId string
}
