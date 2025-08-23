package constants

type TodoStatus string
type Priority string
type UserRole string

const (
	ListIdentifier       = "lists"
	TodosIdentifier      = "todos"
	UsersIdentifier      = "users"
	ListsUsersIdentifier = "lists users"
	UsersTodosIdentifier = "user todos"
	UsersListsIdentifier = "user lists"
)

// adapted
const (
	TodosSQLTableName           = "todos"
	ListsSQLTableName           = "lists"
	UsersSQLTableName           = "users"
	ListsCollaboratorsTableName = "lists_collaborators"
	UserListsTableName          = "lists_and_users"
	UserTodosTableName          = "user_todos"
)

const (
	Admin  UserRole = "admin"
	Writer UserRole = "writer"
	Reader UserRole = "reader"
)

const CONTENT_TYPE = "application/json"

const INVALID_REQUEST_BODY = "invalid request body"

const CONTEXT_NOT_CONTAINING_VALID_LIST_ID = "internal error: request context does not contain a valid list ID"
const CONTEXT_NOT_CONTAINING_VALID_TODO_ID = "internal error: request context does not contain a valid todo ID"
const CONTEXT_NOT_CONTAINING_VALID_USER = "internal error: request context does not contain a valid user"
const CONTEXT_NOT_CONTAINING_VALID_USER_ID = "internal error: request context does not contain a valid user ID"

const MISSING_USER_ID = "internal error: missing user_id"

const STATUS = "status"
const PRIORITY = "priority"
const AFTER = "after"
const FIRST = "first"
const EMAIL = "email"
const NAME = "name"
const ID = "id"
const LAST = "last"
const BEFORE = "before"
const ORDER_BY = "order by"
const LIMIT = "limit"

const LIST_TARGET = "list"
const USER_TARGET = "user"
const TODO_TARGET = "todo"
const REFRESH_TARGET = "refresh token"

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

const ORDER_BY_PRIORITY = 18

const ASC_ORDER = "ASC"
const DESC_ORDER = "DESC"

const OWNER_ROLE = "owner"
const PARTICIPANT_ROLE = "participant"

const ROLE = "role"

const OVERDUE = "overdue"

const TRUE_VALUE = "true"
const FALSE_VALUE = "false"

const EXPIRED = "EXPIRED"

const DATABASE_DOWN_STATUS = "database down"
const READY_STATUS = "ready"
const OK_STATUS = "ok"

const HTTP_COOKIES_MAX_AGE = 15552000
