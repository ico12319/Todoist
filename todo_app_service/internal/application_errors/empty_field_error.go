package application_errors

import "fmt"

type emptyFieldError struct {
	field string
}

func NewEmptyFieldError(field string) *emptyFieldError {
	return &emptyFieldError{field: field}
}

func (e emptyFieldError) Error() string {
	return fmt.Sprintf("field %s can't be empty", e.field)
}
