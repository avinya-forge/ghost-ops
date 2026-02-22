package protocol

import (
	"fmt"
	"runtime/debug"
)

// AppError defines the standard application error interface.
type AppError interface {
	error
	Code() string
	Message() string
	Unwrap() error
	StackTrace() string
}

// BaseError is a concrete implementation of AppError.
type BaseError struct {
	CodeStr string
	Msg     string
	Err     error
	Stack   []byte
}

func (e *BaseError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.CodeStr, e.Msg, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.CodeStr, e.Msg)
}

func (e *BaseError) Code() string {
	return e.CodeStr
}

func (e *BaseError) Message() string {
	return e.Msg
}

func (e *BaseError) Unwrap() error {
	return e.Err
}

func (e *BaseError) StackTrace() string {
	return string(e.Stack)
}

// NewAppError creates a new AppError.
func NewAppError(code, msg string, err error) AppError {
	var stack []byte
	if code == "INTERNAL" {
		stack = debug.Stack()
	}
	return &BaseError{
		CodeStr: code,
		Msg:     msg,
		Err:     err,
		Stack:   stack,
	}
}

// Sentinel Errors
var (
	ErrNotFound     = NewAppError("NOT_FOUND", "resource not found", nil)
	ErrInvalidInput = NewAppError("INVALID_INPUT", "invalid input", nil)
	ErrTimeout      = NewAppError("TIMEOUT", "operation timed out", nil)
	ErrInternal     = NewAppError("INTERNAL", "internal server error", nil)
	ErrConflict     = NewAppError("CONFLICT", "resource conflict", nil)
	ErrUnauthorized = NewAppError("UNAUTHORIZED", "unauthorized access", nil)
)
