package gql_constants

const (
	LISTS_PATH        = "/lists"
	USER_PATH         = "/users"
	TODO_PATH         = "/todos"
	OWNER_PATH        = "/owner"
	COLLABORATOR_PATH = "/collaborators"
	ASSIGNEE_PATH     = "/assignee"
	REFRESH_PATH      = "/refresh"
	TOKENS_PATH       = "/tokens"
	RANDOM_PATH       = "/random"
	ACTIVITIES_PATH   = "/activities"
)

const (
	OPEN_LOWERCASE        = "open"
	DONE_LOWERCASE        = "done"
	IN_PROGRESS_LOWERCASE = "in progress"

	VERY_LOW_PRIORITY_LOWERCASE  = "very low"
	LOW_PRIORITY_LOWERCASE       = "low"
	MEDIUM_PRIORITY_LOWERCASE    = "medium"
	HIGH_PRIORITY_LOWERCASE      = "high"
	VERY_HIGH_PRIORITY_LOWERCASE = "very high"

	ADMIN_LOWERCASE  = "admin"
	WRITER_LOWERCASE = "writer"
	READER_LOWERCASE = "reader"
)

const (
	CURSOR   = "cursor"
	LIMIT    = "limit"
	STATUS   = "status"
	PRIORITY = "priority"
	TYPE     = "overdue"
	ROLE     = "role"
)

const (
	API_ENDPOINT    = "/api"
	HEALTH_ENDPOINT = "/healthz"
	READY_ENDPOINT  = "/readyz"
)
