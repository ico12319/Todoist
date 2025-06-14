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

func (e emptyFieldError) Is(target error) bool {
	t, ok := target.(*emptyFieldError)
	if !ok {
		return false
	}

	return e.field == t.field
}
