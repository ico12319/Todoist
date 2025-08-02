package pagination

type Page struct {
	StartCursor string `json:"start_cursor"`
	EndCursor   string `json:"end_cursor"`
	HasNextPage bool   `json:"has_next_page"`
	HasPrevPage bool   `json:"has_prev_page"`
}
