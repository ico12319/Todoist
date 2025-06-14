package application_errors

import "errors"

var InvalidListOrTodoIdError = errors.New("invalid list_id or todo_id provided")
