// Package domain holds fwb's entities, value objects and the ports
// (interfaces) that infrastructure implements. It has no I/O dependencies.
package domain

import (
	"fmt"

	"github.com/cylixlee/fwb/internal/consts"
)

// Error is a typed, translatable error.
type Error struct {
	Code  consts.ErrorCode
	Args  []any
	Cause error
}

// NewError creates an error with no wrapped cause.
func NewError(code consts.ErrorCode, args ...any) *Error {
	return &Error{Code: code, Args: args}
}

// WrapError creates an error that wraps cause.
func WrapError(code consts.ErrorCode, cause error, args ...any) *Error {
	return &Error{Code: code, Args: args, Cause: cause}
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf(consts.ErrorTextFormat, e.Code, e.Cause)
	}
	return string(e.Code)
}

func (e *Error) Unwrap() error { return e.Cause }
