package application_errors

import "errors"

var InvalidListOrUserIdError = errors.New("invalid list_id or user_id provided")
