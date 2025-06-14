package generators

import "time"

type timeGenerator struct{}

func NewTimeGenerator() *timeGenerator {
	return &timeGenerator{}
}

func (t *timeGenerator) Now() time.Time {
	return time.Now()
}
