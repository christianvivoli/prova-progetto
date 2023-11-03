package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-stack/stack"
)

// Application error codes.
//
// NOTE: These are meant to be generic and they map well to HTTP error codes.
// Different applications can have very different error code requirements so
// these should be expanded as needed (or introduce subcodes).
const (
	// standard errors code.
	ECANCELED       = "canceled"        // canceled request
	ECONFLICT       = "conflict"        // conflict with current state
	EINTERNAL       = "internal"        // internal error
	EINVALID        = "invalid"         // invalid input
	ENOTFOUND       = "not_found"       // resource not found
	ENOTIMPLEMENTED = "not_implemented" // feature not implemented
	EUNAUTHORIZED   = "unauthorized"    // access denied
	EUNKNOWN        = "unknown"         // unknown error
	EFORBIDDEN      = "forbidden"       // access forbidden
	EEXISTS         = "exists"          // resource already exists
	ENOTINJECTED    = "not_injected"    // resource not injected
	EUNAVAILABLE    = "unavailable"     // resource is not available

	// custom errors code.
	ENOTAUTHENTICATED = "not_authenticated" // user not authenticated
	ESHOULDLOGOUT     = "should_logout"     // user should logout
	EMAILALREADYINUSE = "email_already_in_use"

	// sub codes.
	EINTERNAL_INVALID = "internal_invalid" // invalid data for internal state
)

// Error represents an application-specific error. Application errors can be
// unwrapped by the caller to extract out the code & message.
//
// Any non-application error (such as a disk error) should be reported as an
// EINTERNAL error and the human user should only see "Internal error" as the
// message. These low-level internal error details should only be logged and
// reported to the operator of the application (not the end user).
type Error struct {
	// Machine-readable error code.
	Code string

	// Human-readable error message.
	Message string

	// OriginFile file and line where error was raised.
	OriginFile string

	// OriginFn fn where error was raised.
	OriginFn string
}

// Error implements the error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("code=%s message=%s", e.Code, e.Message)
}

// ErrorCode unwraps an application error and returns its code.
// Non-application errors always return EINTERNAL.
func ErrorCode(err error) string {
	var e *Error
	if err == nil {
		return ""
	} else if errors.As(err, &e) {
		return e.Code
	}
	return EINTERNAL
}

// ErrorLogEntries returns the params for a log entry for an application error.
func ErrorLogEntries(err error) []any {
	var e *Error
	if err == nil {
		return []any{}
	} else if errors.As(err, &e) {
		return []any{"code", e.Code, "origin_file", e.OriginFile, "origin_fn", e.OriginFn}
	}
	return []any{"code", EINTERNAL}
}

// ErrorMessage unwraps an application error and returns its message.
// Non-application errors always return "Internal error".
func ErrorMessage(err error) string {
	var e *Error
	if err == nil {
		return ""
	} else if errors.As(err, &e) {
		return e.Message
	}
	return err.Error()
}

// Errorf is a helper function to return an Error with a given code and formatted message.
func Errorf(code string, format string, args ...interface{}) *Error {

	message := fmt.Sprintf(format, args...)

	if code != ECANCELED && strings.Contains(message, context.Canceled.Error()) {
		code = ECANCELED
	}

	caller := stack.Caller(1)

	return &Error{
		Code:       code,
		Message:    message,
		OriginFile: fmt.Sprint(caller),
		OriginFn:   fmt.Sprintf("%+n", caller),
	}
}
