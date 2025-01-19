package errs

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ErrSentinel = New("Sentinel")

func TestErrorStackTrace(t *testing.T) {
	errs := []*Error{getAnError0(), getAnErrorNew0(), getSentinelError0(true)}
	expectedL01 := [][2]string{
		{"gitlab.com/wnf3/common/go/libs/misterr.getAnError1", "gitlab.com/wnf3/common/go/libs/misterr.getAnError0"},
		{"gitlab.com/wnf3/common/go/libs/misterr.getAnErrorNew1", "gitlab.com/wnf3/common/go/libs/misterr.getAnErrorNew0"},
		{"gitlab.com/wnf3/common/go/libs/misterr.getSentinelError1", "gitlab.com/wnf3/common/go/libs/misterr.getSentinelError0"},
	}

	require.Error(t, errs[0])
	require.Error(t, errs[1])
	require.Error(t, errs[2])

	for i, err := range errs {
		lines := strings.Split(err.StackTrace(), "\n")
		require.Greater(t, len(lines), 6)
		assert.Equal(t, expectedL01[i][0], lines[0])
		assert.Contains(t, lines[1], "/error_test.go:")
		assert.Equal(t, expectedL01[i][1], lines[2])
		assert.Contains(t, lines[3], "/error_test.go:")
		assert.Equal(t, "gitlab.com/wnf3/common/go/libs/misterr.TestErrorStackTrace", lines[4])
		assert.Contains(t, lines[5], "/error_test.go:")
	}
}

func TestNakeSentinelErrorStackTrace(t *testing.T) {
	err := getSentinelError0(false)

	require.Error(t, err)

	lines := strings.Split(err.StackTrace(), "\n")
	require.Greater(t, len(lines), 2)
	assert.Equal(t, "gitlab.com/wnf3/common/go/libs/misterr.TestNakeSentinelErrorStackTrace", lines[0])
	assert.Contains(t, lines[1], "/error_test.go:")
}

func TestWrapStackTrace(t *testing.T) {
	originals := []*Error{getAnError0(), getAnErrorNew0(), getSentinelError0(true)}
	expectedL01 := [][2]string{
		{"gitlab.com/wnf3/common/go/libs/misterr.getAnError1", "gitlab.com/wnf3/common/go/libs/misterr.getAnError0"},
		{"gitlab.com/wnf3/common/go/libs/misterr.getAnErrorNew1", "gitlab.com/wnf3/common/go/libs/misterr.getAnErrorNew0"},
		{"gitlab.com/wnf3/common/go/libs/misterr.getSentinelError1", "gitlab.com/wnf3/common/go/libs/misterr.getSentinelError0"},
	}

	require.Error(t, originals[0])
	require.Error(t, originals[1])
	require.Error(t, originals[2])

	for i, e := range originals {
		err := assertMisterr(t, WrapError(e))
		lines := strings.Split(err.StackTrace(), "\n")

		require.Error(t, err)
		require.Greater(t, len(lines), 4)
		assert.Equal(t, expectedL01[i][0], lines[0])
		assert.Contains(t, lines[1], "/error_test.go:")
		assert.Equal(t, expectedL01[i][1], lines[2])
		assert.Contains(t, lines[3], "/error_test.go:")
		assert.Equal(t, "gitlab.com/wnf3/common/go/libs/misterr.TestWrapStackTrace", lines[4])
		assert.Contains(t, lines[5], "/error_test.go:")
	}
}

func TestWrapMisterrError(t *testing.T) {
	original := ErrSentinel
	err := assertMisterr(t, WrapError(original))
	lines := strings.Split(err.StackTrace(), "\n")

	assert.Equal(t, "gitlab.com/wnf3/common/go/libs/misterr.TestWrapMisterrError", lines[0])
	assert.Equal(t, original.Error(), err.Error())
	assert.Equal(t, original.errorCode, err.errorCode)

	err = assertMisterr(t, WrapErrorf(original, ""))
	assert.Equal(t, original.Error(), err.Error())
	assert.Equal(t, original.errorCode, err.errorCode)

	const newMessage = "Invalid2"
	err = assertMisterr(t, WrapError(original, newMessage))

	assert.Equal(t, newMessage, err.message)
	assert.Equal(t, fmt.Sprintf("%s: %s", newMessage, original.message), err.Error())
	assert.Equal(t, original.errorCode, err.errorCode)

	err = assertMisterr(t, WrapErrorf(original, newMessage))
	assert.Equal(t, newMessage, err.message)
	assert.Equal(t, fmt.Sprintf("%s: %s", newMessage, original.message), err.Error())
	assert.Equal(t, original.errorCode, err.errorCode)
}

func TestErrorMessage(t *testing.T) {
	err := Errorf("Unexpected string value %s and number %d", "bad", 42)
	require.Error(t, err)
	assert.Equal(t, "Unexpected string value bad and number 42", err.Error())
}

func TestWrapUnwrap(t *testing.T) {
	msg := "Some error"
	original := errors.New(msg)
	err := assertMisterr(t, WrapError(original))
	require.Error(t, err)

	assert.Equal(t, msg, err.Error())
	assert.Equal(t, original, err.Unwrap())
	assert.Equal(t, UnexpectedError, err.ErrorCode())
}

func TestWrapNil(t *testing.T) {
	err := WrapError(nil)
	if err != nil {
		// Do not replace with assert.Nil() (it leads false negative).
		assert.Fail(t, "WrapError(nil) must be nil")
	}

	err = WrapErrorf(nil, "Some message")
	if err != nil {
		// Do not replace with assert.Nil() (it leads false negative).
		assert.Fail(t, "WrapErrorf(nil, ...) must be nil")
	}
}

