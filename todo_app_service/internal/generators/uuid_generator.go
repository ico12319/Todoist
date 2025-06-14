package generators

import "github.com/google/uuid"

type uuidGenerator struct{}

func NewUuidGenerator() *uuidGenerator {
	return &uuidGenerator{}
}

func (u *uuidGenerator) Generate() string {
	return uuid.New().String()
}
