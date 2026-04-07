package boltrouter

import "strings"

func NewError(message string, details ...string) ErrorType {
	return ErrorType{
		message: message,
		detail:  strings.Join(details, " "),
	}
}

type ErrorType struct {
	message string
	detail  string
}
type Error interface {
	Error() string
	Detail() string
}

func (e ErrorType) Error() string {
	return e.message
}

func (e ErrorType) Detail() string {
	return e.detail
}
