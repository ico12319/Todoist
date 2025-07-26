package gql_errors

import "fmt"

type RestHealthError struct {
	BadStatusCode int
}

func NewRestHealthError(BadStatusCode int) *RestHealthError {
	return &RestHealthError{BadStatusCode: BadStatusCode}
}

func (r *RestHealthError) Error() string {
	return fmt.Sprintf("bad http status code received when calling REST api, received code: %d", r.BadStatusCode)
}
