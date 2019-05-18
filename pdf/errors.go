package pdf

import (
	"fmt"
	"runtime/debug"
)

type ErrTrace struct {
	message string
	trace []byte
}

func NewError(err error) *ErrTrace {
	return &ErrTrace{err.Error(), debug.Stack()}
}

func NewErrorf(format string, a ...interface{}) *ErrTrace {
	return &ErrTrace{fmt.Sprintf(format, a), debug.Stack()}
}

func (err *ErrTrace) Error() string {
	return err.message
}

func (err *ErrTrace) Trace() string {
	return string(err.trace)
}

type ErrUnsupported struct {
	message string
}

func NewErrUnsupported(message string, a ...interface{}) *ErrUnsupported {
	return &ErrUnsupported{fmt.Sprintf(message, a)}
}

func (err *ErrUnsupported) Error() string {
	return err.message
}

type ErrEncrypted struct {
	message string
}

func NewErrEncrypted(message string, a ...interface{}) *ErrEncrypted {
	return &ErrEncrypted{fmt.Sprintf(message, a)}
}

func (err *ErrEncrypted) Error() string {
	return err.message
}
