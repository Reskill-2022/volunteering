package errors

import "fmt"

type Error struct {
	Msg   string
	Code  int
	Cause error
}

func (e Error) Message() string {
	return e.Msg
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Msg, e.cause())
}

func (e Error) cause() string {
	if e.Cause == nil {
		return ""
	}
	return e.Cause.Error()
}

func New(msg string, code int) Error {
	return Error{
		Msg:  msg,
		Code: code,
	}
}

func From(err error, msg string, code int) Error {
	return Error{
		Msg:   msg,
		Code:  code,
		Cause: err,
	}
}

func CodeFrom(err error) int {
	code := 1
	if v, ok := err.(Error); ok {
		code = v.Code
	}
	return code
}
