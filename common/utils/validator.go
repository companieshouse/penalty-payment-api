package utils

import (
	"gopkg.in/go-playground/validator.v9"
)

func GetValidator(request interface{}) error {
	v := validator.New()
	return v.Struct(request)
}
