package utils

import (
	"gopkg.in/go-playground/validator.v9"
)

type Validator struct {
}

func GetValidator() *Validator {
	return &Validator{}
}

func (v *Validator) Validate(request interface{}) error {
	return validator.New().Struct(request)
}
