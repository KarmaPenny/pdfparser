package pdf

import (
	"errors"
	"fmt"
)

var ErrEncrypted = errors.New("Pdf is encrypted")
var ErrStartXrefNotFound = errors.New("Missing startxref")
var ErrStreamDictionaryMissing = errors.New("Stream has no stream dictionary")

type ErrUnsupportedFilter struct {
	filter string
}

func NewErrUnsupportedFilter(filter string) error {
	return &ErrUnsupportedFilter{filter}
}

func (err *ErrUnsupportedFilter) Error() string {
	return fmt.Sprintf("Unsupported filter: %s", err.filter)
}
