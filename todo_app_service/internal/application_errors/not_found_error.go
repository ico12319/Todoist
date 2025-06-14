package application_errors

import "fmt"

type NotFoundError struct {
	target string
	id     string
}

func NewNotFoundError(target string, id string) *NotFoundError {
	return &NotFoundError{target: target, id: id}
}

func (n NotFoundError) Error() string {
	return fmt.Sprintf("%s with id %s not found", n.target, n.id)
}

func (n NotFoundError) Is(target error) bool {
	t, ok := target.(NotFoundError)
	if !ok {
		return false
	}

	return n.target == t.target && n.id == t.id
}
