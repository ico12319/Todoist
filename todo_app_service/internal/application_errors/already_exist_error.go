package application_errors

import "fmt"

type AlreadyExistError struct {
	target string
	name   string
}

func NewAlreadyExistError(target string, name string) *AlreadyExistError {
	return &AlreadyExistError{target: target, name: name}
}

func (a AlreadyExistError) Error() string {
	return fmt.Sprintf("a %s with identifier %q already exists", a.target, a.name)
}

func (a AlreadyExistError) Is(target error) bool {
	t, ok := target.(*AlreadyExistError)
	if !ok {
		return false
	}

	return a.target == t.target && a.name == t.name
}
