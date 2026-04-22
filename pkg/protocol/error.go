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

// newSentinel creates a code/message marker without capturing a stack trace.
// Sentinels are used for code() comparison only; callers should always wrap
// them via NewAppError at the return site so the stack trace reflects the
// actual error location, not package init.
func newSentinel(code, msg string) AppError {
	return &BaseError{CodeStr: code, Msg: msg}
}

// Sentinel Errors — compare via .Code(), never return bare sentinels.
var (
	ErrNotFound     = newSentinel("NOT_FOUND", "resource not found")
	ErrInvalidInput = newSentinel("INVALID_INPUT", "invalid input")
	ErrTimeout      = newSentinel("TIMEOUT", "operation timed out")
	ErrInternal     = newSentinel("INTERNAL", "internal server error")
	ErrConflict     = newSentinel("CONFLICT", "resource conflict")
	ErrUnauthorized = newSentinel("UNAUTHORIZED", "unauthorized access")
)
