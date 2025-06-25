package constants

type TodoStatus string

const (
	Done       TodoStatus = "done"
	InProgress TodoStatus = "in progress"
	Open       TodoStatus = "open"
)

type Priority string

const (
	VeryLow  Priority = "very low"
	Low      Priority = "low"
	Medium   Priority = "priority"
	High     Priority = "high"
	VeryHigh Priority = "very high"
)

type UserRole string

const (
	Admin  UserRole = "admin"
	Writer UserRole = "writer"
	Reader UserRole = "reader"
)

const CONTENT_TYPE = "application/json"

const INVALID_REQUEST_BODY = "invalid request body"
const DESCRIPTION_EMPTY = "description can't be empty"

const CONTEXT_NOT_CONTAINING_VALID_LIST_ID = "internal error: request context does not contain a valid list ID"
const CONTEXT_NOT_CONTAINING_VALID_TODO_ID = "internal error: request context does not contain a valid todo ID"
const CONTEXT_NOT_CONTAINING_VALID_USER = "internal error: request context does not contain a valid user"
const CONTEXT_NOT_CONTAINING_VALID_USER_ID = "internal error: request context does not contain a valid user ID"

const MISSING_USER_ID = "internal error: missing user_id"

const STATUS = "status"
const PRIORITY = "priority"
const LIMIT = "limit"
const CURSOR = "cursor"
const LIST_ID = "list_id"
const USER_ID = "user_id"

const LIST_TARGET = "list"
const USER_TARGET = "user"
const TODO_TARGET = "todo"

const JSON_FORMAT = "json"
const TEXT_FORMAT = "text"

const REQUEST_ID = "request_id"

const DEFAULT_LIMIT_VALUE = "100"

const ADMIN_ORG = "Admins"
const READER_ORG = "Readers"
const WRITER_ORG = "Writers"

const UUID_REGEX = "[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}"

const GOROUTINES_COUNT = 2

const DEFAULT_PRIORITY = 10

const CURSOR_PRIORITY = 15

const LIMIT_PRIORITY = 20