func TestWrapfUnwrapWithMessage(t *testing.T) {
	msg := "Original error"
	original := errors.New(msg)
	err := assertMisterr(t, WrapErrorf(original, "Failed %s", "service"))
	require.Error(t, err)

	assert.Equal(t, "Failed service: Original error", err.Error())
	assert.Equal(t, original, err.Unwrap())
	assert.Equal(t, UnexpectedError, err.ErrorCode())
}

func TestWrapUnwrapWithMessage(t *testing.T) {
	msg := "Original error"
	original := errors.New(msg)
	err := assertMisterr(t, WrapError(original, "Failed wrap"))
	require.Error(t, err)

	assert.Equal(t, "Failed wrap: Original error", err.Error())
	assert.Equal(t, original, err.Unwrap())
	assert.Equal(t, UnexpectedError, err.ErrorCode())
}

func TestWithErrorCode(t *testing.T) {
	err := New("Bad value").WithErrorCode(InvalidValue)
	require.Error(t, err)
	assert.Equal(t, InvalidValue, err.ErrorCode())
	assert.Equal(t, string(InvalidValue), err.Code())

	err = ErrForbidden
	require.Equal(t, Forbidden, err.ErrorCode())
	assert.Equal(t, string(Forbidden), err.Code())

	err2 := err.WithErrorCode(UnexpectedError)

	assert.NotEqual(t, err, err2)
	assert.Equal(t, UnexpectedError, err2.ErrorCode())
	assert.Equal(t, string(UnexpectedError), err2.Code())
	assert.Equal(t, Forbidden, ErrForbidden.ErrorCode())
}

func TestWithMessage(t *testing.T) {
	originalMessage := ErrBadRequest.message
	newMessage := "Some other message"
	err := ErrBadRequest

	err2 := err.WithMessage(newMessage)

	assert.NotEqual(t, err, err2)
	assert.Equal(t, newMessage, err2.message)
	assert.Equal(t, originalMessage, ErrBadRequest.message)
}

func TestWithError(t *testing.T) {
	err1 := getAnError0()
	err2 := ErrBadRequest
	err := err2.WithError(err1)
	lines := strings.Split(err.StackTrace(), "\n")

	assert.NotEqual(t, err1, err)
	assert.NotEqual(t, err2, err)
	assert.Equal(t, err1, err.err)
	assert.Equal(t, err2.message, err.message)
	assert.Equal(t, err2.errorCode, err.ErrorCode())
	require.Greater(t, len(lines), 1)
	assert.Contains(t, lines[0], "misterr.getAnError1")

	err1 = ErrBadRequest
	err2 = getAnError0()
	err = err2.WithError(err1)
	lines = strings.Split(err.StackTrace(), "\n")

	assert.NotEqual(t, err1, err)
	assert.NotEqual(t, err2, err)
	assert.Equal(t, err1, err.err)
	assert.Equal(t, err2.message, err.message)
	assert.Equal(t, err2.errorCode, err.ErrorCode())
	require.Greater(t, len(lines), 1)
	assert.Contains(t, lines[0], "misterr.getAnError1")

	err1 = ErrBadRequest
	err2 = ErrForbidden
	err = err2.WithError(err1)
	lines = strings.Split(err.StackTrace(), "\n")

	assert.NotEqual(t, err1, err)
	assert.NotEqual(t, err2, err)
	assert.Equal(t, err1, err.err)
	assert.Equal(t, err2.message, err.message)
	assert.Equal(t, err2.errorCode, err.ErrorCode())
	require.Greater(t, len(lines), 1)
	assert.Contains(t, lines[0], "misterr.TestWithError")
}

func TestIs(t *testing.T) {
	errForbidden := &Error{
		errorCode: Forbidden,
	}

	ok := errors.Is(errForbidden, ErrForbidden)
	assert.True(t, ok)

	ok = errors.Is(errForbidden, ErrInvalidValue)
	assert.False(t, ok)

	errForbiddenWrapped := assertMisterr(t, WrapErrorf(errors.New("forbidden"), "forbidden error message")).WithErrorCode(Forbidden)

	ok = errors.Is(errForbiddenWrapped, ErrForbidden)
	assert.True(t, ok)

	ok = errors.Is(errForbiddenWrapped, ErrInvalidValue)
	assert.False(t, ok)

	ok = errors.Is(errForbidden, errors.New(ErrForbidden.Error()))
	assert.False(t, ok)

	ok = errors.Is(WrapError(ErrBadRequest), ErrBadRequest)
	assert.True(t, ok)
}

func getAnError0() *Error {
	return getAnError1()
}

func getAnError1() *Error {
	return Errorf("Unexpected")
}

func getAnErrorNew0() *Error {
	return getAnErrorNew1()
}

func getAnErrorNew1() *Error {
	return New("Unexpected")
}

func getSentinelError0(wrap bool) *Error {
	return getSentinelError1(wrap)
}

func getSentinelError1(wrap bool) *Error {
	if wrap {
		return WrapError(ErrSentinel).(*Error) //nolint // test relies on implementation
	}

	return ErrSentinel
}

func assertMisterr(t *testing.T, err error) *Error {
	t.Helper()

	res, ok := err.(*Error) //nolint:errorlint // test relies on implementation
	if !ok {
		require.Fail(t, "expected *Error")
	}

	return res
}
