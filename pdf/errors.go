package pdf

import (
	"fmt"
	"runtime/debug"
)

// print stack traces when true
var Verbose *bool

var ErrorBadXref *Error = NewError("Bad Xref")
var ErrorEncrypted *Error = NewError("Encrypted")
var ErrorUnsupported *Error = NewError("Unsupported")

// error that includes stack trace
type Error struct {
	message string
	trace []byte
}

// NewError creates a new Error from a formatted message
func NewError(format string, a ...interface{}) *Error {
	message := format
	if len(a) > 0 {
		message = fmt.Sprintf(format, a...)
	}
	return &Error{message, debug.Stack()}
}

// WrapError prepends a formatted message to the start of an existing error
func WrapError(err error, format string, a ...interface{}) *Error {
	// use or create new error
	e, ok := err.(*Error)
	if !ok {
		e = NewError(err.Error())
	}

	// wrap the message
	wrap := format
	if len(a) > 0 {
		wrap = fmt.Sprintf(format, a...)
	}
	e.message = fmt.Sprintf("%s: %s", wrap, e.message)

	// return new error
	return e
}

// implement error interface
func (err *Error) Error() string {
	if *Verbose {
		return fmt.Sprintf("%s\n%s", err.message, string(err.trace))
	}
	return err.message
}
