package validators

import (
	"github.com/go-playground/validator/v10"
	"sync"
)

type fieldValidator struct {
	wrappedValidator *validator.Validate
}

var (
	once     sync.Once
	instance *fieldValidator
)

func GetInstance() *fieldValidator {
	once.Do(func() {
		instance = &fieldValidator{wrappedValidator: validator.New()}
	})
	return instance
}

func (f *fieldValidator) Struct(s interface{}) error {
	return f.wrappedValidator.Struct(s)
}
