package app

import "net/http"

type AppError struct {
	Code    int
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

func BadRequest(msg string) error {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: msg,
	}
}
