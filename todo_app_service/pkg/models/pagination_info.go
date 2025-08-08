package models

type PaginationInfo struct {
	FirstID string `json:"first_id"`
	LastID string `json:"last_id"`
	TotalCount int `json:"total_count"`
}
