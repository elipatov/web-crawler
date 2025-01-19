// Package errs provides a custom error implementation with support for
// error codes, stack traces, and error wrapping. It allows developers to
// create structured, meaningful errors with additional context, including
// invariant error codes.
package errs

import (
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

// ErrorCode is a type of error.
type ErrorCode string

const (
	// Unauthorized error code.
	Unauthorized = ErrorCode("UNAUTHORIZED")
	// Forbidden error code.
	Forbidden = ErrorCode("FORBIDDEN")
	// NotFound error code.
	NotFound = ErrorCode("NOT_FOUND")
	// InvalidValue error code.
	InvalidValue = ErrorCode("INVALID_VALUE")
	// UnexpectedError error code.
	UnexpectedError = ErrorCode("UNEXPECTED_ERROR")
)

// Predefined errors with corresponding error codes and default messages.
var (
	// ErrUnauthorized error with Unauthorized code.
	ErrUnauthorized = &Error{errorCode: Unauthorized, message: "unauthorized"}
	// ErrForbidden error with Forbidden code.
	ErrForbidden = &Error{errorCode: Forbidden, message: "forbidden"}
	// ErrNotFound error with NotFound code.
	ErrNotFound = &Error{errorCode: NotFound, message: "not found"}
	// ErrInvalidValue error with InvalidValue code.
	ErrInvalidValue = &Error{errorCode: InvalidValue, message: "invalid value"}
	// ErrUnexpected error with UnexpectedError code.
	ErrUnexpected = &Error{errorCode: UnexpectedError, message: "unexpected error"}
)

// Error represents a custom error type that implements the error interface.
// It includes a human-readable message, an error code, a wrapped error, and
// a stack trace for debugging purposes.
type Error struct {
	message    string    // Human-readable message for clients
	errorCode  ErrorCode // Business specific error code for clients
	err        error     // The original error
	stackTrace string
}

// New returns an error with provided error code and message.
func New(code ErrorCode, message string) *Error {
	return &Error{
		message:    message,
		stackTrace: stackTrace(1),
	}
}

// Errorf returns an error with provided error code and
// with a message formatted according to a format specifier.
func Errorf(code ErrorCode, message string, args ...any) *Error {
	return &Error{
		message:    fmt.Sprintf(message, args...),
		stackTrace: stackTrace(1),
	}
}

// WrapError returns an error wrapping err.
func WrapError(err error, message ...string) error {
	if len(message) > 0 {
		return WrapErrorf(err, message[0]) //nolint:govet
	}

	return wrapErrorf(err, 1, "")
}

// WrapErrorf returns an error wrapping err with message formatted according to a format specifier.
func WrapErrorf(err error, message string, args ...any) error {
	// Do not inline wrapErrorf. WrapError and WrapErrorf need to have same stack trace.
	return wrapErrorf(err, 1, message, args...)
}

// WithErrorCode returns error copy with overridden error code.
func (err *Error) WithErrorCode(code ErrorCode) *Error {
	res := *err
	res.errorCode = code

	return &res
}

// WithMessage returns error copy with overridden message and current stack trace.
// It supposed to be used with sentinel errors: misterr.ErrForbidden.WithMessage("invalid role").
func (err *Error) WithMessage(message string) *Error {
	res := *err
	res.message = message
	res.stackTrace = stackTrace(1)

	return &res
}

// WithMessagef returns error copy with overridden message and current stack trace.
// It supposed to be used with sentinel errors: misterr.ErrForbidden.WithMessage("invalid role").
func (err *Error) WithMessagef(message string, args ...any) *Error {
	res := *err
	res.message = fmt.Sprintf(message, args)
	res.stackTrace = stackTrace(1)

	return &res
}

// WithError returns error copy with overridden inner error.
// Original stack trace is used for misterr errors.
func (err *Error) WithError(newErr error) *Error {
	res := *err
	res.err = newErr

	if e := new(Error); errors.As(newErr, &e) && e.stackTrace != "" {
		res.stackTrace = e.stackTrace
	}

	if res.stackTrace == "" { // Sentinel error.
		res.stackTrace = stackTrace(1)
	}

	return &res
}

// Unwrap method returns wrapped error.
func (err *Error) Unwrap() error {
	return err.err
}

// Error method returns error string representation.
func (err *Error) Error() string {
	if err.err != nil {
		if err.message == "" {
			return err.err.Error()
		}

		return fmt.Sprintf("%s: %s", err.message, err.err.Error())
	}

	return err.message
}

// ErrorCode of the error.
func (err *Error) ErrorCode() ErrorCode {
	return err.errorCode
}

// Code returns ErrorCode as string.
func (err *Error) Code() string {
	return string(err.errorCode)
}

// StackTrace of the error.
func (err *Error) StackTrace() string {
	if err.stackTrace != "" {
		return err.stackTrace
	}

	// A naked sentinel error contains an empty stack trace.
	// The best we can do is return the current one.
	return stackTrace(1)
}

// Is reports whether error matches target.
func (err *Error) Is(target error) bool {
	var t *Error

	ok := errors.As(target, &t)
	if !ok {
		return false
	}

	return err.ErrorCode() == t.ErrorCode()
}

func wrapErrorf(err error, skipFrames int, message string, args ...any) error {
	if err == nil {
		return nil
	}

	if e := new(Error); errors.As(err, &e) {
		// Use original fields for misterr error, but do not duplicate message.
		msg := fmt.Sprintf(message, args...)
		stack := e.stackTrace

		if stack == "" { // Sentinel error.
			stack = stackTrace(skipFrames + 1)
		}

		return &Error{
			message:    msg,
			err:        err,
			stackTrace: stack,
			errorCode:  e.errorCode,
		}
	}

	return &Error{
		message:    fmt.Sprintf(message, args...),
		err:        err,
		stackTrace: stackTrace(skipFrames + 1),
		errorCode:  UnexpectedError,
	}
}

func stackTrace(skipFrames int) string {
	const maxStackDepth = 20

	callers := make([]uintptr, maxStackDepth)
	length := runtime.Callers(skipFrames+2, callers) //nolint:mnd // skip frames from this file + runtime.Callers one.
	frames := runtime.CallersFrames(callers[:length])
	resultBuilder := strings.Builder{}
	frame, more := frames.Next()

	// Sentinel errors called from init() function.
	if strings.HasSuffix(frame.Function, ".init") {
		return ""
	}

	for ; more; frame, more = frames.Next() {
		resultBuilder.WriteString(frame.Function)
		resultBuilder.WriteString("\n\t")
		resultBuilder.WriteString(frame.File)
		resultBuilder.WriteString(":")
		resultBuilder.WriteString(strconv.Itoa(frame.Line))
		resultBuilder.WriteString("\n")
	}

	return resultBuilder.String()
}
