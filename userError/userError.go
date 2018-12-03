package userError

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/tenderly/tenderly-cli/rest/payloads"
)

type UserError struct {
	error       error
	userMessage string
}

func (e UserError) Error() string {
	return fmt.Sprintf("error while calling api: %s, message: %s", e.error, e.userMessage)
}

func NewUserError(error error, userMessage string) UserError {
	return UserError{error: error, userMessage: userMessage}
}

func LogError(err error) {
	if err == nil {
		return
	}
	if err, ok := err.(*UserError); ok {
		logrus.Debug(err.error)
		logrus.Info(err.userMessage)
		return
	}
	if err, ok := err.(*payloads.ApiError); ok {
		logrus.Debug(err)
		logrus.Info(err.Message)
		return
	}
	logrus.Debug(err)
}

func LogErrorf(format string, err error) {
	if err == nil {
		return
	}
	if err, ok := err.(*UserError); ok {
		logrus.Debug(fmt.Errorf(format, err.error))
		logrus.Info(err.userMessage)
		return
	}
	if err, ok := err.(*payloads.ApiError); ok {
		logrus.Debug(fmt.Errorf(format, err.Slug))
		logrus.Info(err.Message)
		return
	}
	logrus.Debug(fmt.Errorf(format, err))
}
