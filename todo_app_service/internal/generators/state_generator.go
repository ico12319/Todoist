package generators

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

type stateGenerator struct{}

func NewStateGenerator() *stateGenerator {
	return &stateGenerator{}
}

func (s *stateGenerator) GenerateState() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("error when trying to generate state")
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
