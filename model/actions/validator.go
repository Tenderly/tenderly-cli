package actions

import (
	"fmt"
)

type ValidatorContext string

func (c ValidatorContext) With(element string) ValidatorContext {
	return ValidatorContext(fmt.Sprintf("%s.%s", c, element))
}

type Validator interface {
	Validate(ValidatorContext) (response ValidateResponse)
}

type ValidateResponse struct {
	Infos  []string
	Errors []string
}

func (v *ValidateResponse) Info(c ValidatorContext, msg string, args ...interface{}) ValidateResponse {
	v.Infos = append(v.Infos, format(c, msg, args...))
	return *v
}

func (v *ValidateResponse) Error(c ValidatorContext, msg string, args ...interface{}) ValidateResponse {
	v.Errors = append(v.Errors, format(c, msg, args...))
	return *v
}

func format(c ValidatorContext, msg string, args ...interface{}) string {
	return fmt.Sprintf("%s: %s", c, fmt.Sprintf(msg, args...))
}

func (v *ValidateResponse) Merge(response ValidateResponse) ValidateResponse {
	for _, msg := range response.Errors {
		v.Errors = append(v.Errors, msg)
	}
	for _, msg := range response.Infos {
		v.Infos = append(v.Infos, msg)
	}
	return *v
}
